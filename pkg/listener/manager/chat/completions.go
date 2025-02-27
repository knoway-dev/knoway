package chat

import (
	"net/http"

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
