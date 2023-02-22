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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

func Test_FanOut(t *testing.T) {
	tests := []struct {
		name             string
		matrix           *v1beta1.Matrix
		wantCombinations Combinations
	}{
		// 	{
		// 	name: "single array in matrix",
		// 	matrix: &v1beta1.Matrix{
		// 		Params: []v1beta1.Param{{
		// 			Name:  "platform",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"linux", "mac", "windows"}},
		// 		}}},
		// 	wantCombinations: Combinations{{
		// 		MatrixID: "0",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "platform",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux"},
		// 		}},
		// 	}, {
		// 		MatrixID: "1",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "platform",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "mac"},
		// 		}},
		// 	}, {
		// 		MatrixID: "2",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "platform",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "windows"},
		// 		}},
		// 	}}},
		// {
		// 	name: "multiple arrays in matrix",
		// 	matrix: &v1beta1.Matrix{
		// 		Params: []v1beta1.Param{{
		// 			Name:  "platform",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"linux", "mac", "windows"}}}, {
		// 			Name:  "browser",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"chrome", "safari", "firefox"}}},
		// 		},
		// 	},
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
		// },
		//  {
		// 	name: "include params in matrix",
		// 	matrix: &v1beta1.Matrix{
		// 		Include: []v1beta1.MatrixInclude{
		// 			{Name: "build-1"},
		// 			{Params: []v1beta1.Param{
		// 				{Name: "IMAGE", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "image-1"}},
		// 				{Name: "DOCKERFILE", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/Dockerfile1"}}}},
		// 			{Name: "build-2"},
		// 			{Params: []v1beta1.Param{
		// 				{Name: "IMAGE", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "image-2"}},
		// 				{Name: "DOCKERFILE", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/Dockerfile2"}}}},
		// 			{Name: "build-3"},
		// 			{Params: []v1beta1.Param{
		// 				{Name: "IMAGE", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "image-3"}},
		// 				{Name: "DOCKERFILE", Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/Dockerfile3"}}}},
		// 		}},
		// 	wantCombinations: Combinations{{
		// 		MatrixID: "0",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "IMAGE",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "image-1"},
		// 		}, {
		// 			Name:  "DOCKERFILE",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/Dockerfile1"},
		// 		}},
		// 	}, {
		// 		MatrixID: "1",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "IMAGE",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "image-2"},
		// 		}, {
		// 			Name:  "DOCKERFILE",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/Dockerfile2"},
		// 		}},
		// 	}, {
		// 		MatrixID: "2",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "IMAGE",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "image-3"},
		// 		}, {
		// 			Name:  "DOCKERFILE",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/Dockerfile3"},
		// 		}},
		// 	}},
		// },
		{
			name: "include params in matrix multiple filters",
			matrix: &v1beta1.Matrix{
				Params: []v1beta1.Param{
					{Name: "GOARCH", Value: v1beta1.ParamValue{ArrayVal: []string{"linux/amd64", "linux/ppc64le", "linux/s390x"}}},
					{Name: "version", Value: v1beta1.ParamValue{ArrayVal: []string{"go1.17", "go1.18.1"}}},
				},
				Include: []v1beta1.MatrixInclude{
					{Name: "s390x-no-race "},
					{Params: []v1beta1.Param{
						{Name: "GOARCH", Value: v1beta1.ParamValue{StringVal: "linux/s390x"}},
						{Name: "version", Value: v1beta1.ParamValue{StringVal: "go1.17"}},
						{Name: "flags", Value: v1beta1.ParamValue{StringVal: "-cover -v"}},
					}},				{Name: "s390x-no-race "},
					{Params: []v1beta1.Param{
						{Name: "GOARCH", Value: v1beta1.ParamValue{StringVal: "linux/s390x"}},
						{Name: "flags", Value: v1beta1.ParamValue{StringVal: "-cover -v"}},
					}},
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
					Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/amd64"},
				}, {
					Name:  "version",
					Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.18.1"},
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
					Name:  "flags",
					Value: v1beta1.ParamValue{StringVal: "-cover -v"}},
				}}, {
				MatrixID: "3",
				Params: []v1beta1.Param{{
					Name:  "GOARCH",
					Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/ppc64le"},
				}, {
					Name:  "version",
					Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.18.1"},
				}},
			}, {
				MatrixID: "4",
				Params: []v1beta1.Param{{
					Name:  "GOARCH",
					Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/s390x"},
				}, {
					Name:  "version",
					Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"},
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
		},
		//  {
		// 	name: "include params in matrix non existent",
		// 	matrix: &v1beta1.Matrix{
		// 		Params: []v1beta1.Param{
		// 			{Name: "GOARCH", Value: v1beta1.ParamValue{ArrayVal: []string{"linux/amd64", "linux/ppc64le", "linux/s390x"}}},
		// 			{Name: "version", Value: v1beta1.ParamValue{ArrayVal: []string{"go1.17", "go1.18.1"}}},
		// 		},
		// 		Include: []v1beta1.MatrixInclude{
		// 			{Name: "non-existent-arch"},
		// 			{Params: []v1beta1.Param{
		// 				{Name: "GOARCH", Value: v1beta1.ParamValue{StringVal: "I-do-not-exist"}},
		// 			}},
		// 		},
		// 	},
		// 	wantCombinations: Combinations{{
		// 		MatrixID: "0",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "GOARCH",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/amd64"},
		// 		}, {
		// 			Name:  "version",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"},
		// 		}},
		// 	}, {
		// 		MatrixID: "1",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "GOARCH",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/amd64"},
		// 		}, {
		// 			Name:  "version",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.18.1"},
		// 		}},
		// 	}, {
		// 		MatrixID: "2",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "GOARCH",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/ppc64le"},
		// 		}, {
		// 			Name:  "version",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"},
		// 		}},
		// 	}, {
		// 		MatrixID: "3",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "GOARCH",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/ppc64le"},
		// 		}, {
		// 			Name:  "version",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.18.1"},
		// 		}},
		// 	}, {
		// 		MatrixID: "4",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "GOARCH",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/s390x"},
		// 		}, {
		// 			Name:  "version",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"},
		// 		}},
		// 	}, {
		// 		MatrixID: "5",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "GOARCH",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/s390x"},
		// 		}, {
		// 			Name:  "version",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.18.1"},
		// 		}},
		// 	}, {
		// 		MatrixID: "6",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "GOARCH",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "I-do-not-exist"}}},
		// 	}},
		// },
		// {
		// 	name: "include params in matrix common package",
		// 	matrix: &v1beta1.Matrix{
		// 		Params: []v1beta1.Param{
		// 			{Name: "GOARCH", Value: v1beta1.ParamValue{ArrayVal: []string{"linux/amd64", "linux/ppc64le", "linux/s390x"}}},
		// 			{Name: "version", Value: v1beta1.ParamValue{ArrayVal: []string{"go1.17", "go1.18.1"}}},
		// 		},
		// 		Include: []v1beta1.MatrixInclude{
		// 			{Name: "common-package"},
		// 			{Params: []v1beta1.Param{
		// 				{Name: "package", Value: v1beta1.ParamValue{StringVal: "path/to/common/package/"}},
		// 			}},
		// 		},
		// 	},
		// 	wantCombinations: Combinations{{
		// 		MatrixID: "0",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "GOARCH",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/amd64"},
		// 		}, {
		// 			Name:  "version",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"},
		// 		}, {
		// 			Name:  "package",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"},
		// 		}},
		// 	}, {
		// 		MatrixID: "1",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "GOARCH",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/amd64"},
		// 		}, {
		// 			Name:  "version",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.18.1"},
		// 		}, {
		// 			Name:  "package",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"},
		// 		}},
		// 	}, {
		// 		MatrixID: "2",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "GOARCH",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/ppc64le"},
		// 		}, {
		// 			Name:  "version",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"},
		// 		}, {
		// 			Name:  "package",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"},
		// 		},
		// 		}}, {
		// 		MatrixID: "3",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "GOARCH",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/ppc64le"},
		// 		}, {
		// 			Name:  "version",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.18.1"},
		// 		}, {
		// 			Name:  "package",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"},
		// 		}},
		// 	}, {
		// 		MatrixID: "4",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "GOARCH",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/s390x"},
		// 		}, {
		// 			Name:  "version",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"},
		// 		}, {
		// 			Name:  "package",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"},
		// 		}},
		// 	}, {
		// 		MatrixID: "5",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "GOARCH",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/s390x"},
		// 		}, {
		// 			Name:  "version",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.18.1"},
		// 		}, {
		// 			Name:  "package",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"},
		// 		}},
		// 	}},
		// },
		// {
		// 	name: "include params in matrix",
		// 	matrix: &v1beta1.Matrix{
		// 		Params: []v1beta1.Param{
		// 			{Name: "GOARCH", Value: v1beta1.ParamValue{ArrayVal: []string{"linux/amd64", "linux/ppc64le", "linux/s390x"}}},
		// 			{Name: "version", Value: v1beta1.ParamValue{ArrayVal: []string{"go1.17", "go1.18.1"}}},
		// 		},
		// 		Include: []v1beta1.MatrixInclude{
		// 			// {Name: "common-package"},
		// 			// {Params: []v1beta1.Param{
		// 			// 	{Name: "package", Value: v1beta1.ParamValue{StringVal: "path/to/common/package/"}},
		// 			// }},
		// 			{Name: "s390x-no-race "},
		// 			{Params: []v1beta1.Param{
		// 				{Name: "GOARCH", Value: v1beta1.ParamValue{StringVal: "linux/s390x"}},
		// 				{Name: "version", Value: v1beta1.ParamValue{StringVal: "go1.17"}},
		// 				{Name: "flags", Value: v1beta1.ParamValue{StringVal: "-cover -v"}},
		// 			}},
		// 			// {Name: "go117-context"},
		// 			// {Params: []v1beta1.Param{
		// 			// 	{Name: "version", Value: v1beta1.ParamValue{StringVal: "go1.17"}},
		// 			// 	{Name: "context", Value: v1beta1.ParamValue{StringVal: "path/to/go117/context"}},
		// 			// }},
		// 			// {Name: "non-existent-arch"},
		// 			// {Params: []v1beta1.Param{
		// 			// 	{Name: "GOARCH", Value: v1beta1.ParamValue{StringVal: "I-do-not-exist"}},
		// 			// }},
		// 		},
		// 	},
		// 	wantCombinations: Combinations{{
		// 		MatrixID: "0",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "GOARCH",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/amd64"},
		// 		}, {
		// 			Name:  "version",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"},
		// 		}, {
		// 			Name:  "package",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"},
		// 		}, {
		// 			Name:  "context",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/go117/context"},
		// 		}},
		// 	}, {
		// 		MatrixID: "1",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "GOARCH",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/amd64"},
		// 		}, {
		// 			Name:  "version",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.18.1"},
		// 		}, {
		// 			Name:  "package",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"},
		// 		}},
		// 	}, {
		// 		MatrixID: "2",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "GOARCH",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/ppc64le"},
		// 		}, {
		// 			Name:  "version",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"},
		// 		}, {
		// 			Name:  "package",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"},
		// 		}, {
		// 			Name:  "context",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/go117/context"},
		// 		}},
		// 	}, {
		// 		MatrixID: "3",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "GOARCH",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/ppc64le"},
		// 		}, {
		// 			Name:  "version",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.18.1"},
		// 		}, {
		// 			Name:  "package",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"},
		// 		}},
		// 	}, {
		// 		MatrixID: "4",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "GOARCH",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/s390x"},
		// 		}, {
		// 			Name:  "version",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.17"},
		// 		}, {
		// 			Name:  "package",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"},
		// 		}, {
		// 			Name:  "flags",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "-cover -v"},
		// 		}, {
		// 			Name:  "context",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/go117/context"},
		// 		}},
		// 	}, {
		// 		MatrixID: "5",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "GOARCH",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "linux/s390x"},
		// 		}, {
		// 			Name:  "version",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "go1.18.1"},
		// 		}, {
		// 			Name:  "package",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "path/to/common/package/"},
		// 		}, {
		// 			Name:  "flags",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "-cover -v"},
		// 		}},
		// 	}, {
		// 		MatrixID: "6",
		// 		Params: []v1beta1.Param{{
		// 			Name:  "GOARCH",
		// 			Value: v1beta1.ParamValue{Type: v1beta1.ParamTypeString, StringVal: "I-do-not-exist"}}},
		// 	}},
		// },
	}
	// { "GOARCH": "linux/amd64", "version": "go1.17", "package": "path/to/common/package/", "context": "path/to/go117/context" }

	// { "GOARCH": "linux/amd64", "version": "go1.18.1", "package": "path/to/common/package/" }

	// { "GOARCH": "linux/ppc64le", "version": "go1.17", "package": "path/to/common/package/", "context": "path/to/go117/context" }

	// { "GOARCH": "linux/ppc64le", "version": "go1.18.1", "package": "path/to/common/package/" }

	// { "GOARCH": "linux/s390x", "version": "go1.17", "package": "path/to/common/package/", "flags": "-cover -v", "context": "path/to/go117/context" }

	// { "GOARCH": "linux/s390x", "version": "go1.18.1", "package": "path/to/common/package/", "flags": "-cover -v" }

	// { "GOARCH": "I-do-not-exist" }
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCombinations := FanOut(tt.matrix)
			if d := cmp.Diff(tt.wantCombinations, gotCombinations); d != "" {
				t.Errorf("Combinations of Parameters did not match the expected Combinations: %s", d)
			}
		})
	}
}
