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

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"k8s.io/utils/strings/slices"
)

// FanOut produces combinations of Parameters of type String from a slice of Parameters of type Array.
func FanOut(matrix v1beta1.Matrix) Combinations {
	var combinations Combinations

	// If Matrix Include params exists, generate explicit combinations
	if matrix.MatrixHasInclude() && !matrix.MatrixHasParams() {
		return combinations.fanOutMatrixIncludeParams(matrix)
	}

	// Generate initial combinations with matrx.Params
	for _, parameter := range matrix.Params {
		combinations = combinations.fanOut(parameter)
	}

	// Matrix has Include and Params
	// uses the combinations from above
	if matrix.MatrixHasInclude() {

		// CREATE MAPS
		mappedMatrixIncludeParamsSlice := mapMatrixIncludeParams(matrix.Include)
		fmt.Println("mappedParamsSlice", mappedMatrixIncludeParamsSlice)

		matrixParamNames := createSetMatrixParamName(matrix.Params)

		fmt.Println("matrixParamNames", matrixParamNames)

		matrixParamsMap := mapMatrixParams(matrix.Params)
		fmt.Println("matrixParamsMap", matrixParamsMap)

		// mappedCombinationsSlice := mapCombinations(combinations)
		// fmt.Println("mappedCombinationsSlice", mappedCombinationsSlice)
		// FILTER 3 works only for replaceCombinations not with appendMissingValues
		combinations = replaceCombinations(mappedMatrixIncludeParamsSlice, combinations, matrixParamsMap)
		combinations = appendMissingValues(mappedMatrixIncludeParamsSlice, combinations, matrixParamsMap)

	}

	return combinations
}

// FILTER 3 WORKS!!!
func replaceCombinations(mappedMatrixIncludeParamsSlice []map[string]string, combinations Combinations, matrixParamsMap map[string][]string) Combinations {
	var finalCombinations Combinations
	// Filter out combinations and replace any missing values in combination params

	// Filter out params to only include new params
	combination := Filter(combinations, mappedMatrixIncludeParamsSlice)
	fmt.Println("filtered", combination)
	printCombinations(finalCombinations)
	fmt.Println("REPLACING COMBINATIONS")
	printCombinations(combinations)
	return combinations
}

// Passing Common Package
// MOVE INTO ELSE
func appendMissingValues(mappedMatrixIncludeParamsSlice []map[string]string, combinations Combinations, matrixParamsMap map[string][]string) Combinations {
		// The parameter name has to be missing from all combinations

		// Add new params at the current combination.MatrixID
		for _, combination := range combinations {
			fmt.Println("BEFORE APPENDING PARAMS:", combination.Params)
			// combination.Params = append(combination.Params, createNewStringParam(name, val))
			fmt.Println("AFTER APPENDING PARAMS:", combination.Params)
		}

	printCombinations(combinations)
	return combinations
}


// Passing Common Package
// MOVE INTO ELSE
func paramMissingFromAllCombinations(mappedMatrixIncludeParamsSlice []map[string]string, combinations Combinations) Combinations {
	// The parameter name has to be missing from all combinations

	// Add new params at the current combination.MatrixID
	for _, combination := range combinations {
		for _, matrixIncludeParamMap := range mappedMatrixIncludeParamsSlice {
			for name, val := range matrixIncludeParamMap {
		fmt.Println("BEFORE APPENDING PARAMS:", combination.Params)
		// combination.Params = append(combination.Params, createNewStringParam(name, val))
		fmt.Println("AFTER APPENDING PARAMS:", combination.Params)
		}
	}

printCombinations(combinations)
return combinations
}
