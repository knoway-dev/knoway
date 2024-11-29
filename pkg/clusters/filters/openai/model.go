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

func NewModelNameRewriteWithConfig(cfg *anypb.Any, _ bootkit.LifeCycle) (clusterfilters.ClusterFilter, error) {
	c, err := protoutils.FromAny(cfg, &v1alpha1.OpenAIModelNameRewriteConfig{})
	if err != nil {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}

	return &modelNameRewriter{
		cfg: c,
	}, nil
}

var _ clusterfilters.ClusterFilterRequestHandler = (*modelNameRewriter)(nil)

type modelNameRewriter struct {
	clusterfilters.IsClusterFilter

	cfg *v1alpha1.OpenAIModelNameRewriteConfig
}

func (f *modelNameRewriter) RequestPreflight(ctx context.Context, request object.LLMRequest) (object.LLMRequest, error) {
	err := request.SetModel(f.cfg.GetModelName())
	if err != nil {
		return nil, err
	}

	return request, nil
}
