package openai

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/protobuf/types/known/anypb"

	"knoway.dev/api/filters/v1alpha1"
	"knoway.dev/pkg/bootkit"
	clusterfilters "knoway.dev/pkg/clusters/filters"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/properties"
	"knoway.dev/pkg/protoutils"
	"knoway.dev/pkg/types/openai"
)

func NewResponseHandlerWithConfig(cfg *anypb.Any, _ bootkit.LifeCycle) (clusterfilters.ClusterFilter, error) {
	c, err := protoutils.FromAny(cfg, &v1alpha1.OpenAIResponseHandlerConfig{})
	if err != nil {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}

	return &responseHandler{
		cfg: c,
	}, nil
}

var _ clusterfilters.ClusterFilterResponseUnmarshaller = (*responseHandler)(nil)
var _ clusterfilters.ClusterFilterResponseModifier = (*responseHandler)(nil)

type responseHandler struct {
	cfg *v1alpha1.OpenAIResponseHandlerConfig
	clusterfilters.ClusterFilter
}

func (f *responseHandler) UnmarshalResponseBody(ctx context.Context, req object.LLMRequest, rawResponse *http.Response, reader *bufio.Reader, pre object.LLMResponse) (object.LLMResponse, error) {
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

func (f *responseHandler) ResponseModifier(ctx context.Context, request object.LLMRequest, response object.LLMResponse) (object.LLMResponse, error) {
	rp := properties.GetRequestFromCtx(ctx)
	if rp.Cluster == nil {
		return response, errors.New("cluster not found in context")
	}

	err := response.SetModel(rp.Cluster.GetName())
	if err != nil {
		return response, err
	}

	return response, nil
}
