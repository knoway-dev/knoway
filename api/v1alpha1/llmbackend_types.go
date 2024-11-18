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
)

type UsageStatsFilter struct {
}

type LLMBackendFilter struct {
	Type       string            `json:"type,omitempty"`
	UsageStats *UsageStatsFilter `json:"usageStats,omitempty"`
}

// LLMBackendSpec defines the desired state of LLMBackend
type LLMBackendSpec struct {
	ModelName string             `json:"modelName,omitempty"`
	Provider  string             `json:"provider,omitempty"`
	SecretRef string             `json:"secretRef,omitempty"`
	Filters   []LLMBackendFilter `json:"filters,omitempty"`
}

// LLMBackendStatus defines the observed state of LLMBackend
type LLMBackendStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

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
