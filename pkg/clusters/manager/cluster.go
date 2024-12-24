package manager

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/samber/lo"
	"github.com/samber/mo"
	"google.golang.org/protobuf/proto"

	"knoway.dev/api/clusters/v1alpha1"
	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/clusters"
	"knoway.dev/pkg/clusters/filters"
	"knoway.dev/pkg/metadata"
	"knoway.dev/pkg/object"
	registryfilters "knoway.dev/pkg/registry/config"
)

var _ clusters.Cluster = (*clusterManager)(nil)

type clusterManager struct {
	cluster *v1alpha1.Cluster
	filters []filters.ClusterFilter
}

func NewWithConfigs(clusterProtoMsg proto.Message, lifecycle bootkit.LifeCycle) (clusters.Cluster, error) {
	var conf *v1alpha1.Cluster
	var clusterFilters []filters.ClusterFilter

	if cluster, ok := clusterProtoMsg.(*v1alpha1.Cluster); !ok {
		return nil, fmt.Errorf("invalid config type %T", cluster)
	} else {
		conf = cluster

		for _, fc := range cluster.GetFilters() {
			if f, err := registryfilters.NewClusterFilterWithConfig(fc.GetName(), fc.GetConfig(), lifecycle); err != nil {
				return nil, err
			} else {
				clusterFilters = append(clusterFilters, f)
			}
		}
	}

	// check lb
	switch conf.GetLoadBalancePolicy() {
	case v1alpha1.LoadBalancePolicy_IP_HASH:
		// TODO: implement
	case v1alpha1.LoadBalancePolicy_LEAST_CONNECTION:
		// TODO: implement
	case v1alpha1.LoadBalancePolicy_ROUND_ROBIN:
		// TODO: implement
	case v1alpha1.LoadBalancePolicy_CUSTOM, v1alpha1.LoadBalancePolicy_LOAD_BALANCE_POLICY_UNSPECIFIED:
		_, ok := lo.Find(clusterFilters, func(f filters.ClusterFilter) bool {
			selector, ok := f.(filters.ClusterFilterEndpointSelector)
			return ok && selector != nil
		})
		if !ok {
			return nil, errors.New("custom load balance policy must be implemented")
		}
	default:
		// if use internal lb, filter must NOT implement SelectEndpoint
		if lo.SomeBy(clusterFilters, func(f filters.ClusterFilter) bool {
			selector, ok := f.(filters.ClusterFilterEndpointSelector)
			return ok && selector != nil
		}) {
			return nil, errors.New("internal load balance policy must NOT be implemented")
		}
	}

	return &clusterManager{
		cluster: conf,
		filters: append(clusterFilters, registryfilters.ClusterDefaultFilters(lifecycle)...),
	}, nil
}

func (m *clusterManager) DoUpstreamRequest(ctx context.Context, llmReq object.LLMRequest) (object.LLMResponse, error) {
	var err error

	rMeta := metadata.RequestMetadataFromCtx(ctx)
	rMeta.UpstreamProvider = m.cluster.GetProvider()

	for _, filter := range m.filters {
		f, ok := filter.(filters.ClusterFilterRequestModifier)
		if !ok {
			continue
		}

		llmReq, err = f.RequestModifier(ctx, m.cluster, llmReq)
		if err != nil {
			return nil, object.LLMErrorOrInternalError(err)
		}
	}

	rMeta.UpstreamRequestModel = llmReq.GetModel()

	var req *http.Request

	for _, filter := range m.filters {
		f, ok := filter.(filters.ClusterFilterUpstreamRequestMarshaller)
		if !ok {
			continue
		}

		req, err = f.MarshalUpstreamRequest(ctx, m.cluster, llmReq, req)
		if err != nil {
			return nil, object.LLMErrorOrInternalError(err)
		}
	}
	if lo.IsNil(req) {
		panic("ClusterFilterUpstreamRequestMarshaller iterated, but returned nil request or no filters found")
	}

	rMeta.UpstreamRequestAt = time.Now()

	// TODO: lb policy
	// TODO: body close
	rawResp, buffer, err := doRequest(req) //nolint:bodyclose
	if err != nil {
		return nil, object.NewErrorBadGateway(err)
	}

	// err != nil means the connection is not possible to establish
	// or find it's way to the destination, or upstream timeout
	rMeta.UpstreamRespondAt = time.Now()

	var llmResp object.LLMResponse

	for _, filter := range m.filters {
		f, ok := filter.(filters.ClusterFilterResponseUnmarshaller)
		if !ok {
			continue
		}

		var err error

		llmResp, err = f.UnmarshalResponseBody(ctx, llmReq, rawResp, buffer, llmResp)
		if err != nil {
			return nil, object.LLMErrorOrInternalError(err)
		}
	}
	if lo.IsNil(llmResp) {
		panic("ClusterFilterResponseUnmarshaller iterated, but returned nil response or no filters found")
	}

	rMeta.UpstreamResponseModel = llmResp.GetModel()

	for _, filter := range m.filters {
		f, ok := filter.(filters.ClusterFilterResponseModifier)
		if !ok {
			continue
		}

		llmResp, err = f.ResponseModifier(ctx, m.cluster, llmReq, llmResp)
		if err != nil {
			return nil, object.LLMErrorOrInternalError(err)
		}
	}

	rMeta.UpstreamResponseStatusCode = rawResp.StatusCode
	rMeta.UpstreamResponseHeader = mo.Some(rawResp.Header)

	if !lo.IsNil(llmResp.GetError()) {
		rMeta.UpstreamResponseErrorMessage = llmResp.GetError().Error()
	}

	return llmResp, nil
}

func (m *clusterManager) DoUpstreamResponseComplete(ctx context.Context, req object.LLMRequest, res object.LLMResponse) error {
	for _, filter := range m.filters {
		f, ok := filter.(filters.ClusterFilterResponseComplete)
		if !ok {
			continue
		}

		err := f.ResponseComplete(ctx, req, res)
		if err != nil {
			return object.LLMErrorOrInternalError(err)
		}
	}

	return nil
}

func doRequest(req *http.Request) (*http.Response, *bufio.Reader, error) {
	// send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}

	return resp, bufio.NewReader(resp.Body), nil
}
