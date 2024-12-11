package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"google.golang.org/protobuf/types/known/anypb"

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
var _ clusterfilters.ClusterFilterRequestMarshaller = (*requestHandler)(nil)

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

func (f *requestHandler) MarshalRequestBody(ctx context.Context, request object.LLMRequest, pre []byte) ([]byte, error) {
	return json.Marshal(request)
}
