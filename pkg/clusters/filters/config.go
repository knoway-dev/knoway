// Package filters defines a set of dedicated interfaces of filters that can be applied to a cluster or take effects within the
// scope of cluster operations. The filters are applied in a chain of responsibility pattern, where each filter is responsible
// for a specific operation. The filters are divided into two categories: request filters and response filters.
//
// For simple illustrations, the workflow can be described as follows:
//
// Incoming Request -> Request Preflight x n -> Request Modifier x n -> Endpoint Selector -> Request Marshaller -> Outgoing Request
//
// Incoming Response -> Response Unmarshaller -> Response Modifier x n -> Response Completer x n -> Outgoing Response
//
// The filters are applied in the order they are defined in the configuration.
package filters

import (
	"bufio"
	"context"
	"net/http"

	"knoway.dev/api/clusters/v1alpha1"
	"knoway.dev/pkg/object"
)

type ClusterFilter interface {
	isClusterFilter()
}

type IsClusterFilter struct{}

func (IsClusterFilter) isClusterFilter() {}

type ClusterFilterRequestPreflight interface {
	ClusterFilter

	RequestPreflight(ctx context.Context, request object.LLMRequest) error
}

type ClusterFilterRequestModifier interface {
	ClusterFilter

	RequestModifier(ctx context.Context, cluster *v1alpha1.Cluster, request object.LLMRequest) (object.LLMRequest, error)
}

type ClusterFilterEndpointSelector interface {
	ClusterFilter

	SelectEndpoint(ctx context.Context, request object.LLMRequest, endpoints []string) string
}

type ClusterFilterUpstreamRequestMarshaller interface {
	ClusterFilter

	// MarshalUpstreamRequest is an optional method that allows the filter to modify the request body before
	// it is sent to the upstream cluster. If pre is not nil, it contains the body of the previous filter
	// in the chain.
	MarshalUpstreamRequest(ctx context.Context, cluster *v1alpha1.Cluster, llmRequest object.LLMRequest, request *http.Request) (*http.Request, error)
}

type ClusterFilterResponseUnmarshaller interface {
	ClusterFilter

	// UnmarshalResponseBody is an optional method that allows the filter to modify the response body
	// before it is sent to the client. If pre is not nil, it contains the body of the previous filter in
	// the chain.
	UnmarshalResponseBody(ctx context.Context, request object.LLMRequest, rawResponse *http.Response, reader *bufio.Reader, pre object.LLMResponse) (object.LLMResponse, error)
}

type ClusterFilterResponseModifier interface {
	ClusterFilter

	ResponseModifier(ctx context.Context, cluster *v1alpha1.Cluster, request object.LLMRequest, response object.LLMResponse) (object.LLMResponse, error)
}

type ClusterFilterResponseComplete interface {
	ClusterFilter

	ResponseComplete(ctx context.Context, request object.LLMRequest, response object.LLMResponse) error
}
