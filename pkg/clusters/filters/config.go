// This package defines a set of dedicated interfaces of filters that can be applied to a cluster or take effects within the
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

	"knoway.dev/pkg/object"
	"knoway.dev/pkg/utils"
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

	RequestModifier(ctx context.Context, request object.LLMRequest) (object.LLMRequest, error)
}

type ClusterFilterEndpointSelector interface {
	ClusterFilter

	SelectEndpoint(ctx context.Context, request object.LLMRequest, endpoints []string) string
}

type ClusterFilterRequestMarshaller interface {
	ClusterFilter

	// MarshalRequestBody is an optional method that allows the filter to modify the request body before
	// it is sent to the upstream cluster. If pre is not nil, it contains the body of the previous filter
	// in the chain.
	MarshalRequestBody(ctx context.Context, request object.LLMRequest, pre []byte) ([]byte, error)
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

	ResponseModifier(ctx context.Context, request object.LLMRequest, response object.LLMResponse) (object.LLMResponse, error)
}

type ClusterFilterResponseComplete interface {
	ClusterFilter

	ResponseComplete(ctx context.Context, request object.LLMRequest, response object.LLMResponse) error
}

type ClusterFilters []ClusterFilter

func (c ClusterFilters) OnRequestPreflights() []ClusterFilterRequestPreflight {
	return utils.TypeAssertFrom[ClusterFilter, ClusterFilterRequestPreflight](c)
}

func (c ClusterFilters) OnRequestModifiers() []ClusterFilterRequestModifier {
	return utils.TypeAssertFrom[ClusterFilter, ClusterFilterRequestModifier](c)
}

func (c ClusterFilters) OnEndpointSelectors() []ClusterFilterEndpointSelector {
	return utils.TypeAssertFrom[ClusterFilter, ClusterFilterEndpointSelector](c)
}

func (c ClusterFilters) OnRequestMarshallers() []ClusterFilterRequestMarshaller {
	return utils.TypeAssertFrom[ClusterFilter, ClusterFilterRequestMarshaller](c)
}

func (c ClusterFilters) OnResponseUnmarshallers() []ClusterFilterResponseUnmarshaller {
	return utils.TypeAssertFrom[ClusterFilter, ClusterFilterResponseUnmarshaller](c)
}

func (c ClusterFilters) OnResponseModifiers() []ClusterFilterResponseModifier {
	return utils.TypeAssertFrom[ClusterFilter, ClusterFilterResponseModifier](c)
}

func (c ClusterFilters) OnResponseCompleters() []ClusterFilterResponseComplete {
	return utils.TypeAssertFrom[ClusterFilter, ClusterFilterResponseComplete](c)
}
