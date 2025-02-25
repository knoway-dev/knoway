package lbfilter

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/samber/mo"
	"google.golang.org/protobuf/types/known/anypb"

	"knoway.dev/api/clusters/v1alpha1"
	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/clusters"
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/metadata"
	"knoway.dev/pkg/object"
	registrycluster "knoway.dev/pkg/registry/cluster"
	registryroute "knoway.dev/pkg/registry/route"
	"knoway.dev/pkg/route"
)

func NewWithConfig(_ *anypb.Any, _ bootkit.LifeCycle) (filters.RequestFilter, error) {
	return &LBFilter{}, nil
}

type LBFilter struct {
	filters.IsRequestFilter
}

var _ filters.RequestFilter = (*LBFilter)(nil)
var _ filters.OnCompletionRequestFilter = (*LBFilter)(nil)
var _ filters.OnImageGenerationsRequestFilter = (*LBFilter)(nil)

func (l *LBFilter) OnImageGenerationsRequest(ctx context.Context, request object.LLMRequest, sourceHTTPRequest *http.Request) filters.RequestFilterResult {
	return onRequest(ctx, request)
}

func (l *LBFilter) OnCompletionRequest(ctx context.Context, request object.LLMRequest, sourceHTTPRequest *http.Request) filters.RequestFilterResult {
	return onRequest(ctx, request)
}

func onRequest(ctx context.Context, request object.LLMRequest) filters.RequestFilterResult {
	var clusterType v1alpha1.ClusterType

	switch request.GetRequestType() {
	case object.RequestTypeChatCompletions, object.RequestTypeCompletions:
		clusterType = v1alpha1.ClusterType_LLM
	case object.RequestTypeImageGenerations:
		clusterType = v1alpha1.ClusterType_IMAGE_GENERATION
	}

	if clusterType == v1alpha1.ClusterType_CLUSTER_TYPE_UNSPECIFIED {
		return filters.NewFailed(fmt.Errorf("unknown request type %s, must be one of %v", request.GetRequestType(), []object.RequestType{object.RequestTypeChatCompletions, object.RequestTypeCompletions, object.RequestTypeImageGenerations}))
	}

	c, ok := findCluster(ctx, request, clusterType)
	if !ok {
		return filters.NewFailed(errors.New("cluster not found"))
	}

	// set destination cluster to context
	req := metadata.RequestMetadataFromCtx(ctx)
	req.SelectedCluster = mo.Some(c)

	return filters.NewOK()
}

func findRoute(ctx context.Context, llmRequest object.LLMRequest) (route.Route, string) {
	var r route.Route
	var clusterName string

	registryroute.ForeachRoute(func(item route.Route) bool {
		if cn, ok := item.Match(ctx, llmRequest); ok {
			clusterName = cn
			r = item

			return false
		}

		return true
	})

	return r, clusterName
}

func findCluster(ctx context.Context, llmRequest object.LLMRequest, expectedType v1alpha1.ClusterType) (clusters.Cluster, bool) {
	r, clusterName := findRoute(ctx, llmRequest)
	if r == nil {
		return nil, false
	}

	c, ok := registrycluster.FindClusterByName(clusterName)
	if !ok {
		return nil, false
	}

	if expectedType != c.Type() {
		return nil, false
	}

	return c, true
}
