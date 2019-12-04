package traefik

import (
	"github.com/linkerd/linkerd2/controller/gw-annotator/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Gateway ..
type Gateway struct {
	Object *unstructured.Unstructured
}

// IsAnnotated ..
func (g *Gateway) IsAnnotated() bool {
	return false
}

// GenerateAnnotationPatch ..
func (g *Gateway) GenerateAnnotationPatch() *util.PatchOperation {
	return nil
}
