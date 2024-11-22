package manager

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sashabaranov/go-openai"
	"google.golang.org/protobuf/proto"
	v1alpha4 "knoway.dev/api/clusters/v1alpha1"
	"knoway.dev/api/listeners/v1alpha1"
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/listener"
	"knoway.dev/pkg/registry/cluster"
	"knoway.dev/pkg/registry/config"
	"net/http"
	"sort"
	"strings"
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
	mux.HandleFunc("/v1/models", l.listModels)
	return nil
}

func (l *ListenerModelsManager) listModels(writer http.ResponseWriter, request *http.Request) {
	// todo add auth, get can access models to filter models
	clusters := cluster.ListModels()
	sort.Slice(clusters, func(i, j int) bool {
		return strings.Compare(clusters[i].Name, clusters[j].Name) < 0
	})

	ms := ClustersToOpenAIModels(clusters)
	body := openai.ModelsList{
		Models: ms,
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(body); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func ClustersToOpenAIModels(clusters []v1alpha4.Cluster) []openai.Model {
	res := make([]openai.Model, 0)
	for _, c := range clusters {
		res = append(res, ClusterToOpenAIModel(c))
	}

	return res
}

func ClusterToOpenAIModel(cluster v1alpha4.Cluster) openai.Model {
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
