package route

import (
	"context"

	routev1alpha1 "knoway.dev/api/route/v1alpha1"

	"knoway.dev/pkg/object"
)

type Route interface {
	// Match returns the cluster name and a boolean indicating if the request matched the route.
	Match(ctx context.Context, request object.LLMRequest) (string, bool)

	GetRouteConfig() *routev1alpha1.Route
}
