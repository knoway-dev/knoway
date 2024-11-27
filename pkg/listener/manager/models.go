package manager

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/samber/lo"

	v1alpha4 "knoway.dev/api/clusters/v1alpha1"
	"knoway.dev/pkg/object"

	"github.com/gorilla/mux"
	goopenai "github.com/sashabaranov/go-openai"
	"google.golang.org/protobuf/proto"

	"knoway.dev/api/listeners/v1alpha1"
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/listener"
	"knoway.dev/pkg/registry/cluster"
	"knoway.dev/pkg/registry/config"
	"knoway.dev/pkg/types/openai"
)

func NewOpenAIModelsListenerWithConfigs(cfg proto.Message) (listener.Listener, error) {
	c, ok := cfg.(*v1alpha1.ChatCompletionListener)
	if !ok {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}

	l := &OpenAIModelsListener{
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

type OpenAIModelsListener struct {
	cfg               *v1alpha1.ChatCompletionListener
	filters           []filters.RequestFilter
	listener.Listener // TODO: implement the interface
}

func (l *OpenAIModelsListener) listModels(writer http.ResponseWriter, request *http.Request) (any, error) {
	llmRequest := &object.BaseLLMRequest{}

	for _, f := range l.filters {
		fResult := f.OnCompletionRequest(request.Context(), llmRequest, request)
		if fResult.IsFailed() {
			return nil, fResult.Error
		}
	}

	clusters := cluster.ListModels()
	if config.HasAuthFilter() {
		clusters = lo.Filter(clusters, func(item *v1alpha4.Cluster, index int) bool {
			return llmRequest.CanAccessModel(item.GetName())
		})
	}

	sort.Slice(clusters, func(i, j int) bool {
		return strings.Compare(clusters[i].Name, clusters[j].Name) < 0
	})

	ms := ClustersToOpenAIModels(clusters)
	body := goopenai.ModelsList{
		Models: ms,
	}
	return body, nil
}

func (l *OpenAIModelsListener) RegisterRoutes(mux *mux.Router) error {
	mux.HandleFunc("/v1/models", openai.WrapHandlerForOpenAIError(l.listModels))

	return nil
}
