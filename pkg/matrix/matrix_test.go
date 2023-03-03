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
	"encoding/json"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

func Test_FanOut(t *testing.T) {
	tests := []struct {
		name             string
		matrix           v1beta1.Matrix
		wantCombinations Combinations
	}{{
		name: "matrix with no params",
		matrix: v1beta1.Matrix{
			Params: []v1beta1.Param{},
		},
		wantCombinations: nil,
	}, {
		name: "single array in matrix",
		matrix: v1beta1.Matrix{
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"linux", "mac", "windows"}},
			}},
		},
		wantCombinations: Combinations{{
			MatrixID: "0",
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux"},
			}},
		}, {
			MatrixID: "1",
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "mac"},
			}},
		}, {
			MatrixID: "2",
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "windows"},
			}},
		}},
	}, {
		name: "multiple arrays in matrix",
		matrix: v1beta1.Matrix{
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"linux", "mac", "windows"}},
			}, {
				Name:  "browser",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"chrome", "safari", "firefox"}},
			}}},
		wantCombinations: Combinations{{
			MatrixID: "0",
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux"},
			}, {
				Name:  "browser",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "chrome"},
			}},
		}, {
			MatrixID: "1",
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "mac"},
			}, {
				Name:  "browser",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "chrome"},
			}},
		}, {
			MatrixID: "2",
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "windows"},
			}, {
				Name:  "browser",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "chrome"},
			}},
		}, {
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux"},
			}, {
				Name:  "browser",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "safari"},
			}},
		}, {
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "mac"},
			}, {
				Name:  "browser",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "safari"},
			}},
		}, {
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "windows"},
			}, {
				Name:  "browser",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "safari"},
			}},
		}, {
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux"},
			}, {
				Name:  "browser",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "firefox"},
			}},
		}, {
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "mac"},
			}, {
				Name:  "browser",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "firefox"},
			}},
		}, {
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "windows"},
			}, {
				Name:  "browser",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "firefox"},
			}},
		}},

		// wantCombinations: Combinations{{
		// 	MatrixID: "0",
		// 	Params: []v1beta1.Param{{
		// 		Name:  "platform",
		// 		Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux"},
		// 	}, {
		// 		Name:  "browser",
		// 		Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "chrome"},
		// 	}},
		// }, {
		// 	name: "multiple arrays in matrix",
		// 	matrix: &v1beta1.Matrix{
		// 		Params: []v1beta1.Param{{
		// 			Name:  "platform",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"linux", "mac", "windows"}},
		// 		}, {
		// 			Name:  "browser",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"chrome", "safari", "firefox"}},
		// 		}}},
		// 	wantCombinations: Combinations{{
		// 		MatrixID: "0",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "platform",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux"},
		// 		}, {
		// 			Name:  "browser",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "chrome"},
		// 		}},
		// 	}, {
		// 		MatrixID: "1",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "platform",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "mac"},
		// 		}, {
		// 			Name:  "browser",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "chrome"},
		// 		}},
		// 	}, {
		// 		MatrixID: "2",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "platform",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "windows"},
		// 		}, {
		// 			Name:  "browser",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "chrome"},
		// 		}},
		// 	}, {
		// 		MatrixID: "3",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "platform",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux"},
		// 		}, {
		// 			Name:  "browser",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "safari"},
		// 		}},
		// 	}, {
		// 		MatrixID: "4",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "platform",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "mac"},
		// 		}, {
		// 			Name:  "browser",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "safari"},
		// 		}},
		// 	}, {
		// 		MatrixID: "5",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "platform",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "windows"},
		// 		}, {
		// 			Name:  "browser",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "safari"},
		// 		}},
		// 	}, {
		// 		MatrixID: "6",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "platform",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux"},
		// 		}, {
		// 			Name:  "browser",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "firefox"},
		// 		}},
		// 	}, {
		// 		MatrixID: "7",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "platform",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "mac"},
		// 		}, {
		// 			Name:  "browser",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "firefox"},
		// 		}},
		// 	}, {
		// 		MatrixID: "8",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "platform",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "windows"},
		// 		}, {
		// 			Name:  "browser",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "firefox"},
		// 		}},
		// 	}},
	}, {
		name: "explicit combinations in matrix",
		matrix: v1beta1.Matrix{
			Include: []v1beta1.MatrixInclude{{
				Name: "build-1"}, {
				Params: []v1beta1.Param{{
					Name: "IMAGE", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "image-1"},
				}, {
					Name: "DOCKERFILE", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/Dockerfile1"}}},
			}, {
				Name: "build-2",
			}, {
				Params: []v1beta1.Param{{
					Name: "IMAGE", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "image-2"},
				}, {
					Name: "DOCKERFILE", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/Dockerfile2"}}},
			}, {
				Name: "build-3",
			}, {
				Params: []v1beta1.Param{{
					Name: "IMAGE", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "image-3"},
				}, {
					Name: "DOCKERFILE", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/Dockerfile3"}}},
			}},
		},
		wantCombinations: Combinations{
			{
				MatrixID: "0",
				Params: []v1beta1.Param{{
					Name: "IMAGE", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "image-1"},
				}, {
					Name: "DOCKERFILE", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/Dockerfile1"}}},
			}, {
				MatrixID: "1",
				Params: []v1beta1.Param{{
					Name: "IMAGE", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "image-2"},
				}, {
					Name: "DOCKERFILE", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/Dockerfile2"}}},
			}, {
				MatrixID: "2",
				Params: []v1beta1.Param{{
					Name: "IMAGE", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "image-3"},
				}, {
					Name: "DOCKERFILE", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/Dockerfile3"}}},
			},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCombinations := FanOut(*tt.matrix)

			gotEncodedCombinations, err := json.Marshal(gotCombinations)
			if err != nil {
				t.Error(err)
			}

			wantEncodedCombinations, err := json.Marshal(tt.wantCombinations)
			if err != nil {
				t.Error(err)
			}

			sort.Slice(wantEncodedCombinations, func(i int, j int) bool { return wantEncodedCombinations[i] < wantEncodedCombinations[j] })
			sort.Slice(gotEncodedCombinations, func(i int, j int) bool { return gotEncodedCombinations[i] < gotEncodedCombinations[j] })

			if d := cmp.Diff(wantEncodedCombinations, gotEncodedCombinations); d != "" {
				t.Errorf("Combinations of Parameters did not match the expected Combinations: %s", d)
			}
		})
	}
}

func Test_FanOut_COMMONPACKAGE(t *testing.T) {
	tests := []struct {
		name             string
		matrix           *v1beta1.Matrix
		wantCombinations Combinations
	}{{
		name: "TEST COMMON PACKAGE",
		matrix: &v1beta1.Matrix{
			Params: []v1beta1.Param{{
				Name: "GOARCH", Value: v1beta1.ParamValue{ArrayVal: []string{"linux/amd64", "linux/ppc64le", "linux/s390x"}},
			}, {
				Name: "version", Value: v1beta1.ParamValue{ArrayVal: []string{"go1.17", "go1.18.1"}}},
			},
			Include: []v1beta1.MatrixInclude{{
				Name: "common-package",
				Params: []v1beta1.Param{{
					Name: "package", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"}}},
			}},
		},
		wantCombinations: Combinations{{
			MatrixID: "1",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/ppc64le"},
			}, {
				Name:  "version",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"},
			}, {
				Name:  "package",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"},
			}},
		}, {
			MatrixID: "0",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/amd64"},
			}, {
				Name:  "version",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"},
			}, {
				Name:  "package",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"},
			}},
		}, {
			MatrixID: "2",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/s390x"},
			}, {
				Name:  "version",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"},
			}, {
				Name:  "package",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"},
			}},
		}, {
			MatrixID: "3",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/s390"},
			}, {
				Name:  "version",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.18.1"},
			}, {
				Name:  "package",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"},
			}},
		}, {
			MatrixID: "4",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/ppc64le"},
			}, {
				Name:  "version",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.18.1"},
			}, {
				Name:  "package",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"},
			}},
		}, {
			MatrixID: "5",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/amd64x"},
			}, {
				Name:  "version",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.18.1"},
			}, {
				Name:  "package",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"},
			}},
		}},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCombinations := FanOut(*tt.matrix)

			gotEncodedCombinations, err := json.Marshal(gotCombinations)
			if err != nil {
				t.Error(err)
			}

			wantEncodedCombinations, err := json.Marshal(tt.wantCombinations)
			if err != nil {
				t.Error(err)
			}

			sort.Slice(wantEncodedCombinations, func(i int, j int) bool { return wantEncodedCombinations[i] < wantEncodedCombinations[j] })
			sort.Slice(gotEncodedCombinations, func(i int, j int) bool { return gotEncodedCombinations[i] < gotEncodedCombinations[j] })

			if d := cmp.Diff(wantEncodedCombinations, gotEncodedCombinations); d != "" {
				t.Errorf("Combinations of Parameters did not match the expected Combinations: %s", d)
			}
		})
	}
}

