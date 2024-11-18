package stats

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/anypb"
	"knoway.dev/api/filters/v1alpha1"
	filters2 "knoway.dev/pkg/clusters/filters"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/protoutils"
)

func NewWithConfig(cfg *anypb.Any) (filters2.ClusterFilter, error) {
	c, err := protoutils.FromAny[*v1alpha1.UsageStatsConfig](cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}
	return &usageStatsFilter{
		cfg: c,
	}, nil
}

type usageStatsFilter struct {
	cfg *v1alpha1.UsageStatsConfig
	filters2.ClusterFilter
}

func (f *usageStatsFilter) OnResponseComplete(ctx context.Context, response object.LLMResponse) error {
	usage := response.GetUsage()
	if usage == nil {
		return nil
	}
	// todo
	return nil
}
