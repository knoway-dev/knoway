package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/anypb"

	v1alpha1clusters "knoway.dev/api/clusters/v1alpha1"
	"knoway.dev/api/filters/v1alpha1"
	"knoway.dev/pkg/bootkit"
	clusterfilters "knoway.dev/pkg/clusters/filters"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/properties"
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

func (f *requestHandler) RequestModifier(ctx context.Context, request object.LLMRequest) (object.LLMRequest, error) {
	cluster, ok := properties.GetClusterFromContext(ctx)
	if !ok {
		return request, errors.New("cluster not found in context")
	}

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
	jsonBody, err := json.Marshal(llmRequest)
	if err != nil {
		return nil, err
	}

	url := cluster.GetUpstream().GetUrl()
	url = strings.TrimSuffix(url, "/")

	switch llmRequest.GetRequestType() {
	case object.RequestTypeChatCompletions:
		url += "/chat/completions"
	case object.RequestTypeCompletions:
		url += "/completions"
	case object.RequestTypeUnknown:
		panic("unknown request type")
	default:
		panic("unknown request type: " + string(llmRequest.GetRequestType()))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	if !llmRequest.IsStream() { // non stream
		req.Header.Set("Content-Type", "application/json")
	} else { // stream
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Connection", "keep-alive")
	}

	// Apply headers
	lo.ForEach(cluster.GetUpstream().GetHeaders(), func(h *v1alpha1clusters.Upstream_Header, _ int) {
		req.Header.Set(h.GetKey(), h.GetValue())
	})

	return req, nil
}
