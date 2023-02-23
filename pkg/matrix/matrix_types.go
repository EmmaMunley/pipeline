/*
   Copyright 2022 The Tekton Authors
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

package matrix

import (
	"fmt"
	"strconv"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

// Combinations is a slice of combinations of Parameters from a Matrix.
type Combinations []*Combination

// Combination is a specific combination of Parameters from a Matrix.
type Combination struct {
	// MatrixID is an identification of a combination from Parameters in a Matrix.
	MatrixID string

	// Params is a specific combination of Parameters in a Matrix.
	Params []v1beta1.Param
}

func (combinations Combinations) fanOutInitialMatrixParams(param v1beta1.Param) Combinations {
	if len(combinations) == 0 {
		return initializeCombinations(param)
	}
	return combinations.distribute(param)
}

func (combinations Combinations) fanOutMatrixIncludeParams(matrix *v1beta1.Matrix) Combinations {
	count := 0
	for i := 1; i < len(matrix.Include); i += 2 {
		include := matrix.Include[i]
		params := include.Params
		fmt.Println("params", params)
		combinations = append(combinations, createIncludeCombination(count, params))
		count++
	}
	return combinations
}

func replaceIncludeMatrixParams(matrix *v1beta1.Matrix, combinations Combinations) Combinations {
	var finalCombinations Combinations

	// Create [][] mapped Params to iterate over instead of nested Params

	var mappedParamsSlice []map[string]string

	for _, matrixInclude := range matrix.Include {
		paramNamesMap := make(map[string]string)
		if len(matrixInclude.Params) > 0 {
			for _, param := range matrixInclude.Params {
				// add param name to set
				paramNamesMap[param.Name] = param.Value.StringVal
			}
			mappedParamsSlice = append(mappedParamsSlice, paramNamesMap)
		}
	}
	fmt.Println("mappedParamsSlice", mappedParamsSlice)

	var matrixParamNames []string
	for _, matrixParam := range matrix.Params {
		// add param name to set
		matrixParamNames = append(matrixParamNames, matrixParam.Name)
	}

	fmt.Println("matrixParamNames", matrixParamNames)

	matrixParamsMap := make(map[string][]string)
	for _, matrixParams := range matrix.Params {
		if len(matrixParams.Value.ArrayVal) > 0 {
				matrixParamsMap[matrixParams.Name] = matrixParams.Value.ArrayVal
		}
	}
	fmt.Println("matrixParamsMap", matrixParamsMap)

	var mappedCombinationsSlice []map[string]string

	for _, combinations := range combinations {
		combinationsMap := make(map[string]string)
		for _, combination := range combinations.Params {
				// add param name to set
				combinationsMap[combination.Name] = combination.Value.StringVal
			}
			mappedCombinationsSlice = append(mappedCombinationsSlice, combinationsMap)
		}


		// Use case: some or all matching params
		for i := 0; i < len(mappedParamsSlice); i++ {
		paramMap := mappedParamsSlice[i]
		// Filter out params to only include new params
		finalCombinations = Filter(combinations, paramMap)



		for name, val := range paramMap {
			// USE CASE
			// handle the use case where the name does not exist and a new combo is appended to combinations
			if !contains(matrixParamNames, name) {
				fmt.Println("match not found", name, val)
				// Add new params at the current combination.MatrixID
				for _, combination := range finalCombinations {
					fmt.Println("BEFORE APPENDING PARAMS:", combination.Params)
					combination.Params = append(combination.Params, createNewStringParam(name, val))
					fmt.Println("AFTER APPENDING PARAMS:", combination.Params)
				}
			}
		}
		}

		// LOGIC FOR COMMON PARAMS TO APPEND THAT ARE NOT IN PARAMS BUT ARE IN INCLUDE PARAMS

		name, val := paramValueNotFound(mappedParamsSlice, matrixParamsMap)
		combinationLen := len(combinations)
		fmt.Println("len Combination", combinationLen)
		if (name != "" && val != "") {
			new := createCombinationOnly(combinationLen,name, val)
			fmt.Println("new Combination??", new)
			fmt.Println("AFTER ADDING NEW COMBO:", combinationLen)
			combinations = append(combinations, new)
			fmt.Println("AFTER ADDING NEW COMBO:", combinationLen)
		}

	// // Otherwise the param value does not exist
	// // New combination generated
	fmt.Println("finalCombinations", finalCombinations.ToMap())
	return finalCombinations
}

func (combinations Combinations) distribute(param v1beta1.Param) Combinations {
	// when there are existing combinations, this is a non-first parameter in the matrix, and we need to distribute
	// it among the existing combinations
	var expandedCombinations Combinations
	var count int
	for _, value := range param.Value.ArrayVal {
		for _, combination := range combinations {
			expandedCombinations = append(expandedCombinations, createCombination(count, param.Name, value, combination.Params))
			count++
		}
	}
	return expandedCombinations
}

func initializeCombinations(param v1beta1.Param) Combinations {
	// when there are no existing combinations, this is the first parameter in the matrix, so we initialize the
	// combinations with the first Parameter
	var combinations Combinations
	for i, value := range param.Value.ArrayVal {
		combinations = append(combinations, createCombination(i, param.Name, value, []v1beta1.Param{}))
	}
	return combinations
}

func createCombination(i int, name string, value string, parameters []v1beta1.Param) *Combination {
	return &Combination{
		MatrixID: strconv.Itoa(i),
		Params: append(parameters, v1beta1.Param{
			Name:  name,
			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: value},
		}),
	}
}

func createCombinationOnly(i int, name string, value string) *Combination {
	return &Combination{
		MatrixID: strconv.Itoa(i),
		Params: []v1beta1.Param{{
			Name:  name,
			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: value},
		}},
	}
}

func createIncludeCombination(i int, parameters []v1beta1.Param) *Combination {
	combination := &Combination{
		MatrixID: strconv.Itoa(i),
		Params:   []v1beta1.Param{},
	}
	for _, param := range parameters {
		newParam := v1beta1.Param{
			Name:  param.Name,
			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: param.Value.StringVal},
		}
		fmt.Println("newParam", newParam)
		combination.Params = append(combination.Params, newParam)
	}
	return combination
}

// ToMap converts a list of Combinations to a map where the key is the matrixId and the values are Parameters.
func (combinations Combinations) ToMap() map[string][]v1beta1.Param {
	m := map[string][]v1beta1.Param{}
	for _, combination := range combinations {
		m[combination.MatrixID] = combination.Params
	}
	return m
}

func Filter(combinations Combinations, paramMap map[string]string) Combinations {
	// len := len(combinations)
	for _, combination := range combinations {
		// id := combination.MatrixID

		// Check if all the values in combinationParams exist in includeParams
		matchesAllValues := matchesAllValues(combination.Params, paramMap)
		if matchesAllValues {
			fmt.Println("combination.Params", combination.Params)
			fmt.Println("paramMap", paramMap)

			filteredParams := removeDupsInParamMap(combination.Params, paramMap)
			fmt.Println("filteredParams", filteredParams)
			// Add new params at the current combination.MatrixID
			for name, val := range filteredParams {
				// fmt.Println("BEFORE APPENDING PARAMS:", combination.Params)
				combination.Params = append(combination.Params, createNewStringParam(name, val))
				// fmt.Println("AFTER APPENDING PARAMS:", combination.Params)
			}
		}



	}
	return combinations
}

// createNewStringParam creates a new string param for include params
func createNewStringParam(name string, val string) v1beta1.Param {
	newParam := v1beta1.Param{
		Name:  name,
		Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: val},
	}
	//OUTPUT IS 	{Name: "flags", Value: v1beta1.ParamValue{Type: "string", StringVal: "-cover -v"}},
	// WANT {Name: "version", Value: {Type: "string", StringVal: "go1.18.1"}},
	return newParam
}

func matchesAllValues(combinationParams []v1beta1.Param, paramNamesMap map[string]string) bool {
	// TAKE EACH COMBINATION CHECK IF EACH COMBINATION PARAM NAME EXISTS IN PARAM MAPPING
	// IF IT DOESN"T EXIST OR IT"S VALUE DOES NOT MATCH RETURN FALSE
	for _, combinationParam := range combinationParams {
		// check does each seeking param exist in the combinationParams
		// if it doesn't exist break the loop
		if val, exist := paramNamesMap[combinationParam.Name]; exist {
			if val != combinationParam.Value.StringVal {
				return false
			}
		} else { //value not found in mapping
			return false
		}
	}
	return true
}

func findMissingParamValues(combinationParams []v1beta1.Param, paramNamesMap map[string]string) (string, string) {
	for _, combinationParam := range combinationParams {
		// check does each seeking param exist in the combinationParams
		// if it doesn't exist break the loop
		if val, exist := paramNamesMap[combinationParam.Name]; exist {
			if val != combinationParam.Value.StringVal {
				return combinationParam.Name, val
			}
		}
	}
	return "",""
}

// NO PARAM VALUES
func paramValueNotFound(paramNamesMap []map[string]string, matrixParamsMap map[string][]string) (string, string)  {
	for _, paramMap := range paramNamesMap {
		// Get the name
		for name, val := range paramMap {
			if matrixVal, exists := matrixParamsMap[name]; exists {
				fmt.Println("NAME EXISTS", matrixVal)
				if (!contains(matrixVal, val)) {
					fmt.Println("VAL DOES NOT EXIST")
					fmt.Println("VAL", val)
						// create a combination
						return name, val
				}

			}
		}
	}
	return "",""
}

func countParamNames(combinationParams []v1beta1.Param, paramNamesMap map[string]string) int {
	count := 0
	for _, combinationParam := range combinationParams {
		// check does each seeking param exist in the combinationParams
		// if it doesn't exist break the loop
		if _, exist := paramNamesMap[combinationParam.Name]; exist {
			count++
		}
	}
	return count
}

func countParamValues(combinationParams []v1beta1.Param, paramNamesMap map[string]string) int {
	count := 0
	for _, combinationParam := range combinationParams {
		// check does each seeking param exist in the combinationParams
		// if it doesn't exist break the loop
		if val, exist := paramNamesMap[combinationParam.Name]; exist {
			if val == combinationParam.Value.StringVal {
				count++
			}
		}
	}
	return count
}

// Remove duplicate params in param map by de-duping with combination params to find the
// new param to append to combination params
func removeDupsInParamMap(combinationParams []v1beta1.Param, paramNamesMap map[string]string) map[string]string {
	fmt.Println("removeDupsInParamMap")
	fmt.Println("BEFORE:", paramNamesMap)
	for _, combinationParam := range combinationParams {
		// check does each seeking param exist in the combinationParams
		if val, exist := paramNamesMap[combinationParam.Name]; exist {
			if val == combinationParam.Value.StringVal {
				delete(paramNamesMap, combinationParam.Name)
			}
		}
	}
	fmt.Println("AFTER:", paramNamesMap)
	return paramNamesMap
}

// Contains returns true if a string exists in a slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func printCombinations(combinations Combinations) {
	for _, combination := range combinations {
		params := combination.Params
		fmt.Println("ID", combination.MatrixID)
		fmt.Println("params", params)
	}
}
