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
	"golang.org/x/exp/maps"
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

// Combinations is the collection of combination maps
type combinations []combination

// Combination maps the param name to the param value
type combination map[string]string

// FanOut produces combinations of Parameters of type String from a slice of Parameters of type Array.
func (m Matrix) FanOut() []Params {
	var cs combinations

	// Otherwise generate initial combinations with matrix.Params of type mapped combinations that
	// will later be converted back to a slice of Parameters of type Array
	for _, param := range m.Params {
		cs = cs.fanOut(param)
	}
	combinationParams := cs.convertToParams()
	return combinationParams
}

// fanOut generates a new combination based on a given Parameter in the Matrix.
func (cs combinations) fanOut(param Param) combinations {
	if len(cs) == 0 {
		return initializeCombinations(param)
	}
	return cs.distribute(param)
}

// initializeCombinations generates a new combination based on the first Parameter in the Matrix.
func initializeCombinations(param Param) combinations {
	var cs combinations
	for _, value := range param.Value.ArrayVal {
		cs = append(cs, map[string]string{param.Name: value})
	}
	return cs
}

// distribute generates a new combination of Parameters by adding a new Parameter to an existing list of Combinations.
func (cs combinations) distribute(param Param) combinations {
	var expandedCombinations combinations
	for _, value := range param.Value.ArrayVal {
		for _, c := range cs {
			newCombination := make(map[string]string)
			maps.Copy(newCombination, c)
			newCombination[param.Name] = value
			expandedCombinations = append(expandedCombinations, newCombination)
		}
	}
	return expandedCombinations
}

// convertToParams converts mapped combinations to an array of params for the reconiler
// to consume
func (cs combinations) convertToParams() []Params {
	var finalParams []Params
	for _, combination := range cs {
		var params Params
		for name, val := range combination {
			params = append(params, createCombinationParam(name, val))
		}
		finalParams = append(finalParams, params)
	}
	return finalParams
}

// createCombinationParam creates and returns a new combination param
func createCombinationParam(name string, value string) Param {
	return Param{
		Name:  name,
		Value: ParamValue{Type: ParamTypeString, StringVal: value},
	}
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
