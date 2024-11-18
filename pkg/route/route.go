package route

import (
	"context"
	"knoway.dev/pkg/object"
)

type Route interface {
	// Match returns the cluster name and a boolean indicating if the request matched the route.
	Match(ctx context.Context, request object.LLMRequest) (string, bool)
}
