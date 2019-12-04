package ambassador

import (
	"github.com/linkerd/linkerd2/controller/gw-annotator/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ConfigMode ..
type ConfigMode int

const (
	// CRD ..
	CRD ConfigMode = iota
	// Ingress ..
	Ingress
	// ServiceAnnotation ..
	ServiceAnnotation
)

// Gateway ..
type Gateway struct {
	Object     *unstructured.Unstructured
	ConfigMode ConfigMode
}

// IsAnnotated ..
func (g *Gateway) IsAnnotated() bool {
	return false
}

// GenerateAnnotationPatch ..
func (g *Gateway) GenerateAnnotationPatch() *util.PatchOperation {
	return nil
}
