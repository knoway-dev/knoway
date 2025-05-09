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
//+kubebuilder:printcolumn:name="Model Name",type=string,JSONPath=`.spec.modelName`
//+kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.spec.upstream.baseUrl`
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`

// ImageGenerationBackend is the Schema for the imagegenerationbackends API.
type ImageGenerationBackend struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ImageGenerationBackendSpec   `json:"spec,omitempty"`
	Status ImageGenerationBackendStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ImageGenerationBackendList contains a list of ImageGenerationBackend.
type ImageGenerationBackendList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ImageGenerationBackend `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ImageGenerationBackend{}, &ImageGenerationBackendList{})
}

// ImageGenerationBackendSpec defines the desired state of ImageGenerationBackend.
type ImageGenerationBackendSpec struct {
	// ModelName specifies the name of the model
	// +kubebuilder:validation:Optional
	// +optional
	ModelName *string `json:"modelName,omitempty"`
	// Provider indicates the organization providing the model
	// +kubebuilder:validation:Enum=OpenAI;vLLM;Ollama
	Provider Provider `json:"provider,omitempty"`
	// Upstream contains information about the upstream configuration
	Upstream ImageGenerationBackendUpstream `json:"upstream,omitempty"`
	// Filters are applied to the model's requests
	Filters []ImageGenerationFilter `json:"filters,omitempty"`
	// MeteringPolicy contains configurations about how to count the usage of the model
	// +kubebuilder:validation:Optional
	// +optional
	MeteringPolicy *ImageGenerationMeteringPolicy `json:"meteringPolicy,omitempty"`
}

// BackendUpstream defines the upstream server configuration.
type ImageGenerationBackendUpstream struct {
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

	DefaultParams   *ImageGenerationModelParams `json:"defaultParams,omitempty"`
	OverrideParams  *ImageGenerationModelParams `json:"overrideParams,omitempty"`
	RemoveParamKeys []string                    `json:"RemoveParamKeys,omitempty"`

	Timeout int32 `json:"timeout,omitempty"`
}

type ImageGenerationModelParams struct {
	// OpenAI model parameters
	OpenAI *OpenAIImageGenerationParam `json:"openai,omitempty"`
}

type ImageGenerationCommonParams struct {
	Model string `json:"model,omitempty"`

	// A text description of the desired image(s).
	Prompt *string `json:"prompt,omitempty"`
}

type OpenAIImageGenerationResponseFormat string

const (
	OpenAIImageGenerationResponseFormatURL     OpenAIImageGenerationResponseFormat = "url"
	OpenAIImageGenerationResponseFormatB64JSON OpenAIImageGenerationResponseFormat = "b64_json"
)

type OpenAIImageGenerationStyle string

const (
	OpenAIImageGenerationStyleVivid   OpenAIImageGenerationStyle = "vivid"
	OpenAIImageGenerationStyleNatural OpenAIImageGenerationStyle = "natural"
)

type OpenAIImageGenerationParam struct {
	ImageGenerationCommonParams `json:",inline"`

	// N specifies the number of images to generate
	N *string `json:"n,omitempty"`
	// Quality specifies the quality of the image that will be generated.
	// hd creates images with finer details and greater consistency across the image.
	// Some of the model doesn't support this parameter.
	Quality *string `json:"quality,omitempty"`
	// ResponseFormat specifies the format in which the generated images are returned.
	// Must be one of url or b64_json.
	// URLs are only valid for 60 minutes after the image has been generated.
	ResponseFormat *OpenAIImageGenerationResponseFormat `json:"response_format,omitempty"`
	// Size specifies the size of the generated images.
	// Must be one of 256x256, 512x512, or 1024x1024 for dall-e-2.
	// Must be one of 1024x1024, 1792x1024, or 1024x1792 for dall-e-3 models.
	Size *string `json:"size,omitempty"`
	// The style of the generated images.
	// Must be one of vivid or natural.
	// Vivid causes the model to lean towards generating hyper-real and dramatic images.
	// Natural causes the model to produce more natural, less hyper-real looking images.
	// This param is only supported for dall-e-3.
	Style *OpenAIImageGenerationStyle `json:"style,omitempty"`
	// A unique identifier representing your end-user, which can help OpenAI to
	// monitor and detect abuse.
	User *string `json:"user,omitempty"`

	// NegativePrompt is a text description of the undesired features of the image(s).
	NegativePrompt *string `json:"negative_prompt,omitempty"`
	// Guidance scale is a number value that controls how much the conditional signal
	// (prompt, negative_prompt, training images, etc.) affects the generation epoch.
	// In Stable Diffusion, 7.5 is generally used.
	// For more information, see: https://sander.ai/2022/05/26/guidance.html
	GuidanceScale *string `json:"guidance_scale,omitempty"`
}

// ImageGenerationFilter represents the image generation backend filter configuration.
type ImageGenerationFilter struct {
	Name string `json:"name,omitempty"` // Filter name

	ImageGenerationFilterFilterConfig `json:",inline"`
}

// ImageGenerationFilterFilterConfig represents the configuration for filters.
// At least one of the following must be specified: CustomConfig
// +kubebuilder:validation:Required
type ImageGenerationFilterFilterConfig struct {
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

// This enum may also be used for LLMBackend to track down token spent
type SizeFrom string

const (
	// For image generation, the size of the generated image is determined by the input parameters.
	//
	// For example, even if the output image is 1024x1024, as long as the input parameter specified
	// 256x256, the size of the generated image will be account as 256x256.
	SizeFromInput SizeFrom = "Input"
	// For image generation, the size of the generated image is determined by the output image.
	// This is done by parsing through the actual generated image file header by using Golang's std
	// library to determine the size of the image.
	//
	// For example, no matter what the input specified, if the output image is 1024x1024, the size of
	// the generated image will be account as 1024x1024.
	SizeFromOutput SizeFrom = "Output"
	// For image generation, the size of the generated image is determined by the greatest size of the
	// input parameters and output image resolution.
	//
	// For example, if the input parameter specified 256x256 and the output image is 1024x1024, the
	// size of the generated image will be account as 1024x1024. On the other hand, if the input
	// parameter specified 1024x1024 and the output image is 256x256, the size of the generated image
	// will be account as 1024x1024.
	SizeFromGreatest SizeFrom = "Greatest"
)

type ImageGenerationMeteringPolicy struct {
	// SizeFromInput indicates whether the size of the generated image is determined by the input parameters.
	//
	// +kubebuilder:validation:Enum=Input;Output;Greatest
	// +kubebuilder:validation:Optional
	// +optional
	SizeFrom *SizeFrom `json:"sizeFrom,omitempty"`
}

// ImageGenerationBackendStatus defines the observed state of ImageGenerationBackend.
type ImageGenerationBackendStatus struct {
	// Status indicates the health of the backend: Unknown, Healthy, or Failed
	// +kubebuilder:validation:Enum=Unknown;Healthy;Failed
	Status StatusEnum `json:"status,omitempty"`

	// Conditions represent the current conditions of the backend
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Endpoints holds the upstream addresses of the current model (pod IP addresses)
	Endpoints []string `json:"endpoints,omitempty"`
}
