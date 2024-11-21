package config

import (
	"fmt"

	"google.golang.org/protobuf/types/known/anypb"
	v1alpha2 "knoway.dev/api/filters/v1alpha1"
	clusterfilters "knoway.dev/pkg/clusters/filters"
	"knoway.dev/pkg/clusters/filters/openai"
	"knoway.dev/pkg/clusters/filters/stats"
	listenerfilters "knoway.dev/pkg/filters"
	"knoway.dev/pkg/filters/auth"
	"knoway.dev/pkg/protoutils"
)

var (
	requestFilters = map[string]func(cfg *anypb.Any) (listenerfilters.RequestFilter, error){}

	clustersFilters = map[string]func(cfg *anypb.Any) (clusterfilters.ClusterFilter, error){}
)

func init() {
	requestFilters[protoutils.TypeURLOrDie(&v1alpha2.APIKeyAuthConfig{})] = auth.NewWithConfig

	clustersFilters[protoutils.TypeURLOrDie(&v1alpha2.UsageStatsConfig{})] = stats.NewWithConfig
	clustersFilters[protoutils.TypeURLOrDie(&v1alpha2.OpenAIRequestMarshallerConfig{})] = openai.NewRequestMarshallerWithConfig
	clustersFilters[protoutils.TypeURLOrDie(&v1alpha2.OpenAIModelNameRewriteConfig{})] = openai.NewModelNameRewriteWithConfig
	clustersFilters[protoutils.TypeURLOrDie(&v1alpha2.OpenAIResponseUnmarshallerConfig{})] = openai.NewResponseUnmarshallerWithConfig
}

func NewRequestFilterWithConfig(name string, cfg *anypb.Any) (listenerfilters.RequestFilter, error) {
	if f, ok := requestFilters[cfg.GetTypeUrl()]; ok {
		return f(cfg)
	}

	return nil, fmt.Errorf("unknown listener filter %q, %s", name, cfg.GetTypeUrl())
}

func NewClusterFilterWithConfig(name string, cfg *anypb.Any) (clusterfilters.ClusterFilter, error) {
	if f, ok := clustersFilters[cfg.GetTypeUrl()]; ok {
		return f(cfg)
	}

	return nil, fmt.Errorf("unknown cluster filter %q, %s", name, cfg.GetTypeUrl())
}
