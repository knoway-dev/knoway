package openai

import (
	"errors"
	"log/slog"
	"net/http"

	"knoway.dev/pkg/metadata"
	"knoway.dev/pkg/utils"
)

var (
	SkipStreamResponse = errors.New("skip writing stream response") //nolint:errname,stylecheck
)

func ResponseHandler() func(resp any, err error, writer http.ResponseWriter, request *http.Request) {
	return func(resp any, err error, writer http.ResponseWriter, request *http.Request) {
		rMeta := metadata.RequestMetadataFromCtx(request.Context())

		if err == nil {
			if resp != nil {
				rMeta.StatusCode = http.StatusOK

				utils.WriteJSONForHTTP(http.StatusOK, resp, writer)
			}

			return
		}

		if errors.Is(err, SkipStreamResponse) {
			// NOTICE: special case where the response is already handled by the stream
			// handler as we assume the stream handler will handle the response as
			// status code 200 OK.
			rMeta.StatusCode = http.StatusOK

			return
		}

		openAIError := NewErrorFromLLMError(err)
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
