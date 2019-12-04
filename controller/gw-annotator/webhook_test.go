package gwannotator

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/linkerd/linkerd2/controller/gw-annotator/nginx"
	"github.com/linkerd/linkerd2/controller/gw-annotator/util"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestAnnotateGateway(t *testing.T) {
	testCases := []struct {
		desc           string
		objectYAML     []byte
		expectedOutput *admissionv1beta1.AdmissionResponse
		expectedError  bool
	}{
		// Errors
		{
			desc:       "invalid ingress yaml",
			objectYAML: []byte(`invalid yaml data`),
			expectedOutput: buildTestAdmissionResponse(&util.PatchOperation{
				Op:    "add",
				Path:  fmt.Sprintf("%s%s", util.AnnotationsPath, nginx.DefaultPrefix+nginx.ConfigSnippetKey),
				Value: fmt.Sprintf("%s\n%s\n", nginx.L5dHeaderHTTP, nginx.L5dHeaderGRPC),
			}),
			expectedError: true,
		},
		// Unknown gateways
		{
			desc: "unknown ingress class",
			objectYAML: []byte(`
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: unknown`,
			),
			expectedOutput: buildTestAdmissionResponse(nil),
			expectedError:  false,
		},
		// Nginx
		{
			desc: "nginx ingress (extensions/v1beta1) not annotated for l5d",
			objectYAML: []byte(`
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: nginx`,
			),
			expectedOutput: buildTestAdmissionResponse(&util.PatchOperation{
				Op:    "add",
				Path:  fmt.Sprintf("%s%s", util.AnnotationsPath, nginx.DefaultPrefix+nginx.ConfigSnippetKey),
				Value: fmt.Sprintf("%s\n%s\n", nginx.L5dHeaderHTTP, nginx.L5dHeaderGRPC),
			}),
			expectedError: false,
		},
		{
			desc: "nginx ingress (networking.k8s.io/v1beta1) not annotated for l5d",
			objectYAML: []byte(`
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: nginx`,
			),
			expectedOutput: buildTestAdmissionResponse(&util.PatchOperation{
				Op:    "add",
				Path:  fmt.Sprintf("%s%s", util.AnnotationsPath, nginx.DefaultPrefix+nginx.ConfigSnippetKey),
				Value: fmt.Sprintf("%s\n%s\n", nginx.L5dHeaderHTTP, nginx.L5dHeaderGRPC),
			}),
			expectedError: false,
		},
		{
			desc: "nginx ingress with configuration-snippet not annotated for l5d",
			objectYAML: []byte(`
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/configuration-snippet: |
      entry1;
      entry2;`,
			),
			expectedOutput: buildTestAdmissionResponse(&util.PatchOperation{
				Op:    "replace",
				Path:  fmt.Sprintf("%s%s", util.AnnotationsPath, nginx.DefaultPrefix+nginx.ConfigSnippetKey),
				Value: fmt.Sprintf("entry1;\nentry2;\n%s\n%s\n", nginx.L5dHeaderHTTP, nginx.L5dHeaderGRPC),
			}),
			expectedError: false,
		},
		{
			desc: "nginx ingress with configuration-snippet already annotated for l5d",
			objectYAML: []byte(`
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/configuration-snippet: |
      proxy_set_header l5d-dst-override $service_name.$namespace.svc.cluster.local:$service_port;`,
			),
			expectedOutput: buildTestAdmissionResponse(nil),
			expectedError:  false,
		},
	}

	recorder := &mockEventRecorder{}
	for i, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("test_%d: %s", i, tc.desc), func(t *testing.T) {
			admissionRequest := &admissionv1beta1.AdmissionRequest{
				Object: runtime.RawExtension{
					Raw: tc.objectYAML,
				},
			}
			output, err := AnnotateGateway(nil, admissionRequest, recorder)
			if err != nil {
				if !tc.expectedError {
					t.Errorf("not expecting error but got %v", err)
				}
			} else {
				if tc.expectedError {
					t.Error("expecting error but got none")
				} else {
					if !reflect.DeepEqual(output, tc.expectedOutput) {
						t.Errorf("expecting output to be\n %v \n(patch: %s) \n but got\n %v \n(patch: %s)",
							tc.expectedOutput, tc.expectedOutput.Patch, output, output.Patch)
					}
				}
			}
		})
	}
}

func buildTestAdmissionResponse(patch *util.PatchOperation) *admissionv1beta1.AdmissionResponse {
	admissionResponse := &admissionv1beta1.AdmissionResponse{
		Allowed: true,
	}
	if patch != nil {
		patchJSON, _ := json.Marshal(patch)
		patchType := admissionv1beta1.PatchTypeJSONPatch
		admissionResponse.PatchType = &patchType
		admissionResponse.Patch = patchJSON
	}
	return admissionResponse
}

type mockEventRecorder struct{}

func (r *mockEventRecorder) Event(object runtime.Object, eventtype, reason, message string) {
}
func (r *mockEventRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}) {
}
func (r *mockEventRecorder) PastEventf(object runtime.Object, timestamp metav1.Time, eventtype, reason, messageFmt string, args ...interface{}) {
}
func (r *mockEventRecorder) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{}) {
}
