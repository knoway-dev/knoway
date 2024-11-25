package manager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/samber/lo"
	"google.golang.org/protobuf/proto"

	"knoway.dev/api/clusters/v1alpha1"
	"knoway.dev/pkg/clusters"
	"knoway.dev/pkg/clusters/filters"
	"knoway.dev/pkg/object"
	registryfilters "knoway.dev/pkg/registry/config"
)

type clusterManager struct {
	cfg     *v1alpha1.Cluster
	filters []filters.ClusterFilter
	clusters.Cluster
}

func NewWithConfigs(cfg proto.Message) (clusters.Cluster, error) {
	var conf *v1alpha1.Cluster
	var fs []filters.ClusterFilter

	if cfg, ok := cfg.(*v1alpha1.Cluster); !ok {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	} else {
		conf = cfg

		for _, fc := range cfg.Filters {
			if f, err := registryfilters.NewClusterFilterWithConfig(fc.Name, fc.Config); err != nil {
				return nil, err
			} else {
				fs = append(fs, f)
			}
		}
	}

	// check lb
	switch conf.LoadBalancePolicy {
	case v1alpha1.LoadBalancePolicy_IP_HASH:
		// TODO: implement
	case v1alpha1.LoadBalancePolicy_LEAST_CONNECTION:
		// TODO: implement
	case v1alpha1.LoadBalancePolicy_ROUND_ROBIN:
		// TODO: implement
	case v1alpha1.LoadBalancePolicy_CUSTOM, v1alpha1.LoadBalancePolicy_LOAD_BALANCE_POLICY_UNSPECIFIED:
		if _, ok := lo.Find(fs, func(f filters.ClusterFilter) bool {
			selector, ok := f.(filters.ClusterFilterEndpointSelector)
			return ok && selector != nil
		}); !ok {
			return nil, fmt.Errorf("custom load balance policy must be implemented")
		}
	default:
		// if use internal lb, filter must NOT implement SelectEndpoint
		if lo.SomeBy(fs, func(f filters.ClusterFilter) bool {
			selector, ok := f.(filters.ClusterFilterEndpointSelector)
			return ok && selector != nil
		}) {
			return nil, fmt.Errorf("internal load balance policy must NOT be implemented")
		}
	}

	return &clusterManager{
		cfg:     conf,
		filters: fs,
	}, nil
}

func (m *clusterManager) LoadFilters() []filters.ClusterFilter {
	res := m.filters
	return append(res, registryfilters.ClusterDefaultFilters()...)
}

func (m *clusterManager) DoUpstreamRequest(ctx context.Context, req object.LLMRequest) (object.LLMResponse, error) {
	var bs []byte

	for _, f := range m.LoadFilters() {
		marshaller, ok := f.(filters.ClusterFilterRequestHandler)
		if ok {
			var err error

			req, err = marshaller.RequestPreflight(ctx, req)
			if err != nil {
				return nil, err
			}
		}
	}

	for _, f := range m.LoadFilters() {
		marshaller, ok := f.(filters.ClusterFilterRequestMarshaller)
		if ok {
			var err error

			bs, err = marshaller.MarshalRequestBody(ctx, req, bs)
			if err != nil {
				return nil, err
			}
		}
	}

	var reader io.ReadCloser

	if bs == nil {
		// default implementation, use json marshal req
		var err error

		bs, err = json.Marshal(req)
		if err != nil {
			return nil, err
		}
	}

	reader = io.NopCloser(bytes.NewReader(bs))
	// TODO: lb policy
	rawResp, buffer, err := doRequest(ctx, m.cfg.Upstream, "", reader)
	if err != nil {
		return nil, err
	}

	var resp object.LLMResponse

	for _, f := range lo.Reverse(m.LoadFilters()) {
		unmarshaller, ok := f.(filters.ClusterFilterResponseUnmarshaller)

		if ok {
			resp, err = unmarshaller.UnmarshalResponseBody(ctx, req, rawResp, buffer, resp)
			if err != nil {
				return nil, err
			}
		}
	}

	return resp, nil
}

func doRequest(ctx context.Context, upstream *v1alpha1.Upstream, _ string, body io.ReadCloser) (*http.Response, *bytes.Buffer, error) {
	// TODO: endpoint
	// TODO: stream
	// TODO: request
	var method string

	switch upstream.Method {
	case v1alpha1.Upstream_GET:
		method = http.MethodGet
	case v1alpha1.Upstream_POST:
		method = http.MethodPost
	case v1alpha1.Upstream_METHOD_UNSPECIFIED:
		return nil, nil, fmt.Errorf("unsupported method %s", upstream.Method)
	default:
		return nil, nil, fmt.Errorf("unsupported method %s", upstream.Method)
	}

	req, err := http.NewRequest(method, upstream.Url, body)
	if err != nil {
		return nil, nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	lo.ForEach(upstream.Headers, func(h *v1alpha1.Upstream_Header, _ int) {
		req.Header.Set(h.Key, h.Value)
	})

	// send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		// REVIEW: should we handle error here?
		// TODO: logging error
		_ = resp.Body.Close()
	}()

	buffer := new(bytes.Buffer)

	_, err = buffer.ReadFrom(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	return resp, buffer, nil
}
