package chat

import (
	"log/slog"
	"net/http"

	registrycluster "knoway.dev/pkg/registry/cluster"

	"github.com/samber/lo"

	"knoway.dev/pkg/metadata"
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

func (l *OpenAIChatListener) chatCompletions(writer http.ResponseWriter, request *http.Request) (any, error) {
	for _, f := range l.filters.OnRequestPreFilters() {
		fResult := f.OnRequestPre(request.Context(), request)
		if fResult.IsFailed() {
			return nil, fResult.Error
		}
	}

	var resp object.LLMResponse
	var err error

	defer func() {
		for _, f := range l.reversedFilters.OnResponsePostFilters() {
			f.OnResponsePost(request.Context(), request, resp, err)
		}
	}()

	llmRequest, err := l.unmarshalChatCompletionsRequestToLLMRequest(request)
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

	c, ok := registrycluster.FindClusterByName(rMeta.DestinationCluster)
	if !ok {
		return nil, openai.NewErrorModelNotFoundOrNotAccessible(llmRequest.GetModel())
	}

	resp, err = l.clusterDoCompletionsRequest(request.Context(), c, writer, request, llmRequest)
	if !llmRequest.IsStream() && !lo.IsNil(resp) {
		for _, f := range l.reversedFilters.OnCompletionResponseFilters() {
			fResult := f.OnCompletionResponse(request.Context(), llmRequest, resp)
			if fResult.IsFailed() {
				slog.Error("error occurred during invoking of OnCompletionResponse filters", "error", fResult.Error)
			}
		}
	}

	rMeta.ResponseModel = llmRequest.GetModel()

	return resp, err
}
