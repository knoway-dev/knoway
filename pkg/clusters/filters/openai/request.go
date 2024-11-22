package openai

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/anypb"
	"knoway.dev/api/filters/v1alpha1"
	filters2 "knoway.dev/pkg/clusters/filters"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/protoutils"
)

func NewRequestMarshallerWithConfig(cfg *anypb.Any) (filters2.ClusterFilter, error) {
	c, err := protoutils.FromAny[*v1alpha1.OpenAIRequestMarshallerConfig](cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}

	return &requestMarshaller{
		cfg: c,
	}, nil
}

var _ filters2.ClusterFilterRequestMarshaller = (*requestMarshaller)(nil)

type requestMarshaller struct {
	filters2.IsClusterFilter

	cfg *v1alpha1.OpenAIRequestMarshallerConfig
}

func (f *requestMarshaller) MarshalRequestBody(ctx context.Context, request object.LLMRequest, pre []byte) ([]byte, error) {
	return request.GetBodyBuffer().Bytes(), nil
}
