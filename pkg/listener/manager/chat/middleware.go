package chat

import (
	"errors"
	"log/slog"
	"net/http"

	"knoway.dev/pkg/kcontext"

	"knoway.dev/pkg/types/openai"
	"knoway.dev/pkg/utils"
)

func ResponseHandler() func(resp any, err error, writer http.ResponseWriter, request *http.Request) {
	return func(resp any, err error, writer http.ResponseWriter, request *http.Request) {
		rMeta := kcontext.RequestMetadataFromCtx(request.Context())

		if err == nil {
			if resp != nil {
				rMeta.StatusCode = http.StatusOK

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

		rMeta.StatusCode = openAIError.Status
		rMeta.ErrorMessage = openAIError.Error()

		utils.WriteJSONForHTTP(openAIError.Status, openAIError, writer)
	}
}
