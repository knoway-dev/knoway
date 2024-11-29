package openai

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/anypb"

	"knoway.dev/api/filters/v1alpha1"
	"knoway.dev/pkg/bootkit"
	clusterfilters "knoway.dev/pkg/clusters/filters"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/protoutils"
)

func NewRequestMarshallerWithConfig(cfg *anypb.Any, _ bootkit.LifeCycle) (clusterfilters.ClusterFilter, error) {
	c, err := protoutils.FromAny(cfg, &v1alpha1.OpenAIRequestMarshallerConfig{})
	if err != nil {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}

	return &requestMarshaller{
		cfg: c,
	}, nil
}

var _ clusterfilters.ClusterFilterRequestMarshaller = (*requestMarshaller)(nil)

type requestMarshaller struct {
	clusterfilters.IsClusterFilter

	cfg *v1alpha1.OpenAIRequestMarshallerConfig
}

func (f *requestMarshaller) MarshalRequestBody(ctx context.Context, request object.LLMRequest, pre []byte) ([]byte, error) {
	return request.GetBodyBuffer().Bytes(), nil
}
