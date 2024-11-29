package manager

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/samber/lo"
	"google.golang.org/protobuf/proto"

	"knoway.dev/api/clusters/v1alpha1"
	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/clusters"
	"knoway.dev/pkg/clusters/filters"
	"knoway.dev/pkg/object"
	registryfilters "knoway.dev/pkg/registry/config"
)

var _ clusters.Cluster = (*clusterManager)(nil)

type clusterManager struct {
	cfg     *v1alpha1.Cluster
	filters filters.ClusterFilters
}

func NewWithConfigs(cfg proto.Message, lifecycle bootkit.LifeCycle) (clusters.Cluster, error) {
	var conf *v1alpha1.Cluster
	var clusterFilters []filters.ClusterFilter

	if cfg, ok := cfg.(*v1alpha1.Cluster); !ok {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	} else {
		conf = cfg

		for _, fc := range cfg.GetFilters() {
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
		cfg:     conf,
		filters: append(clusterFilters, registryfilters.ClusterDefaultFilters(lifecycle)...),
	}, nil
}

func (m *clusterManager) LoadFilters() []filters.ClusterFilter {
	return m.filters
}

func forEachRequestHandler(ctx context.Context, f filters.ClusterFilters, req object.LLMRequest) (object.LLMRequest, error) {
	for _, f := range f.OnRequestHandlers() {
		var err error

		req, err = f.RequestPreflight(ctx, req)
		if err != nil {
			return nil, err
		}
	}

	return req, nil
}

func forEachRequestMarshaller(ctx context.Context, f filters.ClusterFilters, req object.LLMRequest) ([]byte, error) {
	var bs []byte

	for _, f := range f.OnRequestMarshallers() {
		var err error

		bs, err = f.MarshalRequestBody(ctx, req, bs)
		if err != nil {
			return nil, err
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
	f filters.ClusterFilters,
	req object.LLMRequest,
	rawResp *http.Response,
	reader *bufio.Reader,
	resp object.LLMResponse,
) (object.LLMResponse, error) {
	var err error

	for _, f := range lo.Reverse(f).OnResponseUnmarshallers() {
		resp, err = f.UnmarshalResponseBody(ctx, req, rawResp, reader, resp)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func composeRequestBody(ctx context.Context, f filters.ClusterFilters, req object.LLMRequest) (io.Reader, error) {
	var err error

	req, err = forEachRequestHandler(ctx, f, req)
	if err != nil {
		return nil, err
	}

	bs, err := forEachRequestMarshaller(ctx, f, req)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(bs), nil
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
	// TODO: body close
	rawResp, buffer, err := doRequest(ctx, m.cfg.GetUpstream(), body, RequestWithStream(req.IsStream())) //nolint:bodyclose
	if err != nil {
		return nil, err
	}

	return composeLLMResponse(ctx, m.LoadFilters(), req, rawResp, buffer)
}

func (m *clusterManager) DoUpstreamResponseComplete(ctx context.Context, req object.LLMRequest, res object.LLMResponse) error {
	fs := filters.ClusterFilters(m.LoadFilters())

	for _, f := range fs.OnResponseHandlers() {
		err := f.ResponseComplete(ctx, req, res)
		if err != nil {
			return err
		}
	}

	return nil
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

func doRequest(ctx context.Context, upstream *v1alpha1.Upstream, body io.Reader, callOpts ...requestCallOption) (*http.Response, *bufio.Reader, error) {
	opts := &requestOptions{}

	for _, opt := range callOpts {
		opt(opts)
	}

	// TODO: endpoint
	// TODO: stream
	// TODO: request
	var method string

	switch upstream.GetMethod() {
	case v1alpha1.Upstream_GET:
		method = http.MethodGet
	case v1alpha1.Upstream_POST:
		method = http.MethodPost
	case v1alpha1.Upstream_METHOD_UNSPECIFIED:
		return nil, nil, fmt.Errorf("unsupported method %s", upstream.GetMethod())
	default:
		return nil, nil, fmt.Errorf("unsupported method %s", upstream.GetMethod())
	}

	req, err := http.NewRequest(method, upstream.GetUrl(), body)
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

	lo.ForEach(upstream.GetHeaders(), func(h *v1alpha1.Upstream_Header, _ int) {
		req.Header.Set(h.GetKey(), h.GetValue())
	})

	// send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}

	return resp, bufio.NewReader(resp.Body), nil
}
