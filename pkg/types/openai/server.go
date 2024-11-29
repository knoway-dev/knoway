package openai

import (
	"errors"
	"log/slog"
	"net/http"

	"knoway.dev/pkg/utils"
)

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

		var openAIError *ErrorResponse
		if errors.As(err, &openAIError) && openAIError != nil {
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

			return
		}

		slog.Error("failed to handle request, unhandled error occurred", "error", err)
		utils.WriteJSONForHTTP(http.StatusInternalServerError, NewErrorInternalError(), writer)
	}
}
