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

func replaceIncludeMatrixParams(matrix *v1beta1.Matrix, combinations Combinations, paramNamesMap map[string][]string) Combinations {
	// handle the use case where the name does not exist and a new combo is appended to combinations
	// for _, include := range matrix.Include {
	// 	for _, param := range include.Params {
	// 		if _, exist := paramNamesMap[param.Name]; !exist {
	// 						// // THINK THIS WORKS
	// 						fmt.Println("NO MATCH FOUND NEW COMBO")
	// 		fmt.Println("match not found", param.Name, param.Value.StringVal)
	// 		len := len(combinations)
	// 		newCombination := createCombination(len-1, param.Name, param.Value.StringVal, []v1beta1.Param{})
	// 		fmt.Println("NEW combination", newCombination)
	// 		combinations = append(combinations, newCombination)
	// 		fmt.Println("combinations", combinations)
	// 		}
	// 	}
	// }

	// Filter out which include params will be replaced
	mappedParams := [][]v1beta1.Param{}
	for _, matrixInclude := range matrix.Include {
		if (len(matrixInclude.Params) > 0) {
			mappedParams = append(mappedParams, matrixInclude.Params )
		}
	}
	fmt.Println("mappedParams", mappedParams)
	for i := 0; i < len(mappedParams); i++ {
		fmt.Println("i", i)
		params := mappedParams[i]
		param := params[0]
		fmt.Println("param", param)
		matchedIds := Filter(combinations, param)
		fmt.Println("matchedIds", matchedIds)

		// if no matched ids, replace all
		if (len(matchedIds) == 0 ){
			fmt.Print("ADDING NEW PARAM TO ALL")
		for i := 0; i < len(combinations); i+=2 {
			combination := combinations[i]
		// for i, combination := range combinations {
				newParam := v1beta1.Param{
					Name:  param.Name,
					Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: param.Value.StringVal},
				}
				combination.Params = append(combination.Params, newParam)
			}
			fmt.Println("COMBOS AFTER REPLACEALL:", combinations)
		}

		// we have to replace the combination at the given id
		for i := 0; i < len(matchedIds); i++ {
			next := params[1]
			fmt.Println("next", next)
			newParam := v1beta1.Param{
				Name:  next.Name,
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: next.Value.StringVal},
			}
			fmt.Println("NEW:", newParam)
			fmt.Println("combination at matchid", combinations[i+1])
			// fmt.Println("mappedCombination at matchid", mappedCombinations[strconv.Itoa(i+1)])
			fmt.Println("BEFORE:", combinations[i+1].Params)
			combinations[i+1].Params = append(combinations[i+1].Params, newParam)
			fmt.Println("AFTER:", combinations[i+1].Params)
		}
	}

			// UNCOMMENT
			// // If it does not exist, append to all combinations params
			// if val, exist := paramNamesMap[param.Name]; !exist {
			// 	// Param name not found in mapping
			// 	// Common param will overwrite all params
			// 	// No new combinations generated
			// 	fmt.Println("param name not found common package", val)
			// 	fmt.Println("Adds new combination tested", param.Name)
			// 	// for i := 0; i < len(combinations); i+=2 {
			// 	// 	combination := combinations[i]
			// 	// // for i, combination := range combinations {
			// 	// newParam := v1beta1.Param{
			// 	// 	Name:  param.Name,
			// 	// 	Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: param.Value.StringVal},
			// 	// }
			// 	// combination.Params = append(combination.Params, newParam)
			// 	for i := 0; i < len(mappedCombinations); i++ {

			// 		}
			// 		mappedCombinations[strconv.Itoa(i)] = append(mappedCombinations[strconv.Itoa(i)], newParam)
			// 	}
			// 	fmt.Println(mappedCombinations)
			// 	break
			// }



			// // Otherwise the param value does not exist
			// // New combination generated

	// }
	// TODO DELETE FMT
	// for i, combinationParam := range mappedCombinations {
	// 	fmt.Println(i)
	// 	fmt.Println(combinationParam)
	// }
	return combinations
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

func Filter(combinations Combinations, seekingParam v1beta1.Param) []string {
	matchingIds := []string{}
	// when there are no existing combinations, this is the first parameter in the matrix, so we initialize the
	// combinations with the first Parameter
	for _, combination := range combinations {
		id := combination.MatrixID
		firstParam := combination.Params[0]
		if (firstParam.Name == seekingParam.Name && firstParam.Value.StringVal == seekingParam.Value.StringVal) {
			fmt.Println("FOUND")
			fmt.Println("id", id)
			fmt.Println("firstParam", firstParam)
			fmt.Println("seekingParam", seekingParam)
			matchingIds = append(matchingIds, id)
		}
	}
	fmt.Println(matchingIds)
	return matchingIds
}
