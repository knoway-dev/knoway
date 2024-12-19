package chat

import (
	"net/http"
	"sort"
	"strings"

	"github.com/samber/lo"
	goopenai "github.com/sashabaranov/go-openai"

	v1alpha4 "knoway.dev/api/clusters/v1alpha1"
	"knoway.dev/pkg/filters/auth"
	"knoway.dev/pkg/metadata"
	"knoway.dev/pkg/registry/cluster"
)

func (l *OpenAIChatListener) listModels(writer http.ResponseWriter, request *http.Request) (any, error) {
	for _, f := range l.filters.OnRequestPreFilters() {
		fResult := f.OnRequestPre(request.Context(), request)
		if fResult.IsFailed() {
			return nil, fResult.Error
		}
	}

	var resp any
	var err error

	defer func() {
		for _, f := range l.filters.OnResponsePostFilters() {
			f.OnResponsePost(request.Context(), request, resp, err)
		}
	}()

	clusters := cluster.ListModels()

	// auth filters
	rMeta := metadata.RequestMetadataFromCtx(request.Context())

	if rMeta.EnabledAuthFilter {
		if rMeta.AuthInfo != nil {
			clusters = lo.Filter(clusters, func(item *v1alpha4.Cluster, index int) bool {
				return auth.CanAccessModel(item.GetName(), rMeta.AuthInfo.GetAllowModels(), rMeta.AuthInfo.GetDenyModels())
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
