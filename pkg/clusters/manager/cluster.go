package manager

import (
	"bufio"
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

var _ clusters.Cluster = (*clusterManager)(nil)

type clusterManager struct {
	cfg     *v1alpha1.Cluster
	filters []filters.ClusterFilter
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

func forEachRequestHandler(ctx context.Context, f []filters.ClusterFilter, req object.LLMRequest) (object.LLMRequest, error) {
	for _, filter := range f {
		marshaller, ok := filter.(filters.ClusterFilterRequestHandler)
		if ok {
			var err error

			req, err = marshaller.RequestPreflight(ctx, req)
			if err != nil {
				return nil, err
			}
		}
	}

	return req, nil
}

func forEachRequestMarshaller(ctx context.Context, f []filters.ClusterFilter, req object.LLMRequest) ([]byte, error) {
	var bs []byte

	for _, filter := range f {
		marshaller, ok := filter.(filters.ClusterFilterRequestMarshaller)
		if ok {
			var err error

			bs, err = marshaller.MarshalRequestBody(ctx, req, bs)
			if err != nil {
				return nil, err
			}
		}
	}

	if bs == nil {
		// default implementation, use json marshal req
		var err error

		bs, err = json.Marshal(req)
		if err != nil {
			return nil, err
		}
	}

	return bs, nil
}

func forEachResponseMarshaller(
	ctx context.Context,
	f []filters.ClusterFilter,
	req object.LLMRequest,
	rawResp *http.Response,
	reader *bufio.Reader,
	resp object.LLMResponse,
) (object.LLMResponse, error) {
	var err error

	for _, f := range lo.Reverse(f) {
		unmarshaller, ok := f.(filters.ClusterFilterResponseUnmarshaller)

		if ok {
			resp, err = unmarshaller.UnmarshalResponseBody(ctx, req, rawResp, reader, resp)
			if err != nil {
				return nil, err
			}
		}
	}

	return resp, nil
}

func composeRequestBody(ctx context.Context, f []filters.ClusterFilter, req object.LLMRequest) (io.ReadCloser, error) {
	var err error

	req, err = forEachRequestHandler(ctx, f, req)
	if err != nil {
		return nil, err
	}

	bs, err := forEachRequestMarshaller(ctx, f, req)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(bs)), nil
}

func composeLLMResponse(ctx context.Context, f []filters.ClusterFilter, req object.LLMRequest, rawResp *http.Response, reader *bufio.Reader) (object.LLMResponse, error) {
	var resp object.LLMResponse
	return forEachResponseMarshaller(ctx, f, req, rawResp, reader, resp)
}

func (m *clusterManager) DoUpstreamRequest(ctx context.Context, req object.LLMRequest) (object.LLMResponse, error) {
	body, err := composeRequestBody(ctx, m.LoadFilters(), req)
	if err != nil {
		return nil, err
	}

	// TODO: lb policy
	rawResp, buffer, err := doRequest(ctx, m.cfg.Upstream, body, RequestWithStream(req.IsStream()))
	if err != nil {
		return nil, err
	}

	return composeLLMResponse(ctx, m.LoadFilters(), req, rawResp, buffer)
}

type requestOptions struct {
	isStream bool
	endpoint string // TODO: implement
}

type requestCallOption func(*requestOptions)

func RequestWithStream(stream bool) requestCallOption {
	return func(opts *requestOptions) {
		opts.isStream = stream
	}
}

func RequestWithEndpoint(endpoint string) requestCallOption {
	return func(opts *requestOptions) {
		opts.endpoint = endpoint
	}
}

func doRequest(ctx context.Context, upstream *v1alpha1.Upstream, body io.ReadCloser, callOpts ...requestCallOption) (*http.Response, *bufio.Reader, error) {
	opts := &requestOptions{}

	for _, opt := range callOpts {
		opt(opts)
	}

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

	if opts.isStream {
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Connection", "keep-alive")
	}

	lo.ForEach(upstream.Headers, func(h *v1alpha1.Upstream_Header, _ int) {
		req.Header.Set(h.Key, h.Value)
	})

	// send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}

	return resp, bufio.NewReader(resp.Body), nil
}
