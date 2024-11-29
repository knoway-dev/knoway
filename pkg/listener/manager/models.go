package manager

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/samber/lo"

	v1alpha4 "knoway.dev/api/clusters/v1alpha1"
	"knoway.dev/pkg/filters/auth"

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

	for _, fc := range c.GetFilters() {
		f, err := config.NewRequestFilterWithConfig(fc.GetName(), fc.GetConfig())
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

	ctx := request.Context()
	for _, f := range l.filters {
		fResult := f.OnCompletionRequest(ctx, llmRequest, request)
		if fResult.IsFailed() {
			return nil, fResult.Error
		}
	}

	clusters := cluster.ListModels()

	// auth filters
	if auth.EnabledAuthFilterFromCtx(ctx) {
		if authInfo, ok := auth.GetAuthInfoFromCtx(ctx); ok {
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
