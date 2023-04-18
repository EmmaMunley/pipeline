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
	"strings"
	"testing"

	"github.com/tektoncd/pipeline/test/parse"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	knativetest "knative.dev/pkg/test"
	"knative.dev/pkg/test/helpers"
)

var (
	alphaFeatureFlags = map[string]string{
		// Make sure this is running under alpha integration tests
		"enable-api-fields": "alpha",
	}
)

// TestPipelineRunMatrixed is an integration test that will
// verify that pipelinerun works with a matrixed pipelinerun
func TestPipelineRunMatrixed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	fmt.Println("TEST")
	c, namespace := setup(ctx, t, requireAnyGate(alphaFeatureFlags))
	knativetest.CleanupOnInterrupt(func() { tearDown(context.Background(), t, c, namespace) }, t.Logf)
	defer tearDown(context.Background(), t, c, namespace)
	t.Logf("Creating Task in namespace %s", namespace)
	task := parse.MustParseV1beta1Task(t, fmt.Sprintf(`
metadata:
  name: %s
  namespace: %s
spec:
  params:
    - name: platform
      default: ""
    - name: browser
      default: ""
  steps:
    - name: echo
      image: alpine
      script: |
        echo "$(params.platform) and $(params.browser)"
`, helpers.ObjectNameForTest(t), namespace))
	if _, err := c.V1beta1TaskClient.Create(ctx, task, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Failed to create Task `%s`: %s", task.Name, err)
	}

	pipeline := parse.MustParseV1beta1Pipeline(t, fmt.Sprintf(`
metadata:
  name: %s
  namespace: %s
spec:
  tasks:
    - name: platforms-and-browsers
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
      taskRef:
        name:  %s
`, helpers.ObjectNameForTest(t), namespace, task.Name))
	pipelineRun := parse.MustParseV1beta1PipelineRun(t, fmt.Sprintf(`
metadata:
  name: %s
  namespace: %s
spec:
  serviceAccountName: "default"
  pipelineRef:
    name: %s
`, helpers.ObjectNameForTest(t), namespace, pipeline.Name))
	if _, err := c.V1beta1PipelineClient.Create(ctx, pipeline, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Failed to create Pipeline `%s`: %s", pipeline.Name, err)
	}
	if _, err := c.V1beta1PipelineRunClient.Create(ctx, pipelineRun, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Failed to create PipelineRun `%s`: %s", pipelineRun.Name, err)
	}
	name := "matrix"
	prName := pipelineRun.Name
	expectedTaskRunNames := []string{"platforms-and-browsers-0", "platforms-and-browsers-1", "platforms-and-browsers-2", "platforms-and-browsers-3", "platforms-and-browsers-4", "platforms-and-browsers-5", "platforms-and-browsers-6", "platforms-and-browsers-7", "platforms-and-browsers-8"}
	expectedNumberOfEvents := 1
	t.Logf("Waiting for PipelineRun %s in namespace %s to complete", prName, namespace)
	if err := WaitForPipelineRunState(ctx, c, prName, timeout, PipelineRunSucceed(prName), "PipelineRunSuccess", v1beta1Version); err != nil {
		t.Fatalf("Error waiting for PipelineRun %s to finish: %s", prName, err)
	}
	t.Logf("Making sure the expected TaskRuns %s were created", expectedTaskRunNames)
	actualTaskrunList, err := c.V1beta1TaskRunClient.List(ctx, metav1.ListOptions{LabelSelector: fmt.Sprintf("tekton.dev/pipelineRun=%s", prName)})
	if err != nil {
		t.Fatalf("Error listing TaskRuns for PipelineRun %s: %s", prName, err)
	}

	for _, taskRunName := range expectedTaskRunNames {
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
			t.Fatalf("Expected %d number of successful events from pipelinerun and taskrun but got %d; list of receieved events : %#v", expectedNumberOfEvents, len(events), collectedEvents)
		}

		t.Logf("Successfully finished test %q", name)
	}
}
