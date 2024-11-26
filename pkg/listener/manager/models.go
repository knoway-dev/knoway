package manager

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/samber/lo"

	v1alpha12 "knoway.dev/api/service/v1alpha1"
	"knoway.dev/pkg/filters/auth"

	"github.com/gorilla/mux"
	"github.com/sashabaranov/go-openai"
	"google.golang.org/protobuf/proto"

	v1alpha4 "knoway.dev/api/clusters/v1alpha1"
	"knoway.dev/api/listeners/v1alpha1"
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/listener"
	"knoway.dev/pkg/registry/cluster"
	"knoway.dev/pkg/registry/config"
)

func NewModelsManagerWithConfigs(cfg proto.Message) (listener.Listener, error) {
	c, ok := cfg.(*v1alpha1.ChatCompletionListener)
	if !ok {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}

	l := &ListenerModelsManager{
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

type ListenerModelsManager struct {
	cfg               *v1alpha1.ChatCompletionListener
	filters           []filters.RequestFilter
	listener.Listener // todo implement the interface
}

func (l *ListenerModelsManager) RegisterRoutes(mux *mux.Router) error {
	mux.HandleFunc("/v1/models", wrapErrorHandler(l.listModels))
	return nil
}

func (l *ListenerModelsManager) listModels(writer http.ResponseWriter, request *http.Request) (any, error) {
	var err error
	authResp := &v1alpha12.APIKeyAuthResponse{}
	for _, f := range l.filters {
		authFilter, ok := f.(*auth.AuthFilter)
		if ok {
			if authResp, err = Auth(authFilter, request, AuthZOption{}); err != nil {
				return nil, err
			}
		}
	}

	clusters := cluster.ListModels()
	clusters = lo.Filter(clusters, func(item *v1alpha4.Cluster, index int) bool {
		return authResp.CanAccessModel(item.GetName())
	})
	sort.Slice(clusters, func(i, j int) bool {
		return strings.Compare(clusters[i].Name, clusters[j].Name) < 0
	})

	ms := ClustersToOpenAIModels(clusters)
	body := openai.ModelsList{
		Models: ms,
	}
	return body, nil
}

func ClustersToOpenAIModels(clusters []*v1alpha4.Cluster) []openai.Model {
	res := make([]openai.Model, 0)
	for _, c := range clusters {
		res = append(res, ClusterToOpenAIModel(c))
	}

	return res
}

func ClusterToOpenAIModel(cluster *v1alpha4.Cluster) openai.Model {
	// from https://platform.openai.com/docs/api-reference/models/object
	return openai.Model{
		CreatedAt: cluster.Created,
		ID:        cluster.Name,
		// The object type, which is always "model".
		Object:  "model",
		OwnedBy: cluster.Provider,
		// todo ??
		Permission: nil,
		Root:       "",
		Parent:     "",
	}
}
