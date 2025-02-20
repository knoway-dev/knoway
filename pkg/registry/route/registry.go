package route

import (
	"log/slog"
	"sync"

	"github.com/samber/lo"

	"knoway.dev/api/route/v1alpha1"
	"knoway.dev/pkg/route"
	"knoway.dev/pkg/route/manager"
)

var (
	routes        = make([]route.Route, 0)
	routeRegistry = make(map[string]route.Route)
	routeLock     sync.RWMutex
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
					Backend: modelName,
				},
			},
		},
		Filters: nil, // todo future
	}
}

func RegisterRouteWithConfig(cfg *v1alpha1.Route) error {
	routeLock.Lock()
	defer routeLock.Unlock()

	r, err := manager.NewWithConfig(cfg)
	if err != nil {
		return err
	}

	routeRegistry[cfg.GetName()] = r
	routes = lo.Values(routeRegistry)

	slog.Info("register route", "name", cfg.GetName())

	return nil
}

func RemoveRoute(rName string) {
	routeLock.Lock()
	defer routeLock.Unlock()

	delete(routeRegistry, rName)
	routes = lo.Values(routeRegistry)
}

func ForeachRoute(f func(route.Route) bool) {
	routeLock.RLock()
	defer routeLock.RUnlock()

	for _, r := range routes {
		continue_ := f(r)
		if !continue_ {
			break
		}
	}
}
