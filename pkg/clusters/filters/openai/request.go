package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/anypb"

	v1alpha1clusters "knoway.dev/api/clusters/v1alpha1"
	"knoway.dev/api/filters/v1alpha1"
	"knoway.dev/pkg/bootkit"
	clusterfilters "knoway.dev/pkg/clusters/filters"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/protoutils"
)

func NewRequestHandlerWithConfig(cfg *anypb.Any, _ bootkit.LifeCycle) (clusterfilters.ClusterFilter, error) {
	c, err := protoutils.FromAny(cfg, &v1alpha1.OpenAIRequestHandlerConfig{})
	if err != nil {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}

	return &requestHandler{
		cfg: c,
	}, nil
}

var _ clusterfilters.ClusterFilterRequestModifier = (*requestHandler)(nil)
var _ clusterfilters.ClusterFilterUpstreamRequestMarshaller = (*requestHandler)(nil)

type requestHandler struct {
	clusterfilters.IsClusterFilter

	cfg *v1alpha1.OpenAIRequestHandlerConfig
}

func (f *requestHandler) RequestModifier(ctx context.Context, cluster *v1alpha1clusters.Cluster, request object.LLMRequest) (object.LLMRequest, error) {
	err := request.SetModel(cluster.GetName())
	if err != nil {
		return request, err
	}

	err = request.SetDefaultParams(cluster.GetUpstream().GetDefaultParams())
	if err != nil {
		return request, err
	}

	err = request.SetOverrideParams(cluster.GetUpstream().GetOverrideParams())
	if err != nil {
		return request, err
	}

	return request, nil
}

func (f *requestHandler) MarshalUpstreamRequest(ctx context.Context, cluster *v1alpha1clusters.Cluster, llmRequest object.LLMRequest, request *http.Request) (*http.Request, error) {
	upstreamURL := cluster.GetUpstream().GetUrl()
	upstreamURL = strings.TrimSuffix(upstreamURL, "/")

	switch llmRequest.GetRequestType() {
	case object.RequestTypeChatCompletions:
		upstreamURL += "/chat/completions"
	case object.RequestTypeCompletions:
		upstreamURL += "/completions"
	default:
		panic("unknown request type: " + string(llmRequest.GetRequestType()))
	}

	parsedUpstreamURL, err := url.Parse(upstreamURL)
	if err != nil {
		return nil, err
	}

	jsonBody, err := json.Marshal(llmRequest)
	if err != nil {
		return nil, err
	}

	if request == nil {
		request, err = http.NewRequestWithContext(ctx, http.MethodPost, upstreamURL, bytes.NewReader(jsonBody))
		if err != nil {
			return nil, err
		}
	} else {
		request.URL = parsedUpstreamURL
		request.Method = http.MethodPost
		request.Body = io.NopCloser(bytes.NewReader(jsonBody))
	}

	request.Header.Set("Content-Type", "application/json")
	// Apply headers
	if llmRequest.IsStream() { // non stream
		request.Header.Set("Accept", "text/event-stream")
		request.Header.Set("Cache-Control", "no-cache")
		request.Header.Set("Connection", "keep-alive")
	}

	// Apply user-defined headers
	lo.ForEach(cluster.GetUpstream().GetHeaders(), func(h *v1alpha1clusters.Upstream_Header, _ int) {
		request.Header.Set(h.GetKey(), h.GetValue())
	})

	return request, nil
}
