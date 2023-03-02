/*
Copyright 2023 The Tekton Authors
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

package v1

import (
	"context"
	"fmt"

	"github.com/tektoncd/pipeline/pkg/apis/config"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/strings/slices"
	"knative.dev/pkg/apis"
)

// Matrix is used to fan out Tasks in a Pipeline
type Matrix struct {
	// Params is a list of parameters used to fan out the pipelineTask
	// Params takes only `Parameters` of type `"array"`
	// Each array element is supplied to the `PipelineTask` by substituting `params` of type `"string"` in the underlying `Task`.
	// The names of the `params` in the `Matrix` must match the names of the `params` in the underlying `Task` that they will be substituting.
	// +listType=atomic
	Params Params `json:"params,omitempty"`

	// Include is a list of MatrixInclude which allows passing in specific combinations of Parameters into the Matrix.
	// Note that Include is in preview mode and not yet supported.
	// +optional
	// +listType=atomic
	Include []MatrixInclude `json:"include,omitempty"`
}

// MatrixInclude allows passing in a specific combinations of Parameters into the Matrix.
// Note this struct is in preview mode and not yet supported
type MatrixInclude struct {
	// Name the specified combination
	Name string `json:"name,omitempty"`

	// Params takes only `Parameters` of type `"string"`
	// The names of the `params` must match the names of the `params` in the underlying `Task`
	// +listType=atomic
	Params Params `json:"params,omitempty"`
}

// FanOut produces combinations of Parameters of type String from a slice of Parameters of type Array.
func (m Matrix) FanOut() []Params {
	var combinations []Params
	// If Matrix Include params exists, generate explicit combinations
	if m.hasInclude() && !m.hasParams() {
		return fanOutExplicitCombinations(m.Include, combinations)
	}
	// Generate initial combinations with matrx.Params
	for _, parameter := range m.Params {
		combinations = fanOut(parameter, combinations)
	}
	// Replace or append combinations with Matrix Include Params
	if m.hasInclude() {

		mappedMatrixIncludeParamsSlice := mapMatrixIncludeParams(m.Include)
		fmt.Println("mappedMatrixIncludeParamsSlice", mappedMatrixIncludeParamsSlice)

		matrixParamsMap := mapMatrixParams(m.Params)
		fmt.Println("matrixParamsMap", matrixParamsMap)

		printCombinations(combinations)
		combinations = replaceCombinations(mappedMatrixIncludeParamsSlice, combinations)
		printCombinations(combinations)
		combinations = appendMissingValues(mappedMatrixIncludeParamsSlice, combinations)
		combinations = generateNewCombinations(mappedMatrixIncludeParamsSlice, combinations, matrixParamsMap)
	}
	printCombinations(combinations)
	return combinations
}

// replaceCombinations handles the use case when there are some param name and values in combinations that
// match and other param name and values that are missing entirely by filtering the combinations that
// match the include params and appending the missing params at the given combination. It returns the modified
// combinations.
func replaceCombinations(mappedMatrixIncludeParamsSlice []map[string]string, combinations []Params) []Params {
	// Filter out params to only include new params
	for _, matrixIncludeParamMap := range mappedMatrixIncludeParamsSlice {
		// Len must be > 1
		if len(matrixIncludeParamMap) <= 1 {
			continue
		}
		hasAtLeastOneMatch := hasAtLeastOneMatch(combinations, matrixIncludeParamMap)
		if hasAtLeastOneMatch {
			combinations = filterCombinations(combinations, matrixIncludeParamMap)
		}
	}
	return combinations
}

// hasAtLeastOneMatch checks if at least one include param name and values exists in combinations
func hasAtLeastOneMatch(combinations []Params, paramNamesMap map[string]string) bool {
	// Check at least one include param name and values exists in combinations
	for _, combinationParams := range combinations {
		for _, combinationParam := range combinationParams {
			if val, exist := paramNamesMap[combinationParam.Name]; exist {
				if val == combinationParam.Value.StringVal {
					return true
				}
			}
		}
	}
	return false
}

// filterCombinations iterates over combinations with the matrix include parameters and checks that
// there is at least once matching include parameter name and value that exist in combinations and
// ensures that if the param name exists in both, the value must also match to indicate there are missing
// params that need to be appended to existing combinations. It returns modified combinations.
func filterCombinations(combinations []Params, matrixIncludeParamMap map[string]string) []Params {
	for i, combination := range combinations {
		combinationMap := mapCombination(combination)
		matchedParamsCount := 0
		missingParamsCount := 0
		var missingParams Params
		for name, val := range matrixIncludeParamMap {
			if combinationVal, ok := combinationMap[name]; ok {
				if combinationVal == val {
					matchedParamsCount++
				}
			} else {
				missingParamsCount++
				missingParams = append(missingParams, createCombinationParam(name, val))
			}

		}
		if matchedParamsCount+missingParamsCount == len(matrixIncludeParamMap) {
			// replace missing values
			for _, missingParam := range missingParams {
				printCombinations(combinations)
				combinations[i] = append(combinations[i], missingParam)
			}
		}
	}
	return combinations
}

// appendMissingValues handles the use case scenarios when there are some param name and values in combinations
// that match and other param name and values that are missing entirely. This filters the combinations that
// match the include params and append the missing params at the given combination. It returns the modified
// combinations.
func appendMissingValues(mappedMatrixIncludeParamsSlice []map[string]string, combinations []Params) []Params {
	for _, matrixIncludeParamMap := range mappedMatrixIncludeParamsSlice {
		isMissing := paramMissingFromAllCombinations(matrixIncludeParamMap, combinations)
		if isMissing {
			for name, val := range matrixIncludeParamMap {
				fmt.Println(name)
				fmt.Println(val)
				for i := range combinations {
					combinations[i] = append(combinations[i], createCombinationParam(name, val))
				}
			}
		}
	}
	return combinations
}

// generateNewCombinations handles the use case when there is a matching param name but the value is missing from
// the initial combinations so a new combination needs to . It returns the modified combinations.
func generateNewCombinations(mappedMatrixIncludeParamsSlice []map[string]string, combinations []Params, matrixParamsMap map[string][]string) []Params {
	for _, matrixIncludeParamMap := range mappedMatrixIncludeParamsSlice {
		paramValueNotFound := paramValueNotFound(matrixIncludeParamMap, matrixParamsMap)
		if paramValueNotFound {
			for name, val := range matrixIncludeParamMap {
				new := createCombination(name, val, []Param{})
				combinations = append(combinations, new)
			}
		}
	}
	return combinations
}

// paramValueNotFound returns false if the Matrix Include Param value does not exist in Matrix Params for a
// given param name
func paramValueNotFound(matrixIncludeParamMap map[string]string, matrixParamsMap map[string][]string) bool {
	for name, val := range matrixIncludeParamMap {
		if matrixVal, ok := matrixParamsMap[name]; ok {
			if !slices.Contains(matrixVal, val) {
				return true
			}
		}
	}
	return false
}

// paramMissingFromAllCombinations returns true if the matrix include parameter name is missing from all combinations
func paramMissingFromAllCombinations(matrixIncludeParamMap map[string]string, combinations []Params) bool {
	// The parameter name has to be missing from all combinations
	for _, combination := range combinations {
		for _, combinationParam := range combination {
			if _, exist := matrixIncludeParamMap[combinationParam.Name]; exist {
				// value exists in a combination
				return false
			}
		}
	}
	return true
}

// fanOut generates a new combination based on a given Parameter in the Matrix.
func fanOut(param Param, combinations []Params) []Params {
	if len(combinations) == 0 {
		return initializeCombinations(param)
	}
	return distribute(param, combinations)
}

// initializeCombinations generates a new combination based on the first Parameter in the Matrix.
func initializeCombinations(param Param) []Params {
	var combinations []Params
	for _, value := range param.Value.ArrayVal {
		combinations = append(combinations, createCombination(param.Name, value, []Param{}))
	}
	return combinations
}

// distribute generates a new combination of Parameters by adding a new Parameter to an existing list of Combinations.
func distribute(param Param, combinations []Params) []Params {
	var expandedCombinations []Params
	for _, value := range param.Value.ArrayVal {
		for _, combination := range combinations {
			expandedCombinations = append(expandedCombinations, createCombination(param.Name, value, combination))
		}
	}
	return expandedCombinations
}

// fanOutExplicitCombinations handles the use case when there are only matrix include params and no matrix
// params to generate explicit combinations
func fanOutExplicitCombinations(matrixInclude []MatrixInclude, combinations []Params) []Params {
	for i := 0; i < len(matrixInclude); i++ {
		includeParams := matrixInclude[i].Params
		combination := []Param{}

		for _, param := range includeParams {
			newCombination := createCombinationParam(param.Name, param.Value.StringVal)
			combination = append(combination, newCombination)
		}
		combinations = append(combinations, combination)
	}
	return combinations
}

func createCombination(name string, value string, combination Params) Params {
	combination = append(combination, Param{
		Name:  name,
		Value: ParamValue{Type: ParamTypeString, StringVal: value},
	})
	return combination
}

// createCombinationParam creates and returns a new combination param
func createCombinationParam(name string, value string) Param {
	return Param{
		Name:  name,
		Value: ParamValue{Type: ParamTypeString, StringVal: value},
	}
}

func printCombinations(combinations []Params) {
	for _, combination := range combinations {
		fmt.Println(combination)
	}
}

// mapMatrixIncludeParams returs a slice of mapped params with the key: param.Name
// and the val: param.Val.StringVal
func mapMatrixIncludeParams(matrixInclude []MatrixInclude) []map[string]string {
	var mappedMatrixIncludeParamsSlice []map[string]string
	for _, include := range matrixInclude {
		paramMap := make(map[string]string)
		for _, param := range include.Params {
			paramMap[param.Name] = param.Value.StringVal
		}
		mappedMatrixIncludeParamsSlice = append(mappedMatrixIncludeParamsSlice, paramMap)
	}
	return mappedMatrixIncludeParamsSlice
}

// mapCombinations returns a map of combinations with the key params.Name and the
// val: params.Value.StringVal
func mapCombination(combination Params) map[string]string {
	combinationMap := make(map[string]string)
	for _, params := range combination {
		combinationMap[params.Name] = params.Value.StringVal
	}
	return combinationMap
}

// CountCombinations returns the count of combinations of Parameters generated from the Matrix in PipelineTask.
func (m *Matrix) CountCombinations() int {
	// Iterate over matrix.params and compute count of all generated combinations
	count := m.countGeneratedCombinationsFromParams()

	// Add any additional combinations generated from matrix include params
	count += m.countNewCombinationsFromInclude()

	return count
}

// countGeneratedCombinationsFromParams returns the count of combinations of Parameters generated from the matrix
// parameters
func (m *Matrix) countGeneratedCombinationsFromParams() int {
	if !m.hasParams() {
		return 0
	}
	count := 1
	for _, param := range m.Params {
		count *= len(param.Value.ArrayVal)
	}
	return count
}

// countNewCombinationsFromInclude returns the count of combinations of Parameters generated from the matrix
// include parameters
func (m *Matrix) countNewCombinationsFromInclude() int {
	if !m.hasInclude() {
		return 0
	}
	if !m.hasParams() {
		return len(m.Include)
	}
	count := 0
	matrixParamMap := m.Params.extractParamMapArrVals()
	for _, include := range m.Include {
		for _, param := range include.Params {
			if val, exist := matrixParamMap[param.Name]; exist {
				// If the matrix include param values does not exist, a new combination will be generated
				if !slices.Contains(val, param.Value.StringVal) {
					count++
				} else {
					break
				}
			}
		}
	}
	return count
}

func (m *Matrix) hasInclude() bool {
	return m != nil && m.Include != nil && len(m.Include) > 0
}

func (m *Matrix) hasParams() bool {
	return m != nil && m.Params != nil && len(m.Params) > 0
}

func (m *Matrix) validateCombinationsCount(ctx context.Context) (errs *apis.FieldError) {
	matrixCombinationsCount := m.CountCombinations()
	maxMatrixCombinationsCount := config.FromContextOrDefaults(ctx).Defaults.DefaultMaxMatrixCombinationsCount
	if matrixCombinationsCount > maxMatrixCombinationsCount {
		errs = errs.Also(apis.ErrOutOfBoundsValue(matrixCombinationsCount, 0, maxMatrixCombinationsCount, "matrix"))
	}
	return errs
}

// validateParamTypes validates the type of parameter
// for Matrix.Params and Matrix.Include.Params
// Matrix.Params must be of type array. Matrix.Include.Params must be of type string.
func (m *Matrix) validateParamTypes() (errs *apis.FieldError) {
	if m != nil {
		if m.hasInclude() {
			for _, include := range m.Include {
				for _, param := range include.Params {
					if param.Value.Type != ParamTypeString {
						errs = errs.Also(apis.ErrInvalidValue(fmt.Sprintf("parameters of type string only are allowed, but got param type %s", string(param.Value.Type)), "").ViaFieldKey("matrix.include.params", param.Name))
					}
				}
			}
		}
		if m.hasParams() {
			for _, param := range m.Params {
				if param.Value.Type != ParamTypeArray {
					errs = errs.Also(apis.ErrInvalidValue(fmt.Sprintf("parameters of type array only are allowed, but got param type %s", string(param.Value.Type)), "").ViaFieldKey("matrix.params", param.Name))
				}
			}
		}
	}
	return errs
}

// validatePipelineParametersVariablesInMatrixParameters validates all pipeline paramater variables including Matrix.Params and Matrix.Include.Params
// that may contain the reference(s) to other params to make sure those references are used appropriately.
func (m *Matrix) validatePipelineParametersVariablesInMatrixParameters(prefix string, paramNames sets.String, arrayParamNames sets.String, objectParamNameKeys map[string][]string) (errs *apis.FieldError) {
	if m.hasInclude() {
		for _, include := range m.Include {
			for idx, param := range include.Params {
				stringElement := param.Value.StringVal
				// Matrix Include Params must be of type string
				errs = errs.Also(validateStringVariable(stringElement, prefix, paramNames, arrayParamNames, objectParamNameKeys).ViaFieldIndex("", idx).ViaField("matrix.include.params", ""))
			}
		}
	}
	if m.hasParams() {
		for _, param := range m.Params {
			for idx, arrayElement := range param.Value.ArrayVal {
				// Matrix Params must be of type array
				errs = errs.Also(validateArrayVariable(arrayElement, prefix, paramNames, arrayParamNames, objectParamNameKeys).ViaFieldIndex("value", idx).ViaFieldKey("matrix.params", param.Name))
			}
		}
	}
	return errs
}

func (m *Matrix) validateParameterInOneOfMatrixOrParams(params []Param) (errs *apis.FieldError) {
	matrixParameterNames := sets.NewString()
	if m != nil {
		for _, param := range m.Params {
			matrixParameterNames.Insert(param.Name)
		}
	}
	for _, param := range params {
		if matrixParameterNames.Has(param.Name) {
			errs = errs.Also(apis.ErrMultipleOneOf("matrix["+param.Name+"]", "params["+param.Name+"]"))
		}
	}
	return errs
}

// mapMatrixParams returs a mapped params with the key: param.Name
// and the val: param.Val.ArrayVal
func mapMatrixParams(matrixParam Params) map[string][]string {
	matrixParamsMap := make(map[string][]string)
	for _, param := range matrixParam {
		matrixParamsMap[param.Name] = param.Value.ArrayVal
	}
	return matrixParamsMap
}
