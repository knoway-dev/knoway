package manager

import (
	"context"
	"knoway.dev/api/route/v1alpha1"
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/route"
)

type routeManager struct {
	cfg     *v1alpha1.Route
	filters []filters.RequestFilter
	route.Route
}

func NewWithConfig(cfg *v1alpha1.Route) (route.Route, error) {
	rm := &routeManager{
		cfg: cfg,
	}
	return rm, nil
}

func (m *routeManager) Match(ctx context.Context, request object.LLMRequest) (string, bool) {
	if len(m.cfg.GetMatches()) != 0 {
		// todo implement
	}
	return m.cfg.GetClusterName(), true
}
