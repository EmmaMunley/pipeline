//go:build e2e
// +build e2e

/*
Copyright 2021 The Tekton Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package test

import (
	"context"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tektoncd/pipeline/test/diff"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tektoncd/pipeline/test/parse"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	knativetest "knative.dev/pkg/test"
	"knative.dev/pkg/test/helpers"
	"strings"
	"testing"
)

var (
	alphaFeatureFlags = map[string]string{
		// Make sure this is running under alpha integration tests
		"enable-api-fields": "alpha",
	}
	ignoreResourceVersion = cmpopts.IgnoreFields(metav1.ObjectMeta{}, "ResourceVersion")
	ignoreTypeMeta        = cmpopts.IgnoreFields(metav1.TypeMeta{}, "Kind", "APIVersion")
	ignoreTimeout          = cmpopts.IgnoreFields(v1beta1.TaskRunSpec{}, "Timeout")
	trueb = true
)

// TestPipelineRunMatrixed is an integration test that will
// verify that pipelinerun works with a matrixed pipelinerun
func TestPipelineRunMatrixed(t *testing.T) {

	t.Parallel()
	type tests struct {
		name                     string
		task                     *v1beta1.Task
		getTask                  func(namespace string) *v1beta1.Task
		getPipeline              func(namespace string, taskName string) *v1beta1.Pipeline
		getPipelineRun           func(namespace string, pipelineName string) *v1beta1.PipelineRun
		expectedTaskRunNames     []string
		expectedNumberOfTaskRuns int
	}

	tds := []tests{{
		name: "matrix",
		getTask: func(namespace string) *v1beta1.Task {
			return parse.MustParseV1beta1Task(t, fmt.Sprintf(`
metadata:
  name: platforms-and-browsers
  namespace: %s
spec:
  params:
    - name: platform
      default: ""
    - name: browser
      default: ""
    - name: version
      default: ""
  steps:
    - name: echo
      image: alpine
      script: |
        echo "$(params.platform) and $(params.browser)"
`, namespace))
		},
		getPipeline: func(namespace string, taskName string) *v1beta1.Pipeline {
			return parse.MustParseV1beta1Pipeline(t, fmt.Sprintf(`
metadata:
  name: %s
  namespace: %s
spec:
  tasks:
    - name: platforms-and-browsers
      taskRef:
        name: mytask
      matrix:
        params:
          - name: platform
            value:
              - linux
              - mac
              - windows
          - name: browser
            value:
              - chrome
              - safari
              - firefox
      params:
        - name: version
          value: v0.33.0
      taskRef:
        name:  %s
`, helpers.ObjectNameForTest(t), namespace, taskName))
		},
		getPipelineRun: func(namespace string, pipelineName string) *v1beta1.PipelineRun {
			return parse.MustParseV1beta1PipelineRun(t, fmt.Sprintf(`
metadata:
  name: %s
  namespace: %s
spec:
  pipelineRef:
    name: %s
`, helpers.ObjectNameForTest(t), namespace, pipelineName))
		},
		expectedTaskRunNames:     []string{"platforms-and-browsers-0", "platforms-and-browsers-1", "platforms-and-browsers-2", "platforms-and-browsers-3", "platforms-and-browsers-4", "platforms-and-browsers-5", "platforms-and-browsers-6", "platforms-and-browsers-7", "platforms-and-browsers-8"},
		expectedNumberOfTaskRuns: 9,
	}}

	for _, td := range tds {
		td := td // capture range variable
		t.Run(td.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			c, namespace := setup(ctx, t, requireAnyGate(alphaFeatureFlags))
			fmt.Println("namespace", namespace)
			knativetest.CleanupOnInterrupt(func() { tearDown(context.Background(), t, c, namespace) }, t.Logf)
			defer tearDown(context.Background(), t, c, namespace)
			t.Logf("Creating Task in namespace %s", namespace)
			task := td.getTask(namespace)
			if _, err := c.V1beta1TaskClient.Create(ctx, task, metav1.CreateOptions{}); err != nil {
				t.Fatalf("Failed to create Task `%s`: %s", task.Name, err)
			}
			pipeline := td.getPipeline(namespace, task.Name)
			if _, err := c.V1beta1PipelineClient.Create(ctx, pipeline, metav1.CreateOptions{}); err != nil {
				t.Fatalf("Failed to create Pipeline `%s`: %s", pipeline.Name, err)
			}
			pipelineRun := td.getPipelineRun(namespace, pipeline.Name)
			if _, err := c.V1beta1PipelineRunClient.Create(ctx, pipelineRun, metav1.CreateOptions{}); err != nil {
				t.Fatalf("Failed to create PipelineRun `%s`: %s", pipelineRun.Name, err)
			}
			prName := pipelineRun.Name
			t.Logf("Waiting for PipelineRun %s in namespace %s to complete", prName, namespace)
			if err := WaitForPipelineRunState(ctx, c, prName, timeout, PipelineRunSucceed(prName), "PipelineRunSuccess", v1beta1Version); err != nil {
				t.Fatalf("Error waiting for PipelineRun %s to finish: %s", prName, err)
			}
			t.Logf("Making sure the expected TaskRuns %s were created", td.expectedTaskRunNames)
			actualTaskrunList, err := c.V1beta1TaskRunClient.List(ctx, metav1.ListOptions{LabelSelector: fmt.Sprintf("tekton.dev/pipelineRun=%s", prName)})
			if err != nil {
				t.Fatalf("Error listing TaskRuns for PipelineRun %s: %s", prName, err)
			}

			if len(actualTaskrunList.Items) != (td.expectedNumberOfTaskRuns) {
				t.Fatalf("Expected `%d` of task runs to be created, but got %d", td.expectedNumberOfTaskRuns, len(actualTaskrunList.Items))
			}

			for _, taskRunName := range td.expectedTaskRunNames {
				for _, actualTaskRunItem := range actualTaskrunList.Items {
					if strings.HasSuffix(actualTaskRunItem.Name, taskRunName) {
						taskRunName = actualTaskRunItem.Name
					}
				}

				r, err := c.V1beta1TaskRunClient.Get(ctx, taskRunName, metav1.GetOptions{})
				if err != nil {
					t.Fatalf("Couldn't get expected TaskRun %s: %s", taskRunName, err)
				}
				if !r.Status.GetCondition(apis.ConditionSucceeded).IsTrue() {
					t.Fatalf("Expected TaskRun %s to have succeeded but Status is %v", taskRunName, r.Status)
				}

				if _, err := c.V1beta1PipelineRunClient.Get(ctx, pipelineRun.Name, metav1.GetOptions{}); err != nil {
					t.Fatalf("Failed to get PipelineRun `%s`: %s", pipelineRun.Name, err)
				}

				t.Logf("Checking that labels were propagated correctly for TaskRun %s", r.Name)
				checkLabelPropagation(ctx, t, c, namespace, prName, r)
				t.Logf("Checking that annotations were propagated correctly for TaskRun %s", r.Name)
				checkAnnotationPropagation(ctx, t, c, namespace, prName, r)

		matchKinds := map[string][]string{"PipelineRun": {prName}, "TaskRun": expectedTaskRunNames}

		t.Logf("Making sure %d events were created from taskrun and pipelinerun with kinds %v", expectedNumberOfEvents, matchKinds)

		events, err := collectMatchingEvents(ctx, c.KubeClient, namespace, matchKinds, "Succeeded")
		if err != nil {
			t.Fatalf("Failed to collect matching events: %q", err)
		}
		if len(events) != expectedNumberOfEvents {
			collectedEvents := ""
			for i, event := range events {
				collectedEvents += fmt.Sprintf("%#v", event)
				if i < (len(events) - 1) {
					collectedEvents += ", "
				}
			}
		}
	}
	return om
}

func mustParseTaskRunWithObjectMeta(t *testing.T, objectMeta metav1.ObjectMeta, asYAML string) *v1beta1.TaskRun {
	t.Helper()
	tr := parse.MustParseV1beta1TaskRun(t, asYAML)
	tr.ObjectMeta = objectMeta
	return tr
}

func mustParseCustomRunWithObjectMeta(t *testing.T, objectMeta metav1.ObjectMeta, asYAML string) *v1beta1.CustomRun {
	t.Helper()
	r := parse.MustParseCustomRun(t, asYAML)
	r.ObjectMeta = objectMeta
	return r
}

func taskRunObjectMeta(trName, ns, prName, pipelineName, pipelineTaskName string, skipMemberOfLabel bool) metav1.ObjectMeta {
	om := metav1.ObjectMeta{
		Name:      trName,
		Namespace: ns,
		OwnerReferences: []metav1.OwnerReference{{
			Kind:               "PipelineRun",
			Name:               prName,
			APIVersion:         "tekton.dev/v1beta1",
			Controller:         &trueb,
			BlockOwnerDeletion: &trueb,
		}},
		Labels: map[string]string{
			pipeline.PipelineLabelKey:     pipelineName,
			pipeline.PipelineRunLabelKey:  prName,
			pipeline.PipelineTaskLabelKey: pipelineTaskName,
		},
		Annotations: map[string]string{},
	}
	if !skipMemberOfLabel {
		om.Labels[pipeline.MemberOfLabelKey] = v1beta1.PipelineTasks
	}
	return om
}

func mustParseTaskRunWithObjectMeta(t *testing.T, objectMeta metav1.ObjectMeta, asYAML string) *v1beta1.TaskRun {
	t.Helper()
	tr := parse.MustParseV1beta1TaskRun(t, asYAML)
	tr.ObjectMeta = objectMeta
	return tr
}

func mustParseCustomRunWithObjectMeta(t *testing.T, objectMeta metav1.ObjectMeta, asYAML string) *v1beta1.CustomRun {
	t.Helper()
	r := parse.MustParseCustomRun(t, asYAML)
	r.ObjectMeta = objectMeta
	return r
}
