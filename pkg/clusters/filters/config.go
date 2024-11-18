package filters

import (
	"context"
	"knoway.dev/pkg/object"
)

type ClusterFilter interface {
	// MarshalRequestBody is an optional method that allows the filter to modify the request body before it is sent to the upstream cluster.
	// if pre is not nil, it contains the body of the previous filter in the chain.
	MarshalRequestBody(ctx context.Context, request object.LLMRequest, pre []byte) ([]byte, error)
	// UnmarshalResponseBody is an optional method that allows the filter to modify the response body before it is sent to the client.
	// if pre is not nil, it contains the body of the previous filter in the chain.
	UnmarshalResponseBody(ctx context.Context, bs []byte, pre object.LLMResponse) (object.LLMResponse, error)
	SelectEndpoint(ctx context.Context, endpoints []string) string
	OnResponseComplete(ctx context.Context, response object.LLMResponse) error
}
