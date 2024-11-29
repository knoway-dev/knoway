package manager

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/samber/lo"

	"github.com/gorilla/mux"
	goopenai "github.com/sashabaranov/go-openai"
	"google.golang.org/protobuf/proto"

	v1alpha4 "knoway.dev/api/clusters/v1alpha1"
	"knoway.dev/api/listeners/v1alpha1"
	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/filters/auth"
	"knoway.dev/pkg/listener"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/properties"
	"knoway.dev/pkg/registry/cluster"
	"knoway.dev/pkg/registry/config"
	"knoway.dev/pkg/types/openai"
)

func NewOpenAIModelsListenerWithConfigs(cfg proto.Message, lifecycle bootkit.LifeCycle) (listener.Listener, error) {
	c, ok := cfg.(*v1alpha1.ChatCompletionListener)
	if !ok {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}

	l := &OpenAIModelsListener{
		cfg: c,
	}

	for _, fc := range c.GetFilters() {
		f, err := config.NewRequestFilterWithConfig(fc.GetName(), fc.GetConfig(), lifecycle)
		if err != nil {
			return nil, err
		}

		l.filters = append(l.filters, f)
	}

	return l, nil
}

type OpenAIModelsListener struct {
	cfg               *v1alpha1.ChatCompletionListener
	filters           filters.RequestFilters
	listener.Listener // TODO: implement the interface
}

func (l *OpenAIModelsListener) listModels(writer http.ResponseWriter, request *http.Request) (any, error) {
	llmRequest := &object.BaseLLMRequest{}

	for _, f := range l.filters.OnCompletionRequestFilters() {
		fResult := f.OnCompletionRequest(request.Context(), llmRequest, request)
		if fResult.IsFailed() {
			return nil, fResult.Error
		}
	}

	clusters := cluster.ListModels()

	// auth filters
	if properties.EnabledAuthFilterFromCtx(request.Context()) {
		if authInfo, ok := properties.GetAuthInfoFromCtx(request.Context()); ok {
			allowModels := authInfo.GetAllowModels()
			clusters = lo.Filter(clusters, func(item *v1alpha4.Cluster, index int) bool {
				return auth.CanAccessModel(allowModels, item.GetName())
			})
		}
	}

	sort.Slice(clusters, func(i, j int) bool {
		return strings.Compare(clusters[i].GetName(), clusters[j].GetName()) < 0
	})

	ms := ClustersToOpenAIModels(clusters)
	body := goopenai.ModelsList{
		Models: ms,
	}

	return body, nil
}

func (l *OpenAIModelsListener) RegisterRoutes(mux *mux.Router) error {
	mux.HandleFunc("/v1/models", WrapRequest(openai.WrapHandlerForOpenAIError(l.listModels)))

	return nil
}
