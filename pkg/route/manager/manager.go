package manager

import (
	"context"

	"knoway.dev/pkg/filters/lbfilter/loadbanlance"

	routev1alpha1 "knoway.dev/api/route/v1alpha1"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/route"
)

type routeManager struct {
	cfg *routev1alpha1.Route
	// filters []filters.RequestFilter
	route.Route
	lb    loadbanlance.LoadBalancer
	nsMap map[string]string
}

func NewWithConfig(cfg *routev1alpha1.Route) (route.Route, error) {
	rm := &routeManager{
		cfg:   cfg,
		lb:    loadbanlance.New(cfg),
		nsMap: buildBackendNsMap(cfg),
	}

	return rm, nil
}

func (m *routeManager) Match(ctx context.Context, request object.LLMRequest) (string, bool) {
	var (
		clusterName string
		found       bool
	)

	defer func() {
		if found {
			m.lb.Done()
		}
	}()

	matches := m.cfg.GetMatches()
	if len(matches) == 0 {
		return "", false
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

		// default lb policy
		if m.cfg.GetLoadBalancePolicy() == routev1alpha1.LoadBalancePolicy_LOAD_BALANCE_POLICY_UNSPECIFIED {
			return m.cfg.GetTargets()[0].GetDestination().GetCluster(), m.cfg.GetTargets()[0].GetDestination().GetCluster() != ""
		}

		if cluster := m.lb.Next(request); cluster != "" {
			clusterName = cluster
			found = true

			break
		}
	}

	return clusterName, found
}

func buildBackendNsMap(cfg *routev1alpha1.Route) map[string]string {
	nsMap := make(map[string]string)

	for _, target := range cfg.GetTargets() {
		if target.GetDestination() == nil {
			continue
		}

		ns := target.GetDestination().GetNamespace()
		if ns == "" {
			ns = "public"
		}
		nsMap[target.GetDestination().GetBackend()] = ns
	}

	return nsMap
}

func (m *routeManager) GetRouteConfig() *routev1alpha1.Route {
	if m == nil || m.cfg == nil {
		return nil
	}

	return m.cfg
}