func Test_FanOut_NEWCOMBO(t *testing.T) {
	tests := []struct {
		name             string
		matrix           *v1beta1.Matrix
		wantCombinations Combinations
	}{{
		name: "NEW COMBO",
		matrix: &v1beta1.Matrix{
			Params: []v1beta1.Param{{
				Name: "GOARCH", Value: v1beta1.ParamValue{ArrayVal: []string{"linux/amd64", "linux/ppc64le", "linux/s390x"}},
			}, {
				Name: "version", Value: v1beta1.ParamValue{ArrayVal: []string{"go1.17", "go1.18.1"}}},
			},
			Include: []v1beta1.MatrixInclude{{
				Name: "non-existent-arch",
				Params: []v1beta1.Param{{
					Name: "GOARCH", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "I-do-not-exist"}},
				}},
			}},
		wantCombinations: Combinations{{
			MatrixID: "0",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/amd64"},
			}, {
				Name:  "version",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"},
			}},
		}, {
			MatrixID: "1",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/ppc64le"},
			}, {
				Name:  "version",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"},
			}},
		}, {
			MatrixID: "2",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/s390x"},
			}, {
				Name:  "version",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"},
			}},
		}, {
			MatrixID: "3",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/amd64"},
			}, {
				Name:  "version",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.18.1"},
			}},
		}, {
			MatrixID: "4",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/ppc64le"},
			}, {
				Name:  "version",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.18.1"},
			}},
		}, {
			MatrixID: "5",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/s390x"},
			}, {
				Name:  "version",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.18.1"},
			}},
		}, {
			MatrixID: "6",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "I-do-not-exist"},
			}},
		}},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCombinations := FanOut(*tt.matrix)

			gotEncodedCombinations, err := json.Marshal(gotCombinations)
			if err != nil {
				t.Error(err)
			}

			wantEncodedCombinations, err := json.Marshal(tt.wantCombinations)
			if err != nil {
				t.Error(err)
			}

			sort.Slice(wantEncodedCombinations, func(i int, j int) bool { return wantEncodedCombinations[i] < wantEncodedCombinations[j] })
			sort.Slice(gotEncodedCombinations, func(i int, j int) bool { return gotEncodedCombinations[i] < gotEncodedCombinations[j] })

			if d := cmp.Diff(wantEncodedCombinations, gotEncodedCombinations); d != "" {
				t.Errorf("Combinations of Parameters did not match the expected Combinations: %s", d)
			}
		})
	}
}

