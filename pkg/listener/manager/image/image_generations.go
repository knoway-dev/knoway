package image

import (
	"log/slog"
	"net/http"

	"github.com/samber/lo"

	"knoway.dev/pkg/listener"
	"knoway.dev/pkg/metadata"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/types/openai"
)

func (l *OpenAIImageListener) unmarshalImageGenerationsRequestToImageGenerationRequest(request *http.Request) (object.LLMRequest, error) {
	llmRequest, err := openai.NewImageGenerationsRequest(request)
	if err != nil {
		return nil, err
	}

	if llmRequest.GetModel() == "" {
		return nil, openai.NewErrorMissingModel()
	}

	return llmRequest, nil
}

func (l *OpenAIImageListener) imageGeneration(writer http.ResponseWriter, request *http.Request) (any, error) {
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

	llmRequest, err := l.unmarshalImageGenerationsRequestToImageGenerationRequest(request)
	if err != nil {
		return nil, err
	}

	rMeta := metadata.RequestMetadataFromCtx(request.Context())
	rMeta.RequestModel = llmRequest.GetModel()

	for _, f := range l.filters.OnImageGenerationsRequestFilters() {
		fResult := f.OnImageGenerationsRequest(request.Context(), llmRequest, request)
		if fResult.IsFailed() {
			return nil, fResult.Error
		}
	}

	c, ok := listener.FindCluster(request.Context(), llmRequest)
	if !ok {
		return nil, openai.NewErrorModelNotFoundOrNotAccessible(llmRequest.GetModel())
	}

	resp, err = l.clusterDoImageGenerationRequest(request.Context(), c, writer, request, llmRequest)
	if !lo.IsNil(resp) {
		for _, f := range l.reversedFilters.OnImageGenerationsResponseFilters() {
			fResult := f.OnImageGenerationsResponse(request.Context(), llmRequest, resp)
			if fResult.IsFailed() {
				slog.Error("error occurred during invoking of OnImageGenerationsResponse filters", "error", fResult.Error)
			}
		}
	}

	rMeta.ResponseModel = llmRequest.GetModel()

	return resp, err
}
