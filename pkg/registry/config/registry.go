package config

import (
	"fmt"

	"google.golang.org/protobuf/types/known/anypb"

	filtersv1alpha1 "knoway.dev/api/filters/v1alpha1"
	"knoway.dev/pkg/bootkit"
	clusterfilters "knoway.dev/pkg/clusters/filters"
	"knoway.dev/pkg/clusters/filters/openai"
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/filters/auth"
	"knoway.dev/pkg/filters/usage"
	"knoway.dev/pkg/protoutils"
)

var (
	requestFilters = map[string]func(cfg *anypb.Any, lifecycle bootkit.LifeCycle) (filters.RequestFilter, error){}

	clustersFilters = map[string]func(cfg *anypb.Any, lifecycle bootkit.LifeCycle) (clusterfilters.ClusterFilter, error){}
)

func ClusterDefaultFilters(lifecycle bootkit.LifeCycle) []clusterfilters.ClusterFilter {
	res := make([]clusterfilters.ClusterFilter, 0)

	pb, _ := anypb.New(&filtersv1alpha1.OpenAIRequestHandlerConfig{})
	reqMar, _ := NewClusterFilterWithConfig("global", pb, lifecycle)
	res = append(res, reqMar)

	responsePb, _ := anypb.New(&filtersv1alpha1.OpenAIResponseHandlerConfig{})
	respMar, _ := NewClusterFilterWithConfig("global", responsePb, lifecycle)
	res = append(res, respMar)

	return res
}

func init() {
	requestFilters[protoutils.TypeURLOrDie(&filtersv1alpha1.APIKeyAuthConfig{})] = auth.NewWithConfig
	requestFilters[protoutils.TypeURLOrDie(&filtersv1alpha1.UsageStatsConfig{})] = usage.NewWithConfig

	// internal base Filters
	clustersFilters[protoutils.TypeURLOrDie(&filtersv1alpha1.OpenAIRequestHandlerConfig{})] = openai.NewRequestHandlerWithConfig
	clustersFilters[protoutils.TypeURLOrDie(&filtersv1alpha1.OpenAIResponseHandlerConfig{})] = openai.NewResponseHandlerWithConfig
}

func NewRequestFilterWithConfig(name string, cfg *anypb.Any, lifecycle bootkit.LifeCycle) (filters.RequestFilter, error) {
	if f, ok := requestFilters[cfg.GetTypeUrl()]; ok {
		return f(cfg, lifecycle)
	}

	return nil, fmt.Errorf("unknown listener filter %q, %s", name, cfg.GetTypeUrl())
}

func NewClusterFilterWithConfig(name string, cfg *anypb.Any, lifecycle bootkit.LifeCycle) (clusterfilters.ClusterFilter, error) {
	if f, ok := clustersFilters[cfg.GetTypeUrl()]; ok {
		return f(cfg, lifecycle)
	}

	return nil, fmt.Errorf("unknown cluster filter %q, %s", name, cfg.GetTypeUrl())
}

// NewRequestFiltersKeys returns the keys of the requestFilters map
func NewRequestFiltersKeys() []string {
	keys := make([]string, 0, len(requestFilters))
	for k := range requestFilters {
		keys = append(keys, k)
	}

	return keys
}

// NewClustersFiltersKeys returns the keys of the clustersFilters map
func NewClustersFiltersKeys() []string {
	keys := make([]string, 0, len(clustersFilters))
	for k := range clustersFilters {
		keys = append(keys, k)
	}

	return keys
}
