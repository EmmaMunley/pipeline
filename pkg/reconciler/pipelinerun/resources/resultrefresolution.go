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

package resources

import (
	"errors"
	"fmt"
	"sort"

	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

var (
	// ErrInvalidTaskResultReference indicates that the reason for the failure status is that there
	// is an invalid task result reference
	ErrInvalidTaskResultReference = errors.New("Invalid task result reference")
)

// ResolvedResultRefs represents all of the ResolvedResultRef for a pipeline task
type ResolvedResultRefs []*ResolvedResultRef

// ResolvedResultRef represents a result ref reference that has been fully resolved (value has been populated).
// If the value is from a Result, then the ResultReference will be populated to point to the ResultReference
// which resulted in the value
type ResolvedResultRef struct {
	Value           v1.ResultValue
	ResultReference v1.ResultRef
	FromTaskRun     string
	FromRun         string
}

// ResolveResultRef resolves any ResultReference that are found in the target ResolvedPipelineTask
func ResolveResultRef(pipelineRunState PipelineRunState, target *ResolvedPipelineTask) (ResolvedResultRefs, string, error) {
	resolvedResultRefs, pt, err := convertToResultRefs(pipelineRunState, target)
	if err != nil {
		return nil, pt, err
	}
	return removeDup(resolvedResultRefs), "", nil
}

// ResolveResultRefs resolves any ResultReference that are found in the target ResolvedPipelineTask
func ResolveResultRefs(pipelineRunState PipelineRunState, targets PipelineRunState) (ResolvedResultRefs, string, error) {
	var allResolvedResultRefs ResolvedResultRefs
	for _, target := range targets {
		resolvedResultRefs, pt, err := convertToResultRefs(pipelineRunState, target)
		if err != nil {
			return nil, pt, err
		}
		allResolvedResultRefs = append(allResolvedResultRefs, resolvedResultRefs...)
	}
	return removeDup(allResolvedResultRefs), "", nil
}

// validateArrayResultsIndex checks if the result array indexing reference is out of bound of the array size
func validateArrayResultsIndex(allResolvedResultRefs ResolvedResultRefs) error {
	for _, r := range allResolvedResultRefs {
		if r.Value.Type == v1.ParamTypeArray {
			if r.ResultReference.ResultsIndex >= len(r.Value.ArrayVal) {
				return fmt.Errorf("Array Result Index %d for Task %s Result %s is out of bound of size %d", r.ResultReference.ResultsIndex, r.ResultReference.PipelineTask, r.ResultReference.Result, len(r.Value.ArrayVal))
			}
		}
	}
	return nil
}

func removeDup(refs ResolvedResultRefs) ResolvedResultRefs {
	if refs == nil {
		return nil
	}
	resolvedResultRefByRef := make(map[v1.ResultRef]*ResolvedResultRef, len(refs))
	for _, resolvedResultRef := range refs {
		resolvedResultRefByRef[resolvedResultRef.ResultReference] = resolvedResultRef
	}
	deduped := make([]*ResolvedResultRef, 0, len(resolvedResultRefByRef))

	// Sort the resulting keys to produce a deterministic ordering.
	order := make([]v1.ResultRef, 0, len(refs))
	for key := range resolvedResultRefByRef {
		order = append(order, key)
	}
	sort.Slice(order, func(i, j int) bool {
		if order[i].PipelineTask > order[j].PipelineTask {
			return false
		}
		if order[i].Result > order[j].Result {
			return false
		}
		return true
	})

	for _, key := range order {
		deduped = append(deduped, resolvedResultRefByRef[key])
	}
	return deduped
}

// convertToResultRefs walks a PipelineTask looking for result references. If any are
// found they are resolved to a value by searching pipelineRunState. The list of resolved
// references are returned. If an error is encountered due to an invalid result reference
// then a nil list and error is returned instead.
func convertToResultRefs(pipelineRunState PipelineRunState, target *ResolvedPipelineTask) (ResolvedResultRefs, string, error) {
	var resolvedResultRefs ResolvedResultRefs
	for _, ref := range v1.PipelineTaskResultRefs(target.PipelineTask) {
		resolved, pt, err := resolveResultRefs(pipelineRunState, ref)
		if err != nil {
			return nil, pt, err
		}
		resolvedResultRefs = append(resolvedResultRefs, resolved...)
	}
	return resolvedResultRefs, "", nil
}

func resolveResultRefs(pipelineState PipelineRunState, resultRef *v1.ResultRef) (ResolvedResultRefs, string, error) {
	referencedPipelineTask := pipelineState.ToMap()[resultRef.PipelineTask]
	var allResolvedResultRefs ResolvedResultRefs

	if referencedPipelineTask == nil {
		return nil, resultRef.PipelineTask, fmt.Errorf("could not find task %q referenced by result", resultRef.PipelineTask)
	}
	if !referencedPipelineTask.isSuccessful() && !referencedPipelineTask.isFailure() {
		return nil, resultRef.PipelineTask, fmt.Errorf("task %q referenced by result was not finished", referencedPipelineTask.PipelineTask.Name)
	}

	if referencedPipelineTask.IsCustomTask() {

		resolved, pipelineTask, err := resolveResultRefCustomRun(referencedPipelineTask.CustomRuns[0], resultRef)
		if err != nil {
			return nil, pipelineTask, err
		}
		allResolvedResultRefs = append(allResolvedResultRefs, resolved)
	} else {
		if referencedPipelineTask.PipelineTask.IsMatrixed() {
			resolved, pipelineTask, err := resolveMatrixResultRef(referencedPipelineTask.TaskRuns, resultRef)
			if err != nil {
				return nil, pipelineTask, err
			}
			allResolvedResultRefs = append(allResolvedResultRefs, resolved)
		} else {
			resolved, pipelineTask, err := resolveResultRef(referencedPipelineTask.TaskRuns[0], resultRef)
			if err != nil {
				return nil, pipelineTask, err
			}
			allResolvedResultRefs = append(allResolvedResultRefs, resolved)

		}
	}
	return allResolvedResultRefs, "", nil
}

func resolveMatrixResultRef(taskRuns []*v1.TaskRun, resultRef *v1.ResultRef) (*ResolvedResultRef, string, error) {
	resultsMapping := createResultsMapping(taskRuns)
	fmt.Println("resultsMapping?", resultsMapping)
	fmt.Println("resultRef?", resultRef)
	index := resultRef.ResultsIndex

	if arrayValues, ok := resultsMapping[resultRef.Result]; ok {
		fmt.Println("result", arrayValues)
		fmt.Println("index", index)
		// for result, arrayValues := range resultsMapping[resultName] {
		// 	if _, ok := resultsMapping[result]; ok {
		stringVal := arrayValues[index]
		resultValue := v1.ParamValue{
			Type:      v1.ParamTypeString,
			StringVal: stringVal,
		}

		return &ResolvedResultRef{
			Value:           resultValue,
			FromTaskRun:     "taskRunName",
			FromRun:         "runName",
			ResultReference: *resultRef,
		}, "", nil
	}
	return nil, "", nil
}

func resolveResultRef(taskRun *v1.TaskRun, resultRef *v1.ResultRef) (*ResolvedResultRef, string, error) {
	taskRunName := taskRun.Name
	resultValue, err := findTaskResultForParam(taskRun, resultRef)
	if err != nil {
		return nil, resultRef.PipelineTask, err
	}

	return &ResolvedResultRef{
		Value:           resultValue,
		FromTaskRun:     taskRunName,
		FromRun:         "",
		ResultReference: *resultRef,
	}, "", nil
}

func resolveResultRefCustomRun(customRun *v1beta1.CustomRun, resultRef *v1.ResultRef) (*ResolvedResultRef, string, error) {
	runName := customRun.GetObjectMeta().GetName()
	runValue, err := findRunResultForParam(customRun, resultRef)
	resultValue := *v1.NewStructuredValues(runValue)
	if err != nil {
		return nil, resultRef.PipelineTask, err
	}

	return &ResolvedResultRef{
		Value:           resultValue,
		FromTaskRun:     "",
		FromRun:         runName,
		ResultReference: *resultRef,
	}, "", nil
}

func createMap(values []string) map[string][]string {
	mapping := map[string][]string{}
	for _, value := range values {
		key := value[:5]
		if _, ok := mapping[key]; ok {
			mapping[key] = append(mapping[key], value)
		} else {
			mapping[key] = []string{value}
		}
	}
	return mapping
}

func createResultsMapping(taskRuns []*v1.TaskRun) map[string][]string {
	resultsMap := map[string][]string{}
	for _, taskRun := range taskRuns {
		fmt.Println("taskRun", taskRun)
		results := taskRun.Status.Results
		fmt.Println("results?", results)

		for _, result := range results {
			if _, ok := resultsMap[result.Name]; ok {
				val := result.Value
				fmt.Println("val", val)
				resultsMap[result.Name] = append(resultsMap[result.Name], result.Value.StringVal)
			} else {
				resultsMap[result.Name] = []string{result.Value.StringVal}
			}
		}
	}
	fmt.Println("resultsMap?", resultsMap)
	return resultsMap
}

func findRunResultForParam(customRun *v1beta1.CustomRun, reference *v1.ResultRef) (string, error) {
	for _, result := range customRun.Status.Results {
		if result.Name == reference.Result {
			return result.Value, nil
		}
	}
	err := fmt.Errorf("%w: Could not find result with name %s for task %s", ErrInvalidTaskResultReference, reference.Result, reference.PipelineTask)
	return "", err
}

func findTaskResultForParam(taskRun *v1.TaskRun, reference *v1.ResultRef) (v1.ResultValue, error) {
	results := taskRun.Status.TaskRunStatusFields.Results
	for _, result := range results {
		if result.Name == reference.Result {
			return result.Value, nil
		}
	}
	err := fmt.Errorf("%w: Could not find result with name %s for task %s", ErrInvalidTaskResultReference, reference.Result, reference.PipelineTask)
	return v1.ResultValue{}, err
}

func (rs ResolvedResultRefs) getStringReplacements() map[string]string {
	replacements := map[string]string{}
	for _, r := range rs {
		switch r.Value.Type {
		case v1.ParamTypeArray:
			for i := 0; i < len(r.Value.ArrayVal); i++ {
				for _, target := range r.getReplaceTargetfromArrayIndex(i) {
					replacements[target] = r.Value.ArrayVal[i]
				}
			}
		case v1.ParamTypeObject:
			for key, element := range r.Value.ObjectVal {
				for _, target := range r.getReplaceTargetfromObjectKey(key) {
					replacements[target] = element
				}
			}

		case v1.ParamTypeString:
			fallthrough
		default:
			for _, target := range r.getReplaceTarget() {
				replacements[target] = r.Value.StringVal
			}
		}
	}
	return replacements
}

func (rs ResolvedResultRefs) getArrayReplacements() map[string][]string {
	replacements := map[string][]string{}
	for _, r := range rs {
		if r.Value.Type == v1.ParamType(v1.ResultsTypeArray) {
			for _, target := range r.getReplaceTarget() {
				replacements[target] = r.Value.ArrayVal
			}
		}
	}
	return replacements
}

func (rs ResolvedResultRefs) getObjectReplacements() map[string]map[string]string {
	replacements := map[string]map[string]string{}
	for _, r := range rs {
		if r.Value.Type == v1.ParamType(v1.ResultsTypeObject) {
			for _, target := range r.getReplaceTarget() {
				replacements[target] = r.Value.ObjectVal
			}
		}
	}
	return replacements
}

func (r *ResolvedResultRef) getReplaceTarget() []string {
	return []string{
		fmt.Sprintf("%s.%s.%s.%s", v1.ResultTaskPart, r.ResultReference.PipelineTask, v1.ResultResultPart, r.ResultReference.Result),
		fmt.Sprintf("%s.%s.%s[%q]", v1.ResultTaskPart, r.ResultReference.PipelineTask, v1.ResultResultPart, r.ResultReference.Result),
		fmt.Sprintf("%s.%s.%s['%s']", v1.ResultTaskPart, r.ResultReference.PipelineTask, v1.ResultResultPart, r.ResultReference.Result),
	}
}

func (r *ResolvedResultRef) getReplaceTargetfromArrayIndex(idx int) []string {
	return []string{
		fmt.Sprintf("%s.%s.%s.%s[%d]", v1.ResultTaskPart, r.ResultReference.PipelineTask, v1.ResultResultPart, r.ResultReference.Result, idx),
		fmt.Sprintf("%s.%s.%s[%q][%d]", v1.ResultTaskPart, r.ResultReference.PipelineTask, v1.ResultResultPart, r.ResultReference.Result, idx),
		fmt.Sprintf("%s.%s.%s['%s'][%d]", v1.ResultTaskPart, r.ResultReference.PipelineTask, v1.ResultResultPart, r.ResultReference.Result, idx),
	}
}

func (r *ResolvedResultRef) getReplaceTargetfromObjectKey(key string) []string {
	return []string{
		fmt.Sprintf("%s.%s.%s.%s.%s", v1.ResultTaskPart, r.ResultReference.PipelineTask, v1.ResultResultPart, r.ResultReference.Result, key),
		fmt.Sprintf("%s.%s.%s[%q][%s]", v1.ResultTaskPart, r.ResultReference.PipelineTask, v1.ResultResultPart, r.ResultReference.Result, key),
		fmt.Sprintf("%s.%s.%s['%s'][%s]", v1.ResultTaskPart, r.ResultReference.PipelineTask, v1.ResultResultPart, r.ResultReference.Result, key),
	}
}
