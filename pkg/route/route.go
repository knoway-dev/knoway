package route

import (
	"context"

	routev1alpha1 "knoway.dev/api/route/v1alpha1"
	"knoway.dev/pkg/object"
)

type Route interface {
	// Match returns true if the route matches the request
	Match(ctx context.Context, request object.LLMRequest) bool
	// HandleRequest handles the request
	HandleRequest(ctx context.Context, request object.LLMRequest) (object.LLMResponse, error)

	// GetRouteConfig returns the route config
	GetRouteConfig() *routev1alpha1.Route
}
