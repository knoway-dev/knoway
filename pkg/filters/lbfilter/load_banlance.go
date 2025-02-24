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
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/metadata"
	"knoway.dev/pkg/object"
	registryroute "knoway.dev/pkg/registry/route"
)

func NewWithConfig(_ *anypb.Any, _ bootkit.LifeCycle) (filters.RequestFilter, error) {
	return &LBFilter{}, nil
}

type LBFilter struct {
	filters.IsRequestFilter
}

var _ filters.RequestFilter = (*LBFilter)(nil)
var _ filters.OnCompletionRequestFilter = (*LBFilter)(nil)

func (l *LBFilter) OnCompletionRequest(ctx context.Context, request object.LLMRequest, sourceHTTPRequest *http.Request) filters.RequestFilterResult {
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

	c, ok := registryroute.FindCluster(ctx, request, clusterType)
	if !ok {
		return filters.NewFailed(errors.New("cluster not found"))
	}

	// set destination cluster to context
	req := metadata.RequestMetadataFromCtx(ctx)
	req.SelectedCluster = mo.Some(c)

	return filters.NewOK()
}
