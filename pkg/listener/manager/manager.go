package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"google.golang.org/protobuf/proto"
	"io"
	"knoway.dev/api/listeners/v1alpha1"
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/listener"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/registry/cluster"
	"knoway.dev/pkg/registry/config"
	"knoway.dev/pkg/registry/route"
	route2 "knoway.dev/pkg/route"
	"net/http"
)

func NewWithConfigs(cfg proto.Message) (listener.Listener, error) {
	c, ok := cfg.(*v1alpha1.ChatCompletionListener)
	if !ok {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}
	l := &ListenerConnectionManager{
		cfg: c,
	}
	for _, fc := range c.Filters {
		f, err := config.NewRequestFilterWithConfig(fc.Name, fc.Config)
		if err != nil {
			return nil, err
		}
		l.filters = append(l.filters, f)
	}
	return l, nil
}

type ListenerConnectionManager struct {
	cfg               *v1alpha1.ChatCompletionListener
	filters           []filters.RequestFilter
	listener.Listener // todo implement the interface
}

type openAIChatCompletion struct {
	Model string `json:"model,omitempty"`
	// todo add more fields

	object.LLMRequest
}

func (l *ListenerConnectionManager) RegisterRoutes(mux *mux.Router) error {
	mux.HandleFunc("/api/v1/chat/completions", l.OnRequest)
	return nil
}

func (l *ListenerConnectionManager) UnmarshalLLMRequest(ctx context.Context, request *http.Request) (object.LLMRequest, error) {
	bs, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}
	var r = new(openAIChatCompletion)
	err = json.Unmarshal(bs, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (l *ListenerConnectionManager) OnRequest(writer http.ResponseWriter, request *http.Request) {
	req, err := l.UnmarshalLLMRequest(request.Context(), request)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	for _, f := range l.filters {
		if f.OnCompletionRequest != nil {
			res := f.OnCompletionRequest(request.Context(), req)
			if res.Type == filters.ListenerFilterResultTypeFailed {
				http.Error(writer, res.Error.Error(), http.StatusUnauthorized)
				return
			}
		}
	}
	var r route2.Route
	var clusterName string
	// do route
	route.ForeachRoute(func(item route2.Route) bool {
		if cn, ok := item.Match(request.Context(), req); ok {
			clusterName = cn
			r = item
			return false
		}
		return true
	})
	if r == nil {
		http.Error(writer, "no route matched", http.StatusNotFound)
		return
	}
	c, ok := cluster.FindClusterByName(clusterName)
	if !ok {
		http.Error(writer, fmt.Sprintf("cluster %s not found", clusterName), http.StatusNotFound)
		return
	}
	resp, err := c.DoUpstreamRequest(request.Context(), req)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	bs, err := json.Marshal(resp)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = writer.Write(bs)
	return
}
