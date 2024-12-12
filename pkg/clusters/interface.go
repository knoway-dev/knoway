package clusters

import (
	"context"

	"knoway.dev/pkg/clusters/filters"
	"knoway.dev/pkg/object"
)

type Cluster interface {
	DoUpstreamRequest(ctx context.Context, req object.LLMRequest) (object.LLMResponse, error)
	DoUpstreamResponseComplete(ctx context.Context, req object.LLMRequest, res object.LLMResponse) error
	LoadFilters() filters.ClusterFilters
}
