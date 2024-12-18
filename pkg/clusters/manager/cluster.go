package manager

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/samber/lo"
	"google.golang.org/protobuf/proto"

	"knoway.dev/pkg/metadata"

	"knoway.dev/api/clusters/v1alpha1"
	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/clusters"
	"knoway.dev/pkg/clusters/filters"
	"knoway.dev/pkg/object"
	registryfilters "knoway.dev/pkg/registry/config"
)

var _ clusters.Cluster = (*clusterManager)(nil)

type clusterManager struct {
	cluster *v1alpha1.Cluster
	filters filters.ClusterFilters
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

func composeLLMRequestBody(ctx context.Context, f filters.ClusterFilters, cluster *v1alpha1.Cluster, llmReq object.LLMRequest) (*http.Request, error) {
	var err error
	var req *http.Request

	llmReq, err = f.ForEachRequestModifier(ctx, cluster, llmReq)
	if err != nil {
		return nil, err
	}

	req, err = f.ForEachUpstreamRequestMarshaller(ctx, cluster, llmReq, req)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func composeLLMResponseFromBody(ctx context.Context, f filters.ClusterFilters, cluster *v1alpha1.Cluster, req object.LLMRequest, rawResp *http.Response, reader *bufio.Reader) (object.LLMResponse, error) {
	var err error
	var resp object.LLMResponse

	resp, err = f.ForEachResponseUnmarshaller(ctx, req, rawResp, reader, resp)
	if err != nil {
		return nil, err
	}

	resp, err = f.ForEachResponseModifier(ctx, cluster, req, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (m *clusterManager) DoUpstreamRequest(ctx context.Context, llmReq object.LLMRequest) (object.LLMResponse, error) {
	req, err := composeLLMRequestBody(ctx, m.filters, m.cluster, llmReq)
	if err != nil {
		return nil, err
	}

	metadata.RequestMetadataFromCtx(ctx).UpstreamRequestAt = time.Now()

	// TODO: lb policy
	// TODO: body close
	rawResp, buffer, err := doRequest(req) //nolint:bodyclose
	if err != nil {
		return nil, err
	}

	return composeLLMResponseFromBody(ctx, m.filters, m.cluster, llmReq, rawResp, buffer)
}

func (m *clusterManager) DoUpstreamResponseComplete(ctx context.Context, req object.LLMRequest, res object.LLMResponse) error {
	return m.filters.ForEachResponseComplete(ctx, req, res)
}

func doRequest(req *http.Request) (*http.Response, *bufio.Reader, error) {
	// send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}

	return resp, bufio.NewReader(resp.Body), nil
}
