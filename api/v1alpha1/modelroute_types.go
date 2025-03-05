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

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Model Name",type=string,JSONPath=`.spec.modelName`
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`

type RateLimitBasedOn string

const (
	// ModelRouteRateLimitBasedOnAPIKey indicates rate limiting based on API key
	ModelRouteRateLimitBasedOnAPIKey RateLimitBasedOn = "APIKey"
	// ModelRouteRateLimitBasedOnUserID indicates rate limiting based on user identity
	ModelRouteRateLimitBasedOnUserID RateLimitBasedOn = "UserID"

	FilterTypeRateLimit string = "RateLimit"
)

type StringMatch struct {
	// Exact match value
	Exact string `json:"exact,omitempty"`
	// Prefix match value
	Prefix string `json:"prefix,omitempty"`
}

type RateLimitRule struct {
	// Match specifies the match criteria for this rate limit
	Match *StringMatch `json:"match,omitempty"`
	// Number of requests allowed in the duration window
	// If set to 0, rate limiting will be disabled
	Limit int `json:"limit,omitempty"`
	// BasedOn specifies what the rate limit is based on
	// +kubebuilder:validation:Enum=APIKey;UserID
	BasedOn RateLimitBasedOn `json:"basedOn,omitempty"`
	// Default duration is 300 seconds, with the unit being seconds
	Duration int64 `json:"duration,omitempty"`
}

// See also:
// Supported load balancers — envoy 1.34.0-dev-e3a97f documentation
// https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/load_balancing/load_balancers#arch-overview-load-balancing-types
type LoadBalancePolicy string

const (
	LoadBalancePolicyWeightedRoundRobin   LoadBalancePolicy = "WeightedRoundRobin"
	LoadBalancePolicyWeightedLeastRequest LoadBalancePolicy = "WeightedLeastRequest"
)

type ModelRouteRouteTargetDestination struct {
	// Namespace of the backend to lookup for
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`
	// Backend that the route target points to
	// +kubebuilder:validation:Required
	Backend string `json:"backend"`
	// Weight of the target, only used in WeightedRoundRobin and WeightedLeastRequest
	// +kubebuilder:validation:Optional
	// +optional
	Weight *int `json:"weight"`
}

type ModelRouteRouteTarget struct {
	// Destination specifies the destination of the route target
	Destination ModelRouteRouteTargetDestination `json:"destination"`
}

type ModelRouteRouteFallback struct {
	// Order specifies the order of the fallback
	// +kubebuilder:validation:Optional
	// +optional
	Order []string `json:"order"`
}

type ModelRouteRoute struct {
	// LoadBalancePolicy specifies the load balancing policy to use
	// +kubebuilder:validation:Enum=WeightedRoundRobin;WeightedLeastRequest
	LoadBalancePolicy LoadBalancePolicy `json:"loadBalancePolicy"`
	// Targets specifies the targets of the route
	// +kubebuilder:validation:Required
	Targets []ModelRouteRouteTarget `json:"targets"`
}

type RateLimitPolicy struct {
	// Rate limit rules
	// +kubebuilder:validation:Optional
	// +optional
	Rules []*RateLimitRule `json:"rules"`
}

type ModelRouteFallback struct {
	// The delay time before the next retry over request, unit: second
	// +kubebuilder:validation:Optional
	// +optional
	PreDelay *int64 `json:"preDelay"`
	// The delay time after the request is retried, unit: second
	// +kubebuilder:validation:Optional
	// +optional
	PostDelay *int64 `json:"postDelay"`
	// The maximum number of retries
	// +kubebuilder:validation:Optional
	// +optional
	MaxRetires *uint64 `json:"maxRetries"`
}

type ModelRouteFilter struct {
	// Filter name
	// +optional
	Name string `json:"name,omitempty"`
	// Filter type
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=RateLimit
	Type string `json:"type,omitempty"`
	// Rate limit Filter, if the type is RateLimit
	// +kubebuilder:validation:Optional
	// +optional
	RateLimit *RateLimitPolicy `json:"rateLimit"`
}

// ModelRouteSpec defines the desired state of ModelRoute.
type ModelRouteSpec struct {
	ModelName string `json:"modelName"`
	// Filters for the route
	// +kubebuilder:validation:Optional
	Filters []ModelRouteFilter `json:"filters,omitempty"`
	// Route policy
	// +kubebuilder:validation:Optional
	// +optional
	Route *ModelRouteRoute `json:"route"`
	// Fallback
	// +kubebuilder:validation:Optional
	// +optional
	Fallback *ModelRouteFallback `json:"fallback"`
}

type ModelRouteStatusTarget struct {
	Namespace string     `json:"namespace"`
	Backend   string     `json:"backend"`
	ModelName string     `json:"modelName"`
	Status    StatusEnum `json:"status"`
}

// ModelRouteStatus defines the observed state of ModelRoute.
type ModelRouteStatus struct {
	// Status indicates the health of the ModelRoute CR: Unknown, Healthy, or Failed
	// +kubebuilder:validation:Enum=Unknown;Healthy;Failed
	Status StatusEnum `json:"status,omitempty"`

	// Conditions represent the current conditions of the backend
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Targets represents the targets of the model route
	Targets []ModelRouteStatusTarget `json:"targets,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ModelRoute is the Schema for the modelroutes API.
type ModelRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModelRouteSpec   `json:"spec,omitempty"`
	Status ModelRouteStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ModelRouteList contains a list of ModelRoute.
type ModelRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ModelRoute `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ModelRoute{}, &ModelRouteList{})
}
