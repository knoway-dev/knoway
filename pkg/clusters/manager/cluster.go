package manager

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/samber/lo"
	"google.golang.org/protobuf/proto"
	"io"
	"knoway.dev/api/clusters/v1alpha1"
	"knoway.dev/pkg/clusters"
	"knoway.dev/pkg/clusters/filters"
	"knoway.dev/pkg/object"
	registryfilters "knoway.dev/pkg/registry/config"
	"net/http"
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
	case v1alpha1.LoadBalancePolicy_CUSTOM, v1alpha1.LoadBalancePolicy_LOAD_BALANCE_POLICY_UNSPECIFIED:
		//if _, ok := lo.Find(fs, func(f filters.ClusterFilter) bool {
		//	return f.SelectEndpoint != nil
		//}); !ok {
		//	return nil, fmt.Errorf("custom load balance policy must be implemented")
		//}
		// todo remove test
	default:
		// if use internal lb, filter must NOT implement SelectEndpoint
		if lo.SomeBy(fs, func(f filters.ClusterFilter) bool {
			return f.SelectEndpoint != nil
		}) {
			return nil, fmt.Errorf("internal load balance policy must NOT be implemented")
		}
	}

	return &clusterManager{
		cfg:     conf,
		filters: fs,
	}, nil
}

func (m *clusterManager) DoUpstreamRequest(ctx context.Context, req object.LLMRequest) (object.LLMResponse, error) {
	var bs []byte
	for _, f := range m.filters {
		if f.MarshalRequestBody != nil {
			var err error
			bs, err = f.MarshalRequestBody(ctx, req, bs)
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

	// todo lb policy
	body, err := doRequest(ctx, m.cfg.Upstream, "", reader)
	if err != nil {
		return nil, err
	}
	var resp object.LLMResponse
	for _, f := range m.filters {
		if f.UnmarshalResponseBody != nil {
			bs, err := io.ReadAll(body)
			if err != nil {
				if !errors.Is(err, io.EOF) {
					return nil, err
				} else {
					break
				}
			}
			if resp, err = f.UnmarshalResponseBody(ctx, bs, resp); err != nil {
				return nil, err
			}
		}
	}

	return resp, nil
}

func doRequest(ctx context.Context, upstream *v1alpha1.Upstream, endpoint string, body io.ReadCloser) (io.ReadCloser, error) {
	// todo endpoint
	// todo stream
	// send request
	var method string
	switch upstream.Method {
	case v1alpha1.Upstream_GET:
		method = http.MethodGet
	case v1alpha1.Upstream_POST:
		method = http.MethodPost
	default:
		return nil, fmt.Errorf("unsupported method %s", upstream.Method)
	}
	req, err := http.NewRequest(method, upstream.Url, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	lo.ForEach(upstream.Headers, func(h *v1alpha1.Upstream_Header, _ int) {
		req.Header.Set(h.Key, h.Value)
	})
	// send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, err
}
