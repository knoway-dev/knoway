package image

import (
	"context"
	"net/http"

	"github.com/samber/lo"
	"github.com/samber/mo"

	"knoway.dev/pkg/clusters"
	"knoway.dev/pkg/metadata"
	"knoway.dev/pkg/object"
)

func (l *OpenAIImageListener) clusterDoImageGenerationRequest(ctx context.Context, c clusters.Cluster, _ http.ResponseWriter, request *http.Request, llmRequest object.LLMRequest) (object.LLMResponse, error) {
	rMeta := metadata.RequestMetadataFromCtx(ctx)

	resp, err := c.DoUpstreamRequest(ctx, llmRequest)
	if err != nil {
		// Cluster will ensure that error will always be LLMError
		return nil, err
	}

	// For non-streaming responses, usage should be set here
	if !lo.IsNil(resp.GetUsage()) {
		rMeta.LLMUpstreamImagesUsage = mo.Some(lo.Must(object.AsLLMImagesUsage(resp.GetUsage())))
	}

	err = c.DoUpstreamResponseComplete(request.Context(), llmRequest, resp)
	if err != nil {
		// Cluster will ensure that error will always be LLMError
		return resp, err
	}

	if resp.GetError() != nil {
		return resp, resp.GetError()
	}

	return resp, nil
}
