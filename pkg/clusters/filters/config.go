package filters

import (
	"bufio"
	"context"
	"net/http"

	"knoway.dev/pkg/object"
)

type ClusterFilterRequestMarshaller interface {
	isClusterFilter()

	// MarshalRequestBody is an optional method that allows the filter to modify the request body before
	// it is sent to the upstream cluster. If pre is not nil, it contains the body of the previous filter
	// in the chain.
	MarshalRequestBody(ctx context.Context, request object.LLMRequest, pre []byte) ([]byte, error)
}

type ClusterFilterResponseUnmarshaller interface {
	isClusterFilter()

	// UnmarshalResponseBody is an optional method that allows the filter to modify the response body
	// before it is sent to the client. If pre is not nil, it contains the body of the previous filter in
	// the chain.
	UnmarshalResponseBody(ctx context.Context, request object.LLMRequest, rawResponse *http.Response, reader *bufio.Reader, pre object.LLMResponse) (object.LLMResponse, error)
}

type ClusterFilterEndpointSelector interface {
	isClusterFilter()

	SelectEndpoint(ctx context.Context, request object.LLMRequest, endpoints []string) string
}

type ClusterFilterRequestHandler interface {
	isClusterFilter()

	RequestPreflight(ctx context.Context, request object.LLMRequest) (object.LLMRequest, error)
}

type ClusterFilterResponseHandler interface {
	isClusterFilter()

	ResponseComplete(ctx context.Context, request object.LLMRequest, response object.LLMResponse) error
}

type ClusterFilter interface {
	isClusterFilter()
}

type IsClusterFilter struct{}

func (IsClusterFilter) isClusterFilter() {}
