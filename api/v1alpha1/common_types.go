package v1alpha1

type Header struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

// HeaderFromSource represents the source of a set of ConfigMaps or Secrets
type HeaderFromSource struct {
	// An optional identifier to prepend to each key in the ref.
	Prefix string `json:"prefix,omitempty"`
	// Type of the source (ConfigMap or Secret)
	RefType ValueFromType `json:"refType,omitempty"`
	// Name of the source
	RefName string `json:"refName,omitempty"`
}

// ValueFromType defines the type of source for headers.
// +kubebuilder:validation:Enum=ConfigMap;Secret
type ValueFromType string

const (
	// ConfigMap indicates that the header source is a ConfigMap.
	ConfigMap ValueFromType = "ConfigMap"
	// Secret indicates that the header source is a Secret.
	Secret ValueFromType = "Secret"
)

// StatusEnum defines the possible statuses for the LLMBackend, ImageGenerationBackend, and other types.
type StatusEnum string

const (
	Unknown StatusEnum = "Unknown"
	Healthy StatusEnum = "Healthy"
	Failed  StatusEnum = "Failed"
)
