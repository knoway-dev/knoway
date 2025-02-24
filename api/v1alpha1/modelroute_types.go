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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Model Name",type=string,JSONPath=`.spec.modelName`
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`

type ModelRouteRateLimitBasedOn string

const (
	// ModelRouteRateLimitBasedOnAPIKey indicates rate limiting based on API key
	ModelRouteRateLimitBasedOnAPIKey ModelRouteRateLimitBasedOn = "APIKey"
	// ModelRouteRateLimitBasedOnUser indicates rate limiting based on user identity
	ModelRouteRateLimitBasedOnUser ModelRouteRateLimitBasedOn = "User"
)

type ModelRouteRateLimitAdvanceLimitObject struct {
	// BaseOn specifies what the rate limit is based on
	BaseOn ModelRouteRateLimitBasedOn `json:"baseOn"`
	// Value specifies the value to match
	Value string `json:"value"`
}

type ModelRouteRateLimitAdvanceLimit struct {
	// Objects specifies the objects to match for this advance limit
	Objects []ModelRouteRateLimitAdvanceLimitObject `json:"objects"`
	// Number of requests allowed in the duration window
	// If set to 0, rate limiting will be disabled
	Limit int `json:"limit,omitempty"`
	// Default duration is 300 seconds, with the unit being seconds
	Duration time.Duration `json:"duration,omitempty"`
}

// ModelRouteRateLimit provides rate limiting rules that allow more granular control
// over rate limiting based on combinations of API keys and users.
//
// Example:
// ```yaml
// rateLimit:
//
//	basedOn: "APIKey"      # Base rate limit applies to all API keys
//	limit: 10              # Allow 10 requests
//	duration: 300          # 300s
//	advanceLimits:         # Advanced rules for specific cases
//	  - objects:
//	      - baseOn: "User"     # Match specific user
//	        value: "user-123"
//	      - baseOn: "APIKey"   # And specific API key
//	        value: "api-key-abc"
//	    limit: 50
//	    duration: 1m
//
// ```
type ModelRouteRateLimit struct {
	BasedOn ModelRouteRateLimitBasedOn `json:"basedOn,omitempty"`
	// Number of requests allowed in the duration window
	// If set to 0, rate limiting will be disabled
	Limit int `json:"limit,omitempty"`
	// Default duration is 300 seconds, with the unit being seconds
	Duration int64 `json:"duration,omitempty"`
	// Advanced rate limiting rules
	// +optional
	AdvanceLimits []ModelRouteRateLimitAdvanceLimit `json:"advanceLimits,omitempty"`
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

// ModelRouteSpec defines the desired state of ModelRoute.
type ModelRouteSpec struct {
	ModelName string `json:"modelName"`
	// Rate limit policy
	// +kubebuilder:validation:Optional
	// +optional
	RateLimit *ModelRouteRateLimit `json:"rateLimit"`
	// Route policy
	// +kubebuilder:validation:Optional
	// +optional
	Route *ModelRouteRoute `json:"route"`
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
