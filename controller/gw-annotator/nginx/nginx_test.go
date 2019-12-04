package nginx

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/linkerd/linkerd2/controller/gw-annotator/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestIsAnnotated(t *testing.T) {
	testCases := []struct {
		desc           string
		annotations    map[string]interface{}
		expectedOutput bool
	}{
		{
			desc:           "no annotations",
			annotations:    nil,
			expectedOutput: false,
		},
		{
			desc: "empty nginx configuration snippet annotation",
			annotations: map[string]interface{}{
				DefaultPrefix + ConfigSnippetKey: "",
			},
			expectedOutput: false,
		},
		{
			desc: "nginx configuration snippet annotation present but no l5d header",
			annotations: map[string]interface{}{
				DefaultPrefix + ConfigSnippetKey: "entry1",
			},
			expectedOutput: false,
		},
		{
			desc: "invalid l5d header for http traffic",
			annotations: map[string]interface{}{
				DefaultPrefix + ConfigSnippetKey: "proxy_set_header l5d-dst-override",
			},
			expectedOutput: false,
		},
		{
			desc: "another invalid l5d header for http traffic",
			annotations: map[string]interface{}{
				DefaultPrefix + ConfigSnippetKey: "proxy_set_header l5d-dst-overide test",
			},
			expectedOutput: false,
		},
		{
			desc: "valid l5d header for http traffic",
			annotations: map[string]interface{}{
				DefaultPrefix + ConfigSnippetKey: L5dHeaderHTTP,
			},
			expectedOutput: true,
		},
		{
			desc: "valid l5d header for http traffic (not using default annotation prefix)",
			annotations: map[string]interface{}{
				"custom-prefix" + ConfigSnippetKey: L5dHeaderHTTP,
			},
			expectedOutput: true,
		},
		{
			desc: "valid l5d header for grpc traffic",
			annotations: map[string]interface{}{
				DefaultPrefix + ConfigSnippetKey: L5dHeaderGRPC,
			},
			expectedOutput: true,
		},
		{
			desc: "valid l5d header for http and grpc traffic",
			annotations: map[string]interface{}{
				DefaultPrefix + ConfigSnippetKey: L5dHeaderHTTP + "\n" + L5dHeaderGRPC,
			},
			expectedOutput: true,
		},
	}

	for i, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("test_%d: %s", i, tc.desc), func(t *testing.T) {
			obj := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"annotations": tc.annotations,
					},
				},
			}
			g := &Gateway{Object: obj}
			output := g.IsAnnotated()
			if output != tc.expectedOutput {
				t.Errorf("expecting output to be %v but got %v", tc.expectedOutput, output)
			}
		})
	}
}

func TestGenerateAnnotationPatch(t *testing.T) {
	testCases := []struct {
		desc           string
		annotations    map[string]interface{}
		expectedOutput *util.PatchOperation
	}{
		{
			desc:        "no annotations",
			annotations: nil,
			expectedOutput: &util.PatchOperation{
				Op:    "add",
				Path:  fmt.Sprintf("%s%s", util.AnnotationsPath, DefaultPrefix+ConfigSnippetKey),
				Value: fmt.Sprintf("%s\n%s\n", L5dHeaderHTTP, L5dHeaderGRPC),
			},
		},
		{
			desc: "no nginx configuration snippet annotation",
			annotations: map[string]interface{}{
				"k1": "v1",
			},
			expectedOutput: &util.PatchOperation{
				Op:    "add",
				Path:  fmt.Sprintf("%s%s", util.AnnotationsPath, DefaultPrefix+ConfigSnippetKey),
				Value: fmt.Sprintf("%s\n%s\n", L5dHeaderHTTP, L5dHeaderGRPC),
			},
		},
		{
			desc: "empty nginx configuration snippet annotation",
			annotations: map[string]interface{}{
				DefaultPrefix + ConfigSnippetKey: "",
			},
			expectedOutput: &util.PatchOperation{
				Op:    "replace",
				Path:  fmt.Sprintf("%s%s", util.AnnotationsPath, DefaultPrefix+ConfigSnippetKey),
				Value: fmt.Sprintf("%s\n%s\n", L5dHeaderHTTP, L5dHeaderGRPC),
			},
		},
		{
			desc: "nginx configuration snippet annotation has some entries but not l5d ones",
			annotations: map[string]interface{}{
				DefaultPrefix + ConfigSnippetKey: "entry1;\nentry2;",
			},
			expectedOutput: &util.PatchOperation{
				Op:    "replace",
				Path:  fmt.Sprintf("%s%s", util.AnnotationsPath, DefaultPrefix+ConfigSnippetKey),
				Value: fmt.Sprintf("entry1;\nentry2;\n%s\n%s\n", L5dHeaderHTTP, L5dHeaderGRPC),
			},
		},
		{
			desc: "nginx configuration snippet annotation has some entries but not l5d ones (trailing new line)",
			annotations: map[string]interface{}{
				DefaultPrefix + ConfigSnippetKey: "entry1;\nentry2;\n",
			},
			expectedOutput: &util.PatchOperation{
				Op:    "replace",
				Path:  fmt.Sprintf("%s%s", util.AnnotationsPath, DefaultPrefix+ConfigSnippetKey),
				Value: fmt.Sprintf("entry1;\nentry2;\n%s\n%s\n", L5dHeaderHTTP, L5dHeaderGRPC),
			},
		},
	}

	for i, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("test_%d: %s", i, tc.desc), func(t *testing.T) {
			obj := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"annotations": tc.annotations,
					},
				},
			}
			g := &Gateway{Object: obj}
			output := g.GenerateAnnotationPatch()
			if !reflect.DeepEqual(output, tc.expectedOutput) {
				t.Errorf("expecting output to be\n %v\n but got\n %v", tc.expectedOutput, output)
			}
		})
	}
}
