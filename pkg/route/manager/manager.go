package manager

import (
	"context"

	"knoway.dev/api/route/v1alpha1"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/route"
)

type routeManager struct {
	cfg *v1alpha1.Route
	// filters []filters.RequestFilter
	route.Route
}

func NewWithConfig(cfg *v1alpha1.Route) (route.Route, error) {
	rm := &routeManager{
		cfg: cfg,
	}

	return rm, nil
}

func (m *routeManager) Match(ctx context.Context, request object.LLMRequest) (string, bool) {
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

		// TODO: should use load balancing algorithm to select on target
		return m.cfg.GetTargets()[0].GetDestination().GetBackend(), true
	}

	return "", false
}
