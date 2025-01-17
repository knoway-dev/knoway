package listener

import (
	"context"

	"github.com/samber/lo"

	"knoway.dev/api/clusters/v1alpha1"
	"knoway.dev/pkg/clusters"
	"knoway.dev/pkg/object"
	registrycluster "knoway.dev/pkg/registry/cluster"
	registryroute "knoway.dev/pkg/registry/route"
	"knoway.dev/pkg/route"
)

func FindRoute(ctx context.Context, llmRequest object.LLMRequest) (route.Route, string) {
	var r route.Route
	var clusterName string

	// TODO: do route
	registryroute.ForeachRoute(func(item route.Route) bool {
		if cn, ok := item.Match(ctx, llmRequest); ok {
			clusterName = cn
			r = item

			return false
		}

		return true
	})

	return r, clusterName
}

func FindCluster(ctx context.Context, llmRequest object.LLMRequest, expectedTypes []v1alpha1.ClusterType) (clusters.Cluster, bool) {
	r, clusterName := FindRoute(ctx, llmRequest)
	if r == nil {
		return nil, false
	}

	c, ok := registrycluster.FindClusterByName(clusterName)
	if !ok {
		return nil, false
	}
	if !lo.Contains(expectedTypes, c.Type()) {
		return nil, false
	}

	return c, true
}
