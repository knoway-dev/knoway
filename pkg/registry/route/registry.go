package route

import (
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

func RegisterRouteWithConfig(cfg *v1alpha1.Route) error {
	routeLock.Lock()
	defer routeLock.Unlock()

	r, err := manager.NewWithConfig(cfg)
	if err != nil {
		return err
	}

	routeRegistry[cfg.GetName()] = r
	// todo reorder routes
	routes = lo.Values(routeRegistry)

	return nil
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
