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
	"testing"

	"github.com/tektoncd/pipeline/test/parse"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	knativetest "knative.dev/pkg/test"
)

func TestFailingStepOnContinue(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	c, namespace := setup(ctx, t)
	knativetest.CleanupOnInterrupt(func() { tearDown(ctx, t, c, namespace) }, t.Logf)
	defer tearDown(ctx, t, c, namespace)
	taskRunName := "mytaskrun"
	tr := parse.MustParseV1beta1TaskRun(t, fmt.Sprintf(`
metadata:
  name: %s
  namespace: %s
spec:
  taskSpec:
    results:
      - name: result1
      - name: result2
    steps:
      - name: failing-step
        onError: continue
        image: busybox
        script: 'echo -n 123 | tee $(results.result1.path); exit 1; echo -n 456 | tee $(results.result2.path)'
`, taskRunName, namespace))

	if _, err := c.V1beta1TaskRunClient.Create(ctx, tr, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Failed to create TaskRun: %s", err)
	}

	t.Logf("Waiting for TaskRun in namespace %s to finish", namespace)
	if err := WaitForTaskRunState(ctx, c, taskRunName, TaskRunSucceed(taskRunName), "TaskRunSucceeded", v1beta1Version); err != nil {
		t.Errorf("Error waiting for TaskRun to finish: %s", err)
	}

	taskRun, err := c.V1beta1TaskRunClient.Get(ctx, taskRunName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Couldn't get expected TaskRun %s: %s", taskRunName, err)
	}

	if !taskRun.Status.GetCondition(apis.ConditionSucceeded).IsTrue() {
		t.Fatalf("Expected TaskRun %s to have succeeded but Status is %v", taskRunName, taskRun.Status)
	}

	if len(taskRun.Status.TaskRunResults) != 1 {
		t.Fatalf("task1 should have produced a result before failing the step")
	}

	for _, r := range taskRun.Status.TaskRunResults {
		if r.Name == "taskRunesult1" && r.Value.StringVal != "123" {
			t.Fatalf("task1 should have initialized a result \"result1\" to \"123\"")
		}
	}
}
