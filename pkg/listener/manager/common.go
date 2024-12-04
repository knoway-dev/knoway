package manager

import (
	"errors"
	"log/slog"
	"net/http"

	"knoway.dev/pkg/properties"
	"knoway.dev/pkg/types/openai"
	"knoway.dev/pkg/utils"

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

var (
	SkipResponse = errors.New("skip writing response") //nolint:errname,stylecheck
)

// WrapHandlerForOpenAIError
// todo added generic error handling, non-Hardcode openai error
func WrapHandlerForOpenAIError(fn func(writer http.ResponseWriter, request *http.Request) (any, error)) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		resp, err := fn(writer, request)
		if err == nil {
			if resp != nil {
				utils.WriteJSONForHTTP(http.StatusOK, resp, writer)
			}

			return
		}
		if errors.Is(err, SkipResponse) {
			return
		}

		openAIError := openai.NewErrorFromLLMError(err)
		if openAIError.FromUpstream {
			slog.Error("upstream returned an error",
				"status", openAIError.Status,
				"code", openAIError.ErrorBody.Code,
				"message", openAIError.ErrorBody.Message,
				"type", openAIError.ErrorBody.Type,
			)
		} else if openAIError.Status >= http.StatusInternalServerError {
			slog.Error("failed to handle request", "error", openAIError, "cause", openAIError.Cause)
		}

		utils.WriteJSONForHTTP(openAIError.Status, openAIError, writer)
	}
}
