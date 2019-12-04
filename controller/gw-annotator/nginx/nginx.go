package nginx

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/linkerd/linkerd2/controller/gw-annotator/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	// DefaultPrefix ..
	DefaultPrefix = "nginx"
	// ConfigSnippetKey ..
	ConfigSnippetKey = ".ingress.kubernetes.io/configuration-snippet"
)

var (
	l5dHeadersRE = regexp.MustCompile(`(proxy|grpc)_set_header l5d-dst-override .+`)
)

// Gateway ..
type Gateway struct {
	Object *unstructured.Unstructured
}

// IsAnnotated ..
func (g *Gateway) IsAnnotated() bool {
	_, configSnippet, found := g.getConfigSnippetAnnotation()

	// Check if nginx configuration-snippet annotation exists
	if !found {
		return false
	}

	// Check if nginx configuration-snippet annotation has the l5d header
	if l5dHeadersRE.MatchString(configSnippet) {
		return true
	}

	return false
}

// GenerateAnnotationPatch ..
func (g *Gateway) GenerateAnnotationPatch() *util.PatchOperation {
	annotationKey, configSnippet, found := g.getConfigSnippetAnnotation()

	var op string
	if !found {
		op = "add"
	} else {
		op = "replace"
		if configSnippet != "" && !strings.HasSuffix(configSnippet, "\n") {
			configSnippet += "\n"
		}
	}
	// TODO (tegioz): support custom cluster domain
	configSnippet += "proxy_set_header l5d-dst-override $service_name.$namespace.svc.cluster.local:$service_port;\n"
	configSnippet += "grpc_set_header l5d-dst-override $service_name.$namespace.svc.cluster.local:$service_port;\n"

	return &util.PatchOperation{
		Op:    op,
		Path:  fmt.Sprintf("%s%s", util.AnnotationsPath, annotationKey),
		Value: configSnippet,
	}
}

func (g *Gateway) getConfigSnippetAnnotation() (string, string, bool) {
	for k, v := range g.Object.GetAnnotations() {
		if strings.Contains(k, ConfigSnippetKey) {
			return k, v, true
		}
	}
	// TODO (tegioz): potential issue, nginx annotation prefix is configurable by user
	return DefaultPrefix + ConfigSnippetKey, "", false
}
