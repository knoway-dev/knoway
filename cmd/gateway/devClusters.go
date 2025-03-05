package gateway

import (
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/anypb"

	clusters "knoway.dev/api/clusters/v1alpha1"
	filters "knoway.dev/api/filters/v1alpha1"
	"knoway.dev/pkg/bootkit"
	clustermanager "knoway.dev/pkg/clusters/manager"
	routemanager "knoway.dev/pkg/route/manager"
)

var StaticClustersConfig = map[string]*clusters.Cluster{
	"openai/gpt-3.5-turbo": {
		Type:              clusters.ClusterType_LLM,
		Name:              "openai/gpt-3.5-turbo",
		Provider:          clusters.ClusterProvider_OPEN_AI,
		LoadBalancePolicy: clusters.LoadBalancePolicy_ROUND_ROBIN,
		Upstream: &clusters.Upstream{
			Url: "https://openrouter.ai/api/v1/chat/completions",
			Headers: []*clusters.Upstream_Header{
				{
					Key:   "Authorization",
					Value: "Bearer sk-or-v1-",
				},
			},
		},
		TlsConfig: nil,
		Filters: []*clusters.ClusterFilter{
			{
				Name: "openai-request-handler",
				Config: func() *anypb.Any {
					return lo.Must(anypb.New(&filters.OpenAIRequestHandlerConfig{}))
				}(),
			},
			{
				Name: "openai-response-handler",
				Config: func() *anypb.Any {
					return lo.Must(anypb.New(&filters.OpenAIResponseHandlerConfig{}))
				}(),
			},
		},
	},
}

func StaticRegisterClusters(clusterDetails map[string]*clusters.Cluster, lifecycle bootkit.LifeCycle) error {
	for _, c := range clusterDetails {
		if err := clustermanager.UpsertAndRegisterCluster(c, lifecycle); err != nil {
			return err
		}
		if err := routemanager.RegisterBaseRouteWithConfig(routemanager.InitDirectModelRoute(c.GetName()), lifecycle); err != nil {
			return err
		}
	}

	return nil
}
