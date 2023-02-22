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
)

// FanOut produces combinations of Parameters of type String from a slice of Parameters of type Array.
func FanOut(matrix *v1beta1.Matrix) Combinations {
	var combinations Combinations

	// Explicit combinations in the Matrix
	if matrix.MatrixHasInclude() && !matrix.MatrixHasParams() {
		fmt.Println("Explicit")
		return combinations.fanOutMatrixIncludeParams(matrix)
	}

	// Create a set of param names
	paramNamesMap := make(map[string][]string)

	if matrix.MatrixHasParams() {
		for _, param := range matrix.Params {
			combinations = combinations.fanOutInitialMatrixParams(param)
			// add param name to set
			paramNamesMap[param.Name] = param.Value.ArrayVal
		}
	}

	if matrix.MatrixHasInclude() {
		combinations = replaceIncludeMatrixParams(matrix, combinations)
	}
	return combinations

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
