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
	ModelName string `json:"modelName,omitempty"`
	// Provider indicates the organization providing the model
	Provider string `json:"provider,omitempty"`
	// Upstream contains information about the upstream configuration
	Upstream BackendUpstream `json:"upstream,omitempty"`
	// Filters are applied to the model's requests
	Filters []LLMBackendFilter `json:"filters,omitempty"`
}

type Server struct {
	Address          string            `json:"address,omitempty"`
	API              string            `json:"api,omitempty"`
	Method           string            `json:"method,omitempty"`
	WorkloadSelector map[string]string `json:"workloadSelector,omitempty"`
}

// BackendUpstream defines the upstream server configuration.
type BackendUpstream struct {
	// Server: Upstream service configuration
	//	server:
	//      api: /api/v1/chat/completions
	//		method: post
	//		workloadSelector:
	//			modelApp: cus-model
	//
	// 	server:
	//      api: /api/v1/chat/completions
	//		method: post
	// 		address: https://openrouter.ai
	Server Server `json:"server,omitempty"`

	// Headers defines the common headers for the model, such as the authentication header for the API key.
	// Example:
	//
	// headers：
	// 	- key: apikey
	// 	  valueFrom:
	// 		prefix: sk-or-v1-
	//		refType: Secret
	//		refName: common-gpt4-apikey
	//
	// headers：
	// 	- key: apikey
	// 	  value: sk-or-v1-xxxxxxxxxx
	Headers []HeaderDefine `json:"headers,omitempty"`

	Timeout int32 `json:"timeout,omitempty"`
}

type HeaderDefine struct {
	Key string `json:"key,omitempty"`

	// +kubebuilder:validation:OneOf
	Value string `json:"value,omitempty"`
	// +kubebuilder:validation:OneOf
	ValueFrom *ValueFrom `json:"valueFrom,omitempty"`
}

type ValueFrom struct {
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
	// todo 参考了 cilium https://github.com/cilium/cilium/blob/e6ace8e3827a581f9d6d68b7a2434d739fafbc99/pkg/policy/api/cidr.go#L36，但是没有生效
	// +optional
	UsageStats *UsageStatsConfig `json:"usageStats.,omitempty"`

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
	Custom runtime.RawExtension `json:"custom,omitempty"`
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
	Conditions []Condition `json:"conditions,omitempty"`

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

// Condition defines the state of a specific condition
type Condition struct {
	Type    string `json:"type,omitempty"`    // Type of the condition
	Message string `json:"message,omitempty"` // Human-readable message indicating details about the condition
	Ready   bool   `json:"ready,omitempty"`   // Indicates if the backend is ready
}
