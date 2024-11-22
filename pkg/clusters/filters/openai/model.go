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

func NewModelNameRewriteWithConfig(cfg *anypb.Any) (filters2.ClusterFilter, error) {
	c, err := protoutils.FromAny[*v1alpha1.OpenAIModelNameRewriteConfig](cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}

	return &modelNameRewriter{
		cfg: c,
	}, nil
}

var _ filters2.ClusterFilterRequestHandler = (*modelNameRewriter)(nil)

type modelNameRewriter struct {
	filters2.IsClusterFilter

	cfg *v1alpha1.OpenAIModelNameRewriteConfig
}

func (f *modelNameRewriter) RequestPreflight(ctx context.Context, request object.LLMRequest) (object.LLMRequest, error) {
	err := request.SetModel(f.cfg.ModelName)
	if err != nil {
		return nil, err
	}

	return request, nil
}
