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

type OnCompletionRequestFilter interface {
	isRequestFilter()

	OnCompletionRequest(ctx context.Context, request object.LLMRequest, sourceHTTPRequest *http.Request) RequestFilterResult
}

type OnCompletionResponseFilter interface {
	isRequestFilter()

	OnCompletionResponse(ctx context.Context, response object.LLMResponse) RequestFilterResult
}

type OnCompletionStreamResponse interface {
	isRequestFilter()

	OnCompletionStreamResponse(ctx context.Context, response object.LLMRequest, endStream bool) RequestFilterResult
}

type RequestFilter interface {
	isRequestFilter()
}

type RequestFilters []RequestFilter

func (r RequestFilters) OnCompletionRequestFilters() []OnCompletionRequestFilter {
	return utils.TypeAssertFrom[RequestFilter, OnCompletionRequestFilter](r)
}

func (r RequestFilters) OnCompletionResponseFilters() []OnCompletionResponseFilter {
	return utils.TypeAssertFrom[RequestFilter, OnCompletionResponseFilter](r)
}

func (r RequestFilters) OnCompletionStreamResponseFilters() []OnCompletionStreamResponse {
	return utils.TypeAssertFrom[RequestFilter, OnCompletionStreamResponse](r)
}

type IsRequestFilter struct{}

func (IsRequestFilter) isRequestFilter() {}
