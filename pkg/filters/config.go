package filters

import (
	"context"
	"knoway.dev/api/service/v1alpha1"

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

var (
	OK = RequestFilterResult{Type: ListenerFilterResultTypeSucceeded}
)

type RequestFilter interface {
	OnCompletionRequest(ctx context.Context, request object.LLMRequest) RequestFilterResult
	OnCompletionResponse(ctx context.Context, response object.LLMResponse) RequestFilterResult
	OnCompletionStreamResponse(ctx context.Context, response object.LLMRequest, endStream bool) RequestFilterResult

	APIKeyAuth(ctx context.Context, apikey string) (*v1alpha1.APIKeyAuthResponse, error)
}
