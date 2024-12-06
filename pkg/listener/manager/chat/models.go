package chat

import (
	"net/http"
	"sort"
	"strings"

	"github.com/samber/lo"
	goopenai "github.com/sashabaranov/go-openai"

	v1alpha4 "knoway.dev/api/clusters/v1alpha1"
	"knoway.dev/pkg/filters/auth"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/properties"
	"knoway.dev/pkg/registry/cluster"
)

func (l *OpenAIChatListener) onListModelsRequestWithError(writer http.ResponseWriter, request *http.Request) (any, error) {
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
