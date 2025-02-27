package manager

import (
	"context"

	routev1alpha1 "knoway.dev/api/route/v1alpha1"
	"knoway.dev/pkg/clusters"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/registry/cluster"
	"knoway.dev/pkg/route"
	"knoway.dev/pkg/route/loadbalance"
)

var _ route.Route = (*routeManager)(nil)

type routeManager struct {
	cfg          *routev1alpha1.Route
	nsMap        map[string]string
	loadBalancer loadbalance.LoadBalancer
}

func NewWithConfig(cfg *routev1alpha1.Route) (route.Route, error) {
	rm := &routeManager{
		cfg:   cfg,
		nsMap: buildBackendNsMap(cfg),
	}

	rm.loadBalancer = loadbalance.New(cfg)

	return rm, nil
}

func (m *routeManager) SelectCluster(ctx context.Context, request object.LLMRequest) (clusters.Cluster, error) {
	clusterName := m.loadBalancer.Next(ctx, request)

	cluster, ok := cluster.FindClusterByName(clusterName)
	if !ok {
		return nil, nil
	}

	return cluster, nil
}

func (m *routeManager) Match(ctx context.Context, request object.LLMRequest) bool {
	matches := m.cfg.GetMatches()
	if len(matches) == 0 {
		return false
	}

	for _, match := range matches {
		modelNameMatch := match.GetModel()
		if modelNameMatch == nil {
			continue
		}

		exactMatch := modelNameMatch.GetExact()
		if exactMatch == "" {
			continue
		}

		if request.GetModel() != exactMatch {
			continue
		}

		if len(m.cfg.GetTargets()) == 0 {
			continue
		}

		return true
	}

	return false
}

func buildBackendNsMap(cfg *routev1alpha1.Route) map[string]string {
	nsMap := make(map[string]string)

	for _, target := range cfg.GetTargets() {
		if target.GetDestination() == nil {
			continue
		}

		nsMap[target.GetDestination().GetBackend()] = target.GetDestination().GetNamespace()
	}

	return nsMap
}

func (m *routeManager) GetRouteConfig() *routev1alpha1.Route {
	return m.cfg
}
