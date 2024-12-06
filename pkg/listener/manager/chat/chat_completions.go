package chat

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"knoway.dev/pkg/object"
	"knoway.dev/pkg/registry/cluster"
	registryroute "knoway.dev/pkg/registry/route"
	"knoway.dev/pkg/route"
	"knoway.dev/pkg/types/openai"
	"knoway.dev/pkg/utils"
)

func (l *OpenAIChatListener) pipeChatCompletionsStream(ctx context.Context, request object.LLMRequest, resp object.LLMResponse, writer http.ResponseWriter) error {
	streamResp, ok := resp.(object.LLMStreamResponse)
	if !ok {
		return openai.NewErrorInternalError().WithCausef("failed to cast %T to object.LLMStreamResponse", resp)
	}

	utils.WriteEventStreamHeadersForHTTP(writer)

	marshalToWriter := func(chunk object.LLMChunkResponse) error {
		event, err := chunk.ToServerSentEvent()
		if err != nil {
			slog.Error("chunk error", "error", err)
			return openai.NewErrorInternalError().WithCause(err)
		}

		err = event.MarshalTo(writer)
		if err != nil {
			slog.Error("chunk error", "error", err)
			return err
		}
		return nil
	}

	for {
		chunk, err := streamResp.NextChunk()
		if err != nil {
			if errors.Is(err, io.EOF) {
				for _, f := range l.filters.OnCompletionStreamResponseFilters() {
					fResult := f.OnCompletionStreamResponse(ctx, request, streamResp, true)
					if fResult.IsFailed() {
						slog.Error("response handler", "error", fResult.Error)
					}
				}
				if err = marshalToWriter(chunk); err != nil {
					return nil
				}
				break
			}

			return openai.NewErrorInternalError().WithCause(err)
		}

		if chunk.IsEmpty() {
			continue
		}

		for _, f := range l.filters.OnCompletionStreamResponseFilters() {
			fResult := f.OnCompletionStreamResponse(ctx, request, streamResp, false)
			if fResult.IsFailed() {
				slog.Error("response handler", "error", fResult.Error)
			}
		}
		if err = marshalToWriter(chunk); err != nil {
			return err
		}
	}

	return nil
}

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

	var r route.Route
	var clusterName string

	// TODO: do route
	registryroute.ForeachRoute(func(item route.Route) bool {
		if cn, ok := item.Match(request.Context(), llmRequest); ok {
			clusterName = cn
			r = item

			return false
		}

		return true
	})

	if r == nil {
		return nil, openai.NewErrorModelNotFoundOrNotAccessible(llmRequest.GetModel())
	}

	c, ok := cluster.FindClusterByName(clusterName)
	if !ok {
		return nil, openai.NewErrorModelNotFoundOrNotAccessible(llmRequest.GetModel())
	}

	resp, err := c.DoUpstreamRequest(request.Context(), llmRequest)
	if err != nil {
		return nil, openai.NewErrorInternalError().WithCause(err)
	}

	if resp.GetError() != nil || !resp.IsStream() {
		err := c.DoUpstreamResponseComplete(request.Context(), llmRequest, resp)
		if err != nil {
			return nil, openai.NewErrorInternalError().WithCause(err)
		}

		for _, f := range l.filters.OnCompletionResponseFilters() {
			fResult := f.OnCompletionResponse(request.Context(), llmRequest, resp)
			if fResult.IsFailed() {
				slog.Error("response handler", "error", fResult.Error)
			}
		}

		if resp.GetError() != nil {
			return nil, resp.GetError()
		}
		return resp, nil
	}

	err = l.pipeChatCompletionsStream(request.Context(), llmRequest, resp, writer)
	if err != nil {
		return nil, err
	}

	// REVIEW: better way to compose the in and out actions?
	err = c.DoUpstreamResponseComplete(request.Context(), llmRequest, resp)
	if err != nil {
		return nil, openai.NewErrorInternalError().WithCause(err)
	}

	return nil, SkipResponse
}
