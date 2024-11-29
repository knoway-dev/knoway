package openai

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/protobuf/types/known/anypb"

	"knoway.dev/api/filters/v1alpha1"
	"knoway.dev/pkg/bootkit"
	clusterfilters "knoway.dev/pkg/clusters/filters"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/protoutils"
	"knoway.dev/pkg/types/openai"
)

func NewResponseUnmarshallerWithConfig(cfg *anypb.Any, _ bootkit.LifeCycle) (clusterfilters.ClusterFilter, error) {
	c, err := protoutils.FromAny(cfg, &v1alpha1.OpenAIResponseUnmarshallerConfig{})
	if err != nil {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}

	return &responseUnmarshaller{
		cfg: c,
	}, nil
}

var _ clusterfilters.ClusterFilterResponseUnmarshaller = (*responseUnmarshaller)(nil)

type responseUnmarshaller struct {
	cfg *v1alpha1.OpenAIResponseUnmarshallerConfig
	clusterfilters.ClusterFilter
}

func (f *responseUnmarshaller) UnmarshalResponseBody(ctx context.Context, req object.LLMRequest, rawResponse *http.Response, reader *bufio.Reader, pre object.LLMResponse) (object.LLMResponse, error) {
	contentType := rawResponse.Header.Get("Content-Type")

	switch {
	case strings.HasPrefix(contentType, "application/json"):
		return openai.NewChatCompletionResponse(req, rawResponse, reader)
	case strings.HasPrefix(contentType, "text/event-stream"):
		return openai.NewChatCompletionStreamResponse(req, rawResponse, reader)
	default:
		return nil, fmt.Errorf("unsupported content type %s", contentType)
	}
}
