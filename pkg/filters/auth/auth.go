package auth

import (
	context "context"
	"fmt"

	"google.golang.org/protobuf/types/known/anypb"
	"knoway.dev/api/filters/v1alpha1"
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/protoutils"
)

func NewWithConfig(cfg *anypb.Any) (filters.RequestFilter, error) {
	c, err := protoutils.FromAny[*v1alpha1.APIKeyAuthConfig](cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}
	return &AuthFilter{
		config: c,
	}, nil
}

type AuthFilter struct {
	config *v1alpha1.APIKeyAuthConfig
}

func (a *AuthFilter) OnCompletionRequest(ctx context.Context, request object.LLMRequest) filters.RequestFilterResult {
	return filters.OK
}

func (a *AuthFilter) OnCompletionResponse(ctx context.Context, response object.LLMResponse) filters.RequestFilterResult {
	return filters.OK
}

func (a *AuthFilter) OnCompletionStreamResponse(ctx context.Context, response object.LLMRequest, endStream bool) filters.RequestFilterResult {
	return filters.OK
}
