package controller

import (
	"github.com/samber/lo"

	"knoway.dev/api/clusters/v1alpha1"
	routev1alpha1 "knoway.dev/api/route/v1alpha1"
	knowaydevv1alpha1 "knoway.dev/api/v1alpha1"
)

var (
	mapClusterProviderBackendProvider = map[v1alpha1.ClusterProvider]knowaydevv1alpha1.Provider{
		v1alpha1.ClusterProvider_OPEN_AI: knowaydevv1alpha1.ProviderOpenAI,
		v1alpha1.ClusterProvider_VLLM:    knowaydevv1alpha1.ProviderVLLM,
		v1alpha1.ClusterProvider_OLLAMA:  knowaydevv1alpha1.ProviderOllama,
	}
	mapBackendProviderClusterProvider = map[knowaydevv1alpha1.Provider]v1alpha1.ClusterProvider{
		knowaydevv1alpha1.ProviderOpenAI: v1alpha1.ClusterProvider_OPEN_AI,
		knowaydevv1alpha1.ProviderVLLM:   v1alpha1.ClusterProvider_VLLM,
		knowaydevv1alpha1.ProviderOllama: v1alpha1.ClusterProvider_OLLAMA,
	}
)

func MapClusterProviderToBackendProvider(provider v1alpha1.ClusterProvider) knowaydevv1alpha1.Provider {
	return mapClusterProviderBackendProvider[provider]
}

func MapBackendProviderToClusterProvider(provider knowaydevv1alpha1.Provider) v1alpha1.ClusterProvider {
	return mapBackendProviderClusterProvider[provider]
}

var (
	mapClusterSizeFromBackendSizeFrom = map[knowaydevv1alpha1.SizeFrom]v1alpha1.ClusterMeteringPolicy_SizeFrom{
		knowaydevv1alpha1.SizeFromInput:    v1alpha1.ClusterMeteringPolicy_SIZE_FROM_INPUT,
		knowaydevv1alpha1.SizeFromOutput:   v1alpha1.ClusterMeteringPolicy_SIZE_FROM_OUTPUT,
		knowaydevv1alpha1.SizeFromGreatest: v1alpha1.ClusterMeteringPolicy_SIZE_FROM_GREATEST,
	}
	mapBackendSizeFromClusterSizeFrom = map[v1alpha1.ClusterMeteringPolicy_SizeFrom]knowaydevv1alpha1.SizeFrom{}
)

func MapClusterSizeFromBackendSizeFrom(sizeFrom *v1alpha1.ClusterMeteringPolicy_SizeFrom) *knowaydevv1alpha1.SizeFrom {
	if sizeFrom == nil {
		return nil
	}

	return lo.ToPtr(mapBackendSizeFromClusterSizeFrom[*sizeFrom])
}

func MapBackendSizeFromClusterSizeFrom(sizeFrom *knowaydevv1alpha1.SizeFrom) *v1alpha1.ClusterMeteringPolicy_SizeFrom {
	if sizeFrom == nil {
		return nil
	}

	return lo.ToPtr(mapClusterSizeFromBackendSizeFrom[*sizeFrom])
}

var (
	mapClusterLoadBalancePolicyBackendLoadBalancePolicy = map[knowaydevv1alpha1.LoadBalancePolicy]routev1alpha1.LoadBalancePolicy{
		knowaydevv1alpha1.LoadBalancePolicyWeightedLeastRequest: routev1alpha1.LoadBalancePolicy_LOAD_BALANCE_POLICY_LEAST_REQUEST,
		knowaydevv1alpha1.LoadBalancePolicyWeightedRoundRobin:   routev1alpha1.LoadBalancePolicy_LOAD_BALANCE_POLICY_ROUND_ROBIN,
	}
	mapBackendLoadBalancePolicyClusterLoadBalancePolicy = map[routev1alpha1.LoadBalancePolicy]knowaydevv1alpha1.LoadBalancePolicy{
		routev1alpha1.LoadBalancePolicy_LOAD_BALANCE_POLICY_LEAST_REQUEST: knowaydevv1alpha1.LoadBalancePolicyWeightedLeastRequest,
		routev1alpha1.LoadBalancePolicy_LOAD_BALANCE_POLICY_ROUND_ROBIN:   knowaydevv1alpha1.LoadBalancePolicyWeightedRoundRobin,
	}
)

func MapModelRouteLoadBalancePolicyRouteLoadBalancePolicy(policy routev1alpha1.LoadBalancePolicy) knowaydevv1alpha1.LoadBalancePolicy {
	return mapBackendLoadBalancePolicyClusterLoadBalancePolicy[policy]
}

func MapModelRouteLoadBalancePolicyModelRouteLoadBalancePolicy(policy knowaydevv1alpha1.LoadBalancePolicy) routev1alpha1.LoadBalancePolicy {
	return mapClusterLoadBalancePolicyBackendLoadBalancePolicy[policy]
}

var (
	mapClusterRateLimitBaseOnBackendRateLimitBaseOn = map[knowaydevv1alpha1.ModelRouteRateLimitBasedOn]routev1alpha1.RateLimitBaseOn{
		knowaydevv1alpha1.ModelRouteRateLimitBasedOnUserID: routev1alpha1.RateLimitBaseOn_USER_ID,
		knowaydevv1alpha1.ModelRouteRateLimitBasedOnAPIKey: routev1alpha1.RateLimitBaseOn_API_KEY,
	}
	mapBackendRateLimitBaseOnClusterRateLimitBaseOn = map[routev1alpha1.RateLimitBaseOn]knowaydevv1alpha1.ModelRouteRateLimitBasedOn{
		routev1alpha1.RateLimitBaseOn_USER_ID: knowaydevv1alpha1.ModelRouteRateLimitBasedOnUserID,
		routev1alpha1.RateLimitBaseOn_API_KEY: knowaydevv1alpha1.ModelRouteRateLimitBasedOnAPIKey,
	}
)

func MapModelRouteRateLimitBaseOnRouteRateLimitBaseOn(baseOn routev1alpha1.RateLimitBaseOn) knowaydevv1alpha1.ModelRouteRateLimitBasedOn {
	return mapBackendRateLimitBaseOnClusterRateLimitBaseOn[baseOn]
}

func MapModelRouteRateLimitBaseOnModelRouteRateLimitBaseOn(baseOn knowaydevv1alpha1.ModelRouteRateLimitBasedOn) routev1alpha1.RateLimitBaseOn {
	return mapClusterRateLimitBaseOnBackendRateLimitBaseOn[baseOn]
}