//	Matrix: &Matrix{
//		Params: []Param{{
//			Name: "GOARCH", Value: v1beta1.ParamValue{ArrayVal: []string{"linux/amd64", "linux/ppc64le", "linux/s390x"}},
//		}, {
//			Name: "version", Value: v1beta1.ParamValue{ArrayVal: []string{"go1.17", "go1.18.1"}}},
//		},
//		Include: []MatrixInclude{{
//			Name: "s390x-no-race",
//			Params: []Param{{
//				Name: "GOARCH", Value: v1beta1.ParamValue{Type: ParamTypeString, StringVal: "linux/s390x"},
//			}, {
//				Name: "flags", Value: v1beta1.ParamValue{Type: ParamTypeString, StringVal: "-cover -v"},
//			}, {
//				Name: "version", Value: v1beta1.ParamValue{Type: ParamTypeString, StringVal: "go1.17"}}},
//		}},
//	}},
func Test_FanOut_FILTER3(t *testing.T) {
	tests := []struct {
		name             string
		matrix           *v1beta1.Matrix
		wantCombinations Combinations
	}{{
		name: "FILTER 3",
		matrix: &v1beta1.Matrix{
			Params: []v1beta1.Param{{
				Name: "GOARCH", Value: v1beta1.ParamValue{ArrayVal: []string{"linux/amd64", "linux/ppc64le", "linux/s390x"}},
			}, {
				Name: "version", Value: v1beta1.ParamValue{ArrayVal: []string{"go1.17", "go1.18.1"}}},
			},
			Include: []v1beta1.MatrixInclude{{
				Name: "390x-no-race",
				Params: []v1beta1.Param{{
					Name: "GOARCH", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/s390x"}}, {
					Name: "flags", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "-cover -v"}}, {
					Name: "version", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"}}},
			},
				{
					Name: "amd64-no-race",
					Params: []v1beta1.Param{{
						Name: "GOARCH", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/amd64"}}, {
						Name: "flags", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "-cover -v"}}, {
						Name: "version", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"}}},
				},
			},
		},
		wantCombinations: Combinations{{
			MatrixID: "0",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/amd64"},
			}, {
				Name:  "version",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"},
			}},
		}, {
			MatrixID: "1",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/ppc64le"},
			}, {
				Name:  "version",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"},
			}, {
				Name:  "flags",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "-cover -v"},
			}},
		}, {
			MatrixID: "2",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/s390x"},
			}, {
				Name:  "version",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"},
			}, {
				Name:  "flags",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "-cover -v"},
			}},
		}, {
			MatrixID: "3",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/amd64"},
			}, {
				Name:  "version",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.18.1"},
			}},
		}, {
			MatrixID: "4",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/ppc64le"},
			}, {
				Name:  "version",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.18.1"},
			}},
		}, {
			MatrixID: "5",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/s390x"},
			}, {
				Name:  "version",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.18.1"},
			}},
		}},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCombinations := FanOut(*tt.matrix)

			gotEncodedCombinations, err := json.Marshal(gotCombinations)
			if err != nil {
				t.Error(err)
			}

			wantEncodedCombinations, err := json.Marshal(tt.wantCombinations)
			if err != nil {
				t.Error(err)
			}

			sort.Slice(wantEncodedCombinations, func(i int, j int) bool { return wantEncodedCombinations[i] < wantEncodedCombinations[j] })
			sort.Slice(gotEncodedCombinations, func(i int, j int) bool { return gotEncodedCombinations[i] < gotEncodedCombinations[j] })

			if d := cmp.Diff(wantEncodedCombinations, gotEncodedCombinations); d != "" {
				t.Errorf("Combinations of Parameters did not match the expected Combinations: %s", d)
			}
		})
	}
}

