package manager

import (
	"net/http"

	"knoway.dev/pkg/properties"

	goopenai "github.com/sashabaranov/go-openai"

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
		CreatedAt: cluster.GetCreated(),
		ID:        cluster.GetName(),
		// The object type, which is always "model".
		Object:  "model",
		OwnedBy: cluster.GetProvider(),
		// todo ??
		Permission: nil,
		Root:       "",
		Parent:     "",
	}
}

func WrapRequest(fn func(writer http.ResponseWriter, request *http.Request)) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		fn(writer, request.WithContext(properties.NewPropertiesContext(request.Context())))
	}
}
