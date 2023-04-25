//go:build e2e
// +build e2e

/*
Copyright 2019 The Tekton Authors

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

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tektoncd/pipeline/test/parse"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	knativetest "knative.dev/pkg/test"
	"knative.dev/pkg/test/helpers"
)

func Test(t *testing.T) {
	t.Parallel()
	type tests struct {
		name                   string
		testSetup              func(ctx context.Context, t *testing.T, c *clients, namespace string, index int) *v1beta1.Pipeline
		expectedTaskRuns       []string
		expectedNumberOfEvents int
		pipelineRunFunc        func(*testing.T, int, string, string) *v1beta1.PipelineRun
	}

	tds := []tests{{
		name: "fan-in and fan-out",
		testSetup: func(ctx context.Context, t *testing.T, c *clients, namespace string, _ int) *v1beta1.Pipeline {
			t.Helper()
			tasks := getFanInFanOutTasks(t, namespace)
			for _, task := range tasks {
				if _, err := c.V1beta1TaskClient.Create(ctx, task, metav1.CreateOptions{}); err != nil {
					t.Fatalf("Failed to create Task `%s`: %s", task.Name, err)
				}
			}

			p := getFanInFanOutPipeline(t, namespace, tasks)
			if _, err := c.V1beta1PipelineClient.Create(ctx, p, metav1.CreateOptions{}); err != nil {
				t.Fatalf("Failed to create Pipeline `%s`: %s", p.Name, err)
			}

			return p
		},
		pipelineRunFunc:  getFanInFanOutPipelineRun,
		expectedTaskRuns: []string{"create-file-kritis", "create-fan-out-1", "create-fan-out-2", "check-fan-in"},
		// 1 from PipelineRun and 4 from Tasks defined in pipelinerun
		expectedNumberOfEvents: 5,
	}, {
		name: "service account propagation and pipeline param",
		testSetup: func(ctx context.Context, t *testing.T, c *clients, namespace string, index int) *v1beta1.Pipeline {
			t.Helper()
			t.Skip("build-crd-testing project got removed, the secret-sauce doesn't exist anymore, skipping")
			if _, err := c.KubeClient.CoreV1().Secrets(namespace).Create(ctx, getPipelineRunSecret(index, namespace), metav1.CreateOptions{}); err != nil {
				t.Fatalf("Failed to create secret `%s`: %s", getName(secretName, index), err)
			}

			if _, err := c.KubeClient.CoreV1().ServiceAccounts(namespace).Create(ctx, getPipelineRunServiceAccount(index, namespace), metav1.CreateOptions{}); err != nil {
				t.Fatalf("Failed to create SA `%s`: %s", getName(saName, index), err)
			}

			task := parse.MustParseV1beta1Task(t, fmt.Sprintf(`
metadata:
  name: %s
  namespace: %s
spec:
  params:
  - name: the.path
    type: string
  - name: the.dest
    type: string
  steps:
  - name: config-docker
    image: gcr.io/tekton-releases/dogfooding/skopeo:latest
    command: ['skopeo']
    args: ['copy', '$(params["the.path"])', '$(params["the.dest"])']
`, helpers.ObjectNameForTest(t), namespace))
			if _, err := c.V1beta1TaskClient.Create(ctx, task, metav1.CreateOptions{}); err != nil {
				t.Fatalf("Failed to create Task `%s`: %s", task.Name, err)
			}
			p := getHelloWorldPipelineWithSingularTask(t, namespace, task.Name)
			if _, err := c.V1beta1PipelineClient.Create(ctx, p, metav1.CreateOptions{}); err != nil {
				t.Fatalf("Failed to create Pipeline `%s`: %s", p.Name, err)
			}

			return p
		},
		expectedTaskRuns: []string{task1Name},
		// 1 from PipelineRun and 1 from Tasks defined in pipelinerun
		expectedNumberOfEvents: 2,
		pipelineRunFunc:        getHelloWorldPipelineRun,
	}, {
		name: "pipelinerun succeeds with LimitRange minimum in namespace",
		testSetup: func(ctx context.Context, t *testing.T, c *clients, namespace string, index int) *v1beta1.Pipeline {
			t.Helper()
			t.Skip("build-crd-testing project got removed, the secret-sauce doesn't exist anymore, skipping")
			if _, err := c.KubeClient.CoreV1().LimitRanges(namespace).Create(ctx, getLimitRange("prlimitrange", namespace, "100m", "99Mi", "100m"), metav1.CreateOptions{}); err != nil {
				t.Fatalf("Failed to create LimitRange `%s`: %s", "prlimitrange", err)
			}

			if _, err := c.KubeClient.CoreV1().Secrets(namespace).Create(ctx, getPipelineRunSecret(index, namespace), metav1.CreateOptions{}); err != nil {
				t.Fatalf("Failed to create secret `%s`: %s", getName(secretName, index), err)
			}

			if _, err := c.KubeClient.CoreV1().ServiceAccounts(namespace).Create(ctx, getPipelineRunServiceAccount(index, namespace), metav1.CreateOptions{}); err != nil {
				t.Fatalf("Failed to create SA `%s`: %s", getName(saName, index), err)
			}

			task := parse.MustParseV1beta1Task(t, fmt.Sprintf(`
metadata:
  name: %s
  namespace: %s
spec:
  params:
  - name: the.path
    type: string
  - name: the.dest
    type: string
  steps:
  - name: config-docker
    image: gcr.io/tekton-releases/dogfooding/skopeo:latest
    command: ['skopeo']
    args: ['copy', '$(params["the.path"])', '$(params["the.dest"])']
`, helpers.ObjectNameForTest(t), namespace))
			if _, err := c.V1beta1TaskClient.Create(ctx, task, metav1.CreateOptions{}); err != nil {
				t.Fatalf("Failed to create Task `%s`: %s", fmt.Sprint("task", index), err)
			}

			p := getHelloWorldPipelineWithSingularTask(t, namespace, task.Name)
			if _, err := c.V1beta1PipelineClient.Create(ctx, p, metav1.CreateOptions{}); err != nil {
				t.Fatalf("Failed to create Pipeline `%s`: %s", p.Name, err)
			}

			return p
		},
		expectedTaskRuns: []string{task1Name},
		// 1 from PipelineRun and 1 from Tasks defined in pipelinerun
		expectedNumberOfEvents: 2,
		pipelineRunFunc:        getHelloWorldPipelineRun,
	}}

	for i, td := range tds {
		i := i   // capture range variable
		td := td // capture range variable
		t.Run(td.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			c, namespace := setup(ctx, t)

			knativetest.CleanupOnInterrupt(func() { tearDown(ctx, t, c, namespace) }, t.Logf)
			defer tearDown(ctx, t, c, namespace)

			t.Logf("Setting up test resources for %q test in namespace %s", td.name, namespace)
			p := td.testSetup(ctx, t, c, namespace, i)

			pipelineRun := td.pipelineRunFunc(t, i, namespace, p.Name)
			prName := pipelineRun.Name
			_, err := c.V1beta1PipelineRunClient.Create(ctx, pipelineRun, metav1.CreateOptions{})
			if err != nil {
				t.Fatalf("Failed to create PipelineRun `%s`: %s", prName, err)
			}

			t.Logf("Waiting for PipelineRun %s in namespace %s to complete", prName, namespace)
			if err := WaitForPipelineRunState(ctx, c, prName, timeout, PipelineRunSucceed(prName), "PipelineRunSuccess", v1beta1Version); err != nil {
				t.Fatalf("Error waiting for PipelineRun %s to finish: %s", prName, err)
			}
			t.Logf("Making sure the expected TaskRuns %s were created", td.expectedTaskRuns)
			actualTaskrunList, err := c.V1beta1TaskRunClient.List(ctx, metav1.ListOptions{LabelSelector: fmt.Sprintf("tekton.dev/pipelineRun=%s", prName)})
			if err != nil {
				t.Fatalf("Error listing TaskRuns for PipelineRun %s: %s", prName, err)
			}
			expectedTaskRunNames := []string{}
			for _, runName := range td.expectedTaskRuns {
				taskRunName := strings.Join([]string{prName, runName}, "-")
				// check the actual task name starting with prName+runName with a random suffix
				for _, actualTaskRunItem := range actualTaskrunList.Items {
					if strings.HasPrefix(actualTaskRunItem.Name, taskRunName) {
						taskRunName = actualTaskRunItem.Name
					}
				}
				expectedTaskRunNames = append(expectedTaskRunNames, taskRunName)
				r, err := c.V1beta1TaskRunClient.Get(ctx, taskRunName, metav1.GetOptions{})
				if err != nil {
					t.Fatalf("Couldn't get expected TaskRun %s: %s", taskRunName, err)
				}
				if !r.Status.GetCondition(apis.ConditionSucceeded).IsTrue() {
					t.Fatalf("Expected TaskRun %s to have succeeded but Status is %v", taskRunName, r.Status)
				}

				t.Logf("Checking that labels were propagated correctly for TaskRun %s", r.Name)
				checkLabelPropagation(ctx, t, c, namespace, prName, r)
				t.Logf("Checking that annotations were propagated correctly for TaskRun %s", r.Name)
				checkAnnotationPropagation(ctx, t, c, namespace, prName, r)
			}

			matchKinds := map[string][]string{"PipelineRun": {prName}, "TaskRun": expectedTaskRunNames}

			t.Logf("Making sure %d events were created from taskrun and pipelinerun with kinds %v", td.expectedNumberOfEvents, matchKinds)

			events, err := collectMatchingEvents(ctx, c.KubeClient, namespace, matchKinds, "Succeeded")
			if err != nil {
				t.Fatalf("Failed to collect matching events: %q", err)
			}
			if len(events) != td.expectedNumberOfEvents {
				collectedEvents := ""
				for i, event := range events {
					collectedEvents += fmt.Sprintf("%#v", event)
					if i < (len(events) - 1) {
						collectedEvents += ", "
					}
				}
				t.Fatalf("Expected %d number of successful events from pipelinerun and taskrun but got %d; list of receieved events : %#v", td.expectedNumberOfEvents, len(events), collectedEvents)
			}

			t.Logf("Successfully finished test %q", td.name)
		})
	}
}
