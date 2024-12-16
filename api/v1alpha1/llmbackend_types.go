/*
Copyright 2024.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Provider",type=string,JSONPath=`.spec.provider`
//+kubebuilder:printcolumn:name="Name",type=string,JSONPath=`.spec.name`
//+kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.spec.upstream.baseUrl`
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`

// LLMBackend is the Schema for the llmbackends API
type LLMBackend struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LLMBackendSpec   `json:"spec,omitempty"`
	Status LLMBackendStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LLMBackendList contains a list of LLMBackend
type LLMBackendList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LLMBackend `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LLMBackend{}, &LLMBackendList{})
}

// LLMBackendSpec defines the desired state of LLMBackend
type LLMBackendSpec struct {
	// ModelName specifies the name of the model
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`
	// Provider indicates the organization providing the model
	// +kubebuilder:validation:Required
	Provider string `json:"provider,omitempty"`
	// Upstream contains information about the upstream configuration
	Upstream BackendUpstream `json:"upstream,omitempty"`
	// Filters are applied to the model's requests
	Filters []LLMBackendFilter `json:"filters,omitempty"`
}

// BackendUpstream defines the upstream server configuration.
type BackendUpstream struct {
	// BaseUrl define upstream endpoint url
	// Example:
	// 		https://openrouter.ai/api/v1/chat/completions
	//
	//  	http://phi3-mini.default.svc.cluster.local:8000/api/v1/chat/completions
	BaseURL string `json:"baseUrl,omitempty"`

	// Headers defines the common headers for the model, such as the authentication header for the API key.
	// Example:
	//
	// headers：
	// 	- key: apikey
	// 	  value: "sk-or-v1-xxxxxxxxxx"
	Headers []Header `json:"headers,omitempty"`
	// Headers defines the common headers for the model, such as the authentication header for the API key.
	// Example:
	//
	// headersFrom：
	// 	- prefix: sk-or-v1-
	//	  refType: Secret
	//	  refName: common-gpt4-apikey
	HeadersFrom []HeaderFromSource `json:"headersFrom,omitempty"`

	DefaultParams  *ModelParams `json:"defaultParams,omitempty"`
	OverrideParams *ModelParams `json:"overrideParams,omitempty"`

	Timeout int32 `json:"timeout,omitempty"`
}

type ModelParams struct {
	// OpenAI model parameters
	OpenAI *OpenAIParam `json:"openai,omitempty"`
}

type CommonParams struct {
	Model string `json:"model,omitempty"`

	// Temperature is the sampling temperature, between 0 and 2.
	// Higher values like 0.8 make the output more random, while lower values like 0.2 make it more focused and deterministic.
	Temperature *string `json:"temperature,omitempty" floatString:"true"`
}

type OpenAIParam struct {
	CommonParams `json:",inline"`

	// MaxTokens is deprecated. Use MaxCompletionTokens instead.
	// This value is not compatible with o1 series models.
	MaxTokens *int `json:"max_tokens,omitempty"`
	// MaxCompletionTokens limits the maximum number of tokens for completion.
	MaxCompletionTokens *int `json:"max_completion_tokens,omitempty"`
	// TopP is the nucleus sampling probability, between 0 and 1.
	TopP *string `json:"top_p,omitempty" floatString:"true"`
	// Stream specifies whether to enable streaming responses.
	Stream *bool `json:"stream,omitempty"`
	// StreamOptions defines additional options for streaming responses.
	StreamOptions *StreamOptions `json:"stream_options,omitempty"`
}

type StreamOptions struct {
	// IncludeUsage indicates whether to include usage statistics before the [DONE] message.
	IncludeUsage *bool `json:"include_usage,omitempty"`
}

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

// LLMBackendFilter represents the backend filter configuration.
type LLMBackendFilter struct {
	Name string `json:"name,omitempty"` // Filter name

	FilterConfig `json:",inline"`
}

// FilterConfig represents the configuration for filters.
// At least one of the following must be specified: UsageStatsConfig, ModelRewriteConfig, or CustomConfig
// +kubebuilder:validation:Required
type FilterConfig struct {
	// UsageStats:  Usage stats configuration
	// +kubebuilder:validation:OneOf
	// +optional
	UsageStats *UsageStatsConfig `json:"usageStats,omitempty"`

	//ModelRewrite: Model rewrite configuration
	// +kubebuilder:validation:OneOf
	// +optional
	ModelRewrite *OpenAIModelNameRewriteConfig `json:"modelRewrite,omitempty"`

	// Custom: Custom plugin configuration
	// Example:
	//
	// 	custom:
	// 		pluginName: examplePlugin
	// 		pluginVersion: "1.0.0"
	// 		settings:
	//   		setting1: value1
	//   		setting2: value2
	//
	// +kubebuilder:validation:OneOf
	// +optional
	Custom *runtime.RawExtension `json:"custom,omitempty"`
}

// UsageStatsConfig defines the configuration for usage statistics.
type UsageStatsConfig struct {
	Address string `json:"address,omitempty"`
}

// OpenAIModelNameRewriteConfig defines the configuration for rewriting OpenAI model names.
type OpenAIModelNameRewriteConfig struct {
	ModelName string `json:"modelName,omitempty"`
}

// LLMBackendStatus defines the observed state of LLMBackend
type LLMBackendStatus struct {
	// Status indicates the health of the backend: Unknown, Healthy, or Failed
	// +kubebuilder:validation:Enum=Unknown;Healthy;Failed
	Status StatusEnum `json:"status,omitempty"`

	// Conditions represent the current conditions of the backend
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Endpoints holds the upstream addresses of the current model (pod IP addresses)
	Endpoints []string `json:"endpoints,omitempty"`
}

// StatusEnum defines the possible statuses for the LLMBackend
type StatusEnum string

const (
	Unknown StatusEnum = "Unknown"
	Healthy StatusEnum = "Healthy"
	Failed  StatusEnum = "Failed"
)
