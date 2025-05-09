package manager

import (
	"context"
	"log/slog"
	"sync"

	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/metadata"
	"knoway.dev/pkg/object"

	"knoway.dev/api/route/v1alpha1"
	"knoway.dev/pkg/route"
	rroute "knoway.dev/pkg/route/route"

	"github.com/samber/lo"
)

var (
	matchRouteRegistry = make(map[string]route.Route)
	routeRegistry      = make(map[string]route.Route)

	routes    = make([]route.Route, 0)
	routeLock sync.RWMutex
)

func InitDirectModelRoute(modelName string) *v1alpha1.Route {
	return &v1alpha1.Route{
		Name: modelName,
		Matches: []*v1alpha1.Match{
			{
				Model: &v1alpha1.StringMatch{
					Match: &v1alpha1.StringMatch_Exact{
						Exact: modelName,
					},
				},
			},
		},
		Targets: []*v1alpha1.RouteTarget{
			{
				Destination: &v1alpha1.RouteDestination{
					Cluster: modelName,
				},
			},
		},
		Filters: nil, // todo future
	}
}

func RegisterMatchRouteWithConfig(cfg *v1alpha1.Route, lifecycle bootkit.LifeCycle) error {
	routeLock.Lock()
	defer routeLock.Unlock()

	r, err := rroute.NewWithConfig(cfg, lifecycle)
	if err != nil {
		return err
	}

	matchRouteRegistry[cfg.GetName()] = r
	routes = mergeRoutes()

	slog.Info("register match route", "name", cfg.GetName())

	return nil
}

func RemoveMatchRoute(rName string) {
	routeLock.Lock()
	defer routeLock.Unlock()

	delete(matchRouteRegistry, rName)
	routes = mergeRoutes()

	slog.Info("remove match route", "name", rName)
}

func RegisterBaseRouteWithConfig(cfg *v1alpha1.Route, lifecycle bootkit.LifeCycle) error {
	routeLock.Lock()
	defer routeLock.Unlock()

	r, err := rroute.NewWithConfig(cfg, lifecycle)
	if err != nil {
		return err
	}

	routeRegistry[cfg.GetName()] = r

	if _, exists := matchRouteRegistry[cfg.GetName()]; exists {
		slog.Info("route exists in matchRouteRegistry, skipping base route registration", "name", cfg.GetName())
		return nil
	}

	routes = mergeRoutes()

	slog.Info("register base route", "name", cfg.GetName())

	return nil
}

func RemoveBaseRoute(rName string) {
	routeLock.Lock()
	defer routeLock.Unlock()

	delete(routeRegistry, rName)
	routes = mergeRoutes()

	slog.Info("remove base route", "name", rName)
}

func mergeRoutes() []route.Route {
	uniqueRoutes := make(map[string]route.Route)

	for k, v := range matchRouteRegistry {
		uniqueRoutes[k] = v
	}

	for k, v := range routeRegistry {
		if _, exists := uniqueRoutes[k]; !exists {
			uniqueRoutes[k] = v
		}
	}

	return lo.Values(uniqueRoutes)
}

func MatchRoute(ctx context.Context, request object.LLMRequest) route.Route {
	routeLock.RLock()
	defer routeLock.RUnlock()

	for _, r := range routes {
		if r.Match(ctx, request) {
			return r
		}
	}

	return nil
}

func HandleRequest(ctx context.Context, llmRequest object.LLMRequest) (object.LLMResponse, error) {
	route := MatchRoute(ctx, llmRequest)
	if route == nil {
		return nil, object.NewErrorModelNotFoundOrNotAccessible(llmRequest.GetModel())
	}

	rMeta := metadata.RequestMetadataFromCtx(ctx)
	rMeta.MatchRoute = route

	return route.HandleRequest(ctx, llmRequest)
}

func DebugDumpAllRoutes() []*v1alpha1.Route {
	routeLock.Lock()
	defer routeLock.Unlock()

	return lo.Map(routes, func(r route.Route, _ int) *v1alpha1.Route {
		return r.GetRouteConfig()
	})
}
