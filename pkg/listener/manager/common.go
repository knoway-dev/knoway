package manager

import (
	"net/http"

	goopenai "github.com/sashabaranov/go-openai"

	"knoway.dev/pkg/context"

	v1alpha4 "knoway.dev/api/clusters/v1alpha1"
)

func ClustersToOpenAIModels(clusters []*v1alpha4.Cluster) []goopenai.Model {
	res := make([]goopenai.Model, 0)
	for _, c := range clusters {
		res = append(res, ClusterToOpenAIModel(c))
	}

	return res
}

func ClusterToOpenAIModel(cluster *v1alpha4.Cluster) goopenai.Model {
	// from https://platform.openai.com/docs/api-reference/models/object
	return goopenai.Model{
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

func WrapRequest(fn func(writer http.ResponseWriter, request *http.Request)) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		fn(writer, request.WithContext(context.InitProperties(request.Context())))
	}
}
