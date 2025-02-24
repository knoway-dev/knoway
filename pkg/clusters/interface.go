package clusters

import (
	"context"

	"knoway.dev/api/clusters/v1alpha1"
	"knoway.dev/pkg/object"
)

type Cluster interface {
	Type() v1alpha1.ClusterType
	Config() *v1alpha1.Cluster
	DoUpstreamRequest(ctx context.Context, req object.LLMRequest) (object.LLMResponse, error)
	DoUpstreamResponseComplete(ctx context.Context, req object.LLMRequest, res object.LLMResponse) error
}
