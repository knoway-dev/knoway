package openai

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/protobuf/types/known/anypb"
	"knoway.dev/api/filters/v1alpha1"
	filters2 "knoway.dev/pkg/clusters/filters"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/protoutils"
	"knoway.dev/pkg/types/openai"
)

func NewResponseUnmarshallerWithConfig(cfg *anypb.Any) (filters2.ClusterFilter, error) {
	c, err := protoutils.FromAny[*v1alpha1.OpenAIResponseUnmarshallerConfig](cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}

	return &responseUnmarshaller{
		cfg: c,
	}, nil
}

var _ filters2.ClusterFilterResponseUnmarshaller = (*responseUnmarshaller)(nil)

type responseUnmarshaller struct {
	cfg *v1alpha1.OpenAIResponseUnmarshallerConfig
	filters2.ClusterFilter
}

func (f *responseUnmarshaller) UnmarshalResponseBody(ctx context.Context, req object.LLMRequest, rawResponse *http.Response, buffer *bytes.Buffer, pre object.LLMResponse) (object.LLMResponse, error) {
	contentType := rawResponse.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		return nil, fmt.Errorf("unexpected content type %s on status %d", contentType, rawResponse.StatusCode)
	}

	return openai.NewChatCompletionResponse(rawResponse, buffer)
}
