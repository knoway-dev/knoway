package chat

import (
	"log/slog"
	"net/http"

	"knoway.dev/pkg/object"
	"knoway.dev/pkg/types/openai"
)

func (l *OpenAIChatListener) unmarshalChatCompletionsRequestToLLMRequest(request *http.Request) (object.LLMRequest, error) {
	llmRequest, err := openai.NewChatCompletionRequest(request)
	if err != nil {
		return nil, err
	}

	if llmRequest.GetModel() == "" {
		return nil, openai.NewErrorMissingModel()
	}

	return llmRequest, nil
}

func (l *OpenAIChatListener) onChatCompletionsRequestWithError(writer http.ResponseWriter, request *http.Request) (any, error) {
	llmRequest, err := l.unmarshalChatCompletionsRequestToLLMRequest(request)
	if err != nil {
		return nil, err
	}

	for _, f := range l.filters.OnCompletionRequestFilters() {
		fResult := f.OnCompletionRequest(request.Context(), llmRequest, request)
		if fResult.IsFailed() {
			return nil, fResult.Error
		}
	}

	c, ok := l.findCluster(request.Context(), llmRequest)
	if !ok {
		return nil, openai.NewErrorModelNotFoundOrNotAccessible(llmRequest.GetModel())
	}

	resp, err := l.clusterDoCompletionsRequest(c, writer, request, llmRequest)
	if !llmRequest.IsStream() {
		for _, f := range l.filters.OnCompletionResponseFilters() {
			fResult := f.OnCompletionResponse(request.Context(), llmRequest, resp)
			if fResult.IsFailed() {
				slog.Error("error occurred during invoking of OnCompletionResponse filters", "error", fResult.Error)
			}
		}
	}

	return resp, err
}
