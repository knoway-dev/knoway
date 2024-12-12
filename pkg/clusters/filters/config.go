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

	"github.com/samber/lo"

	"knoway.dev/api/clusters/v1alpha1"
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

type ClusterFilterUpstreamRequestMarshaller interface {
	ClusterFilter

	// MarshalRequestBody is an optional method that allows the filter to modify the request body before
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

	ResponseModifier(ctx context.Context, request object.LLMRequest, response object.LLMResponse) (object.LLMResponse, error)
}

type ClusterFilterResponseComplete interface {
	ClusterFilter

	ResponseComplete(ctx context.Context, request object.LLMRequest, response object.LLMResponse) error
}

type ClusterFilters []ClusterFilter

func (c ClusterFilters) RequestPreflights() []ClusterFilterRequestPreflight {
	return utils.TypeAssertFrom[ClusterFilter, ClusterFilterRequestPreflight](c)
}

func (c ClusterFilters) ForEachRequestPreflight(ctx context.Context, request object.LLMRequest) error {
	for _, f := range c.RequestPreflights() {
		err := f.RequestPreflight(ctx, request)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c ClusterFilters) RequestModifiers() []ClusterFilterRequestModifier {
	return utils.TypeAssertFrom[ClusterFilter, ClusterFilterRequestModifier](c)
}

func (c ClusterFilters) ForEachRequestModifier(ctx context.Context, request object.LLMRequest) (object.LLMRequest, error) {
	for _, f := range c.RequestModifiers() {
		var err error

		request, err = f.RequestModifier(ctx, request)
		if err != nil {
			return nil, err
		}
	}

	return request, nil
}

func (c ClusterFilters) EndpointSelectors() []ClusterFilterEndpointSelector {
	return utils.TypeAssertFrom[ClusterFilter, ClusterFilterEndpointSelector](c)
}

func (c ClusterFilters) ForEachEndpointSelector(ctx context.Context, request object.LLMRequest, endpoints []string) string {
	for _, f := range c.EndpointSelectors() {
		selected := f.SelectEndpoint(ctx, request, endpoints)
		if selected != "" {
			return selected
		}
	}

	return ""
}

func (c ClusterFilters) UpstreamRequestMarshallers() []ClusterFilterUpstreamRequestMarshaller {
	return utils.TypeAssertFrom[ClusterFilter, ClusterFilterUpstreamRequestMarshaller](c)
}

func (c ClusterFilters) ForEachUpstreamRequestMarshaller(ctx context.Context, cluster *v1alpha1.Cluster, llmRequest object.LLMRequest, request *http.Request) (*http.Request, error) {
	for _, f := range c.UpstreamRequestMarshallers() {
		var err error

		request, err = f.MarshalUpstreamRequest(ctx, cluster, llmRequest, request)
		if err != nil {
			return nil, err
		}
	}
	if request == nil {
		panic("ClusterFilterUpstreamRequestMarshaller iterated, but returned nil request or no filters found")
	}

	return request, nil
}

func (c ClusterFilters) ResponseUnmarshallers() []ClusterFilterResponseUnmarshaller {
	return utils.TypeAssertFrom[ClusterFilter, ClusterFilterResponseUnmarshaller](c)
}

func (c ClusterFilters) ForEachResponseUnmarshaller(ctx context.Context, request object.LLMRequest, rawResponse *http.Response, reader *bufio.Reader, pre object.LLMResponse) (object.LLMResponse, error) {
	for _, f := range c.ResponseUnmarshallers() {
		var err error

		pre, err = f.UnmarshalResponseBody(ctx, request, rawResponse, reader, pre)
		if err != nil {
			return nil, err
		}
	}

	return pre, nil
}

func (c ClusterFilters) ResponseModifiers() []ClusterFilterResponseModifier {
	return utils.TypeAssertFrom[ClusterFilter, ClusterFilterResponseModifier](c)
}

func (c ClusterFilters) ForEachResponseModifier(ctx context.Context, request object.LLMRequest, response object.LLMResponse) (object.LLMResponse, error) {
	for _, f := range lo.Reverse(c).ResponseModifiers() {
		var err error

		response, err = f.ResponseModifier(ctx, request, response)
		if err != nil {
			return nil, err
		}
	}

	return response, nil
}

func (c ClusterFilters) ResponseCompleters() []ClusterFilterResponseComplete {
	return utils.TypeAssertFrom[ClusterFilter, ClusterFilterResponseComplete](c)
}

func (c ClusterFilters) ForEachResponseComplete(ctx context.Context, request object.LLMRequest, response object.LLMResponse) error {
	for _, f := range lo.Reverse(c).ResponseCompleters() {
		err := f.ResponseComplete(ctx, request, response)
		if err != nil {
			return err
		}
	}

	return nil
}
