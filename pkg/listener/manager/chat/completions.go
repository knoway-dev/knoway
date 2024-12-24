package chat

import (
	"log/slog"
	"net/http"

	"github.com/samber/lo"

	"knoway.dev/pkg/metadata"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/types/openai"
)

func (l *OpenAIChatListener) unmarshalCompletionsRequestToLLMRequest(request *http.Request) (object.LLMRequest, error) {
	llmRequest, err := openai.NewCompletionsRequest(request)
	if err != nil {
		return nil, err
	}

	if llmRequest.GetModel() == "" {
		return nil, openai.NewErrorMissingModel()
	}

	rMeta := metadata.RequestMetadataFromCtx(request.Context())
	rMeta.RequestModel = llmRequest.GetModel()

	return llmRequest, nil
}

func (l *OpenAIChatListener) completions(writer http.ResponseWriter, request *http.Request) (any, error) {
	for _, f := range l.filters.OnRequestPreFilters() {
		fResult := f.OnRequestPre(request.Context(), request)
		if fResult.IsFailed() {
			return nil, fResult.Error
		}
	}

	var resp object.LLMResponse
	var err error

	defer func() {
		for _, f := range l.filters.OnResponsePostFilters() {
			f.OnResponsePost(request.Context(), request, resp, err)
		}
	}()

	llmRequest, err := l.unmarshalCompletionsRequestToLLMRequest(request)
	if err != nil {
		return nil, err
	}

	rMeta := metadata.RequestMetadataFromCtx(request.Context())
	rMeta.RequestModel = llmRequest.GetModel()

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

	resp, err = l.clusterDoCompletionsRequest(request.Context(), c, writer, request, llmRequest)
	if !llmRequest.IsStream() && !lo.IsNil(resp) {
		for _, f := range l.filters.OnCompletionResponseFilters() {
			fResult := f.OnCompletionResponse(request.Context(), llmRequest, resp)
			if fResult.IsFailed() {
				slog.Error("error occurred during invoking of OnCompletionResponse filters", "error", fResult.Error)
			}
		}
	}

	rMeta.ResponseModel = llmRequest.GetModel()

	return resp, err
}
