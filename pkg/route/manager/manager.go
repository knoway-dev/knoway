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

	// TODO: implement
	for _, match := range matches {
		modelNameMatch := match.GetModel()
		if modelNameMatch == nil {
			continue
		}

		exactMatch := modelNameMatch.GetExact()
		if exactMatch == "" {
			continue
		}

		if request.GetModel() == exactMatch {
			return m.cfg.GetClusterName(), true
		}
	}

	return "", false
}
