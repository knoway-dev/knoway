package filters

import (
	"context"
	"net/http"

	"knoway.dev/pkg/object"
	"knoway.dev/pkg/utils"
)

const (
	ListenerFilterResultTypeSucceeded = iota
	ListenerFilterResultTypeFailed
	ListenerFilterResultTypeSkipped
)

type RequestFilterResult struct {
	// Type Succeeded, Failed, or Skipped
	Type  int
	Error error
}

func (r RequestFilterResult) IsFailed() bool {
	return r.Type == ListenerFilterResultTypeFailed
}

func (r RequestFilterResult) IsSkipped() bool {
	return r.Type == ListenerFilterResultTypeSkipped
}

func (r RequestFilterResult) IsSSucceeded() bool {
	return r.Type == ListenerFilterResultTypeSucceeded
}

func NewOK() RequestFilterResult {
	return RequestFilterResult{Type: ListenerFilterResultTypeSucceeded}
}

func NewFailed(err error) RequestFilterResult {
	return RequestFilterResult{Type: ListenerFilterResultTypeFailed, Error: err}
}

type RequestFilter interface {
	isRequestFilter()
}

var _ RequestFilter = IsRequestFilter{}

type IsRequestFilter struct{}

func (IsRequestFilter) isRequestFilter() {}

type OnRequestPreFilter interface {
	RequestFilter

	OnRequestPre(ctx context.Context, sourceHTTPRequest *http.Request) RequestFilterResult
}

type OnCompletionRequestFilter interface {
	RequestFilter

	OnCompletionRequest(ctx context.Context, request object.LLMRequest, sourceHTTPRequest *http.Request) RequestFilterResult
}

type OnImageGenerationsRequestFilter interface {
	RequestFilter

	OnImageGenerationsRequest(ctx context.Context, request object.LLMRequest, sourceHTTPRequest *http.Request) RequestFilterResult
}

type OnCompletionResponseFilter interface {
	RequestFilter

	OnCompletionResponse(ctx context.Context, request object.LLMRequest, response object.LLMResponse) RequestFilterResult
}

type OnCompletionStreamResponseFilter interface {
	RequestFilter

	OnCompletionStreamResponse(ctx context.Context, request object.LLMRequest, response object.LLMStreamResponse, responseChunk object.LLMChunkResponse) RequestFilterResult
}

type OnImageGenerationsResponseFilter interface {
	RequestFilter

	OnImageGenerationsResponse(ctx context.Context, request object.LLMRequest, response object.LLMResponse) RequestFilterResult
}

type OnResponsePostFilter interface {
	RequestFilter

	OnResponsePost(ctx context.Context, request *http.Request, response any, err error)
}

type RequestFilters []RequestFilter

func (r RequestFilters) OnRequestPreFilters() []OnRequestPreFilter {
	return utils.TypeAssertFrom[RequestFilter, OnRequestPreFilter](r)
}

func (r RequestFilters) OnCompletionRequestFilters() []OnCompletionRequestFilter {
	return utils.TypeAssertFrom[RequestFilter, OnCompletionRequestFilter](r)
}

func (r RequestFilters) OnImageGenerationsRequestFilters() []OnImageGenerationsRequestFilter {
	return utils.TypeAssertFrom[RequestFilter, OnImageGenerationsRequestFilter](r)
}

func (r RequestFilters) OnCompletionResponseFilters() []OnCompletionResponseFilter {
	return utils.TypeAssertFrom[RequestFilter, OnCompletionResponseFilter](r)
}

func (r RequestFilters) OnCompletionStreamResponseFilters() []OnCompletionStreamResponseFilter {
	return utils.TypeAssertFrom[RequestFilter, OnCompletionStreamResponseFilter](r)
}

func (r RequestFilters) OnImageGenerationsResponseFilters() []OnImageGenerationsResponseFilter {
	return utils.TypeAssertFrom[RequestFilter, OnImageGenerationsResponseFilter](r)
}

func (r RequestFilters) OnResponsePostFilters() []OnResponsePostFilter {
	return utils.TypeAssertFrom[RequestFilter, OnResponsePostFilter](r)
}
