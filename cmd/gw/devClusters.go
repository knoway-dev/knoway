package gw

import (
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/anypb"

	v1alpha4 "knoway.dev/api/clusters/v1alpha1"
	v1alpha2 "knoway.dev/api/filters/v1alpha1"
	"knoway.dev/pkg/bootkit"
	cluster2 "knoway.dev/pkg/registry/cluster"
	"knoway.dev/pkg/registry/route"
)

var StaticClustersConfig = map[string]*v1alpha4.Cluster{
	"openai/gpt-3.5-turbo": {
		Name:              "openai/gpt-3.5-turbo",
		Provider:          "openai",
		LoadBalancePolicy: v1alpha4.LoadBalancePolicy_ROUND_ROBIN,
		Upstream: &v1alpha4.Upstream{
			Url:    "https://openrouter.ai/api/v1/chat/completions",
			Method: v1alpha4.Upstream_POST,
			Headers: []*v1alpha4.Upstream_Header{
				{
					Key:   "Authorization",
					Value: "Bearer sk-or-v1-",
				},
			},
		},
		TlsConfig: nil,
		Filters: []*v1alpha4.ClusterFilter{
			{
				Name: "openai-usage-stats",
				Config: func() *anypb.Any {
					return lo.Must(anypb.New(&v1alpha2.UsageStatsConfig{}))
				}(),
			},
			{
				Name: "openai-request-unmarshaller",
				Config: func() *anypb.Any {
					return lo.Must(anypb.New(&v1alpha2.OpenAIRequestMarshallerConfig{}))
				}(),
			},
			{
				Name: "model-name-mapping",
				Config: func() *anypb.Any {
					return lo.Must(anypb.New(&v1alpha2.OpenAIModelNameRewriteConfig{ModelName: "gpt-3.5-turbo"}))
				}(),
			},
			{
				Name: "openai-response-unmarshaller",
				Config: func() *anypb.Any {
					return lo.Must(anypb.New(&v1alpha2.OpenAIResponseUnmarshallerConfig{}))
				}(),
			},
		},
	},
}

func StaticRegisterClusters(clusterDetails map[string]*v1alpha4.Cluster, lifecycle bootkit.LifeCycle) error {
	for _, cluster := range clusterDetails {
		if err := cluster2.UpsertAndRegisterCluster(cluster, lifecycle); err != nil {
			return err
		}
		if err := route.RegisterRouteWithConfig(route.InitDirectModelRoute(cluster.GetName())); err != nil {
			return err
		}
	}

	return nil
}
