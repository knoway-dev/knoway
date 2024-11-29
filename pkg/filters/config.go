package filters

import (
	"context"
	"net/http"

	"knoway.dev/pkg/object"
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
	OnCompletionRequest(ctx context.Context, request object.LLMRequest, sourceHTTPRequest *http.Request) RequestFilterResult
	OnCompletionResponse(ctx context.Context, response object.LLMResponse) RequestFilterResult
	OnCompletionStreamResponse(ctx context.Context, response object.LLMRequest, endStream bool) RequestFilterResult
}
