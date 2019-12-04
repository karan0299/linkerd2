package util

const (
	// AnnotationsPath ..
	AnnotationsPath = "/metadata/annotations/"
)

// PatchOperation ..
type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}