func Test_FanOut_FILTER_TEST_ALL(t *testing.T) {
	tests := []struct {
		name             string
		matrix           *v1beta1.Matrix
		wantCombinations Combinations
	}{{
		name: "ORIGINAL COMBINATION ALL 3",
		matrix: &v1beta1.Matrix{
			Params: []v1beta1.Param{{
				Name: "GOARCH", Value: v1beta1.ParamValue{ArrayVal: []string{"linux/amd64", "linux/ppc64le", "linux/s390x"}},
			}, {
				Name: "version", Value: v1beta1.ParamValue{ArrayVal: []string{"go1.17", "go1.18.1"}}},
			},
			Include: []v1beta1.MatrixInclude{{
				Name: "common-package",
				Params: []v1beta1.Param{{
					Name: "package", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"}}},
			}, {
				Name: "s390x-no-race",
				Params: []v1beta1.Param{{
					Name: "GOARCH", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/s390x"},
				}, {
					Name: "flags", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "-cover -v"}}},
			}, {
				Name: "go117-context",
				Params: []v1beta1.Param{{
					Name: "version", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"},
				}, {
					Name: "context", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/go117/context"}}},
			}, {
				Name: "non-existent-arch",
				Params: []v1beta1.Param{{
					Name: "GOARCH", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "I-do-not-exist"}},
				},
			}},
		},
		// { "GOARCH": "linux/amd64", "version": "go1.17", "package": "path/to/common/package/", "context": "path/to/go117/context" }

		// { "GOARCH": "linux/amd64", "version": "go1.18.1", "package": "path/to/common/package/" }

		// { "GOARCH": "linux/ppc64le", "version": "go1.17", "package": "path/to/common/package/", "context": "path/to/go117/context" }

		// { "GOARCH": "linux/ppc64le", "version": "go1.18.1", "package": "path/to/common/package/" }

		// { "GOARCH": "linux/s390x", "version": "go1.17", "package": "path/to/common/package/", "flags": "-cover -v", "context": "path/to/go117/context" }

		// { "GOARCH": "linux/s390x", "version": "go1.18.1", "package": "path/to/common/package/", "flags": "-cover -v" }

		// { "GOARCH": "I-do-not-exist" }
		wantCombinations: Combinations{{
			MatrixID: "0",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/amd64"},
			}, {
				Name:  "version",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"},
			}, {
				Name:  "package",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"},
			}, {
				Name:  "context",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/go117/context"},
			}},
		}, {
			MatrixID: "1",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/amd64"},
			}, {
				Name:  "version",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.18.1"},
			}, {
				Name:  "package",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"},
			}},
		}, {
			MatrixID: "2",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/ppc64le"},
			}, {
				Name:  "version",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"},
			}, {
				Name:  "package",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"},
			}, {
				Name:  "context",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/go117/context"},
			}},
		}, {
			MatrixID: "3",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/ppc64le"},
			}, {
				Name:  "version",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.18.1"},
			}, {
				Name:  "package",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"},
			}},
		}, {
			MatrixID: "4",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/s390x"},
			}, {
				Name:  "version",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"},
			}, {
				Name:  "package",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"},
			}, {
				Name:  "flags",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "-cover -v"},
			}, {
				Name:  "context",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/go117/context"},
			}},
		}, {
			MatrixID: "5",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/s390x"},
			}, {
				Name:  "version",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.18.1"},
			}, {
				Name:  "package",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"},
			}, {
				Name:  "flags",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "-cover -v"},
			}}}, {
			MatrixID: "6",
			Params: []v1beta1.Param{{
				Name:  "GOARCH",
				Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "I-do-not-exist"},
			}},
		}},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCombinations := FanOut(*tt.matrix)

			gotEncodedCombinations, err := json.Marshal(gotCombinations)
			if err != nil {
				t.Error(err)
			}

			wantEncodedCombinations, err := json.Marshal(tt.wantCombinations)
			if err != nil {
				t.Error(err)
			}

			sort.Slice(wantEncodedCombinations, func(i int, j int) bool { return wantEncodedCombinations[i] < wantEncodedCombinations[j] })
			sort.Slice(gotEncodedCombinations, func(i int, j int) bool { return gotEncodedCombinations[i] < gotEncodedCombinations[j] })

			if d := cmp.Diff(wantEncodedCombinations, gotEncodedCombinations); d != "" {
				t.Errorf("Combinations of Parameters did not match the expected Combinations: %s", d)
			}
		})
	}
}
