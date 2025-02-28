package listener

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/samber/lo"
	"github.com/samber/mo"
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/metadata"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/route/manager"
	"knoway.dev/pkg/types/openai"
	"knoway.dev/pkg/utils"
)

func CommonListenerHandler(
	listenerFilters filters.RequestFilters,
	reversedFilters filters.RequestFilters,
	parseRequest func(request *http.Request) (object.LLMRequest, error),
) func(writer http.ResponseWriter, request *http.Request) (any, error) {
	return func(writer http.ResponseWriter, request *http.Request) (any, error) {
		var err error

		for _, f := range listenerFilters.OnRequestPreFilters() {
			fResult := f.OnRequestPre(request.Context(), request)
			if fResult.IsFailed() {
				return nil, fResult.Error
			}
		}

		var resp object.LLMResponse

		defer func() {
			for _, f := range reversedFilters.OnResponsePostFilters() {
				f.OnResponsePost(request.Context(), request, resp, err)
			}
		}()

		llmRequest, err := parseRequest(request)
		if err != nil {
			return nil, err
		}

		switch llmRequest.GetRequestType() {
		case object.RequestTypeChatCompletions, object.RequestTypeCompletions:
			for _, f := range listenerFilters.OnCompletionRequestFilters() {
				fResult := f.OnCompletionRequest(request.Context(), llmRequest, request)
				if fResult.IsFailed() {
					return nil, fResult.Error
				}
			}
		case object.RequestTypeImageGenerations:
			for _, f := range listenerFilters.OnImageGenerationsRequestFilters() {
				fResult := f.OnImageGenerationsRequest(request.Context(), llmRequest, request)
				if fResult.IsFailed() {
					return nil, fResult.Error
				}
			}
		}

		resp, err = manager.HandleRequest(request.Context(), llmRequest)
		if err != nil {
			return resp, err
		}

		// Non-streaming responses
		if !resp.IsStream() {
			return resp, err
		}

		// Streaming responses
		streamResp, ok := resp.(object.LLMStreamResponse)
		if !ok {
			return resp, openai.NewErrorInternalError().WithCausef("failed to cast %T to object.LLMStreamResponse", resp)
		}

		streamResp.OnChunk(func(ctx context.Context, stream object.LLMStreamResponse, chunk object.LLMChunkResponse) {
			for _, f := range reversedFilters.OnCompletionStreamResponseFilters() {
				fResult := f.OnCompletionStreamResponse(ctx, llmRequest, streamResp, chunk)
				if fResult.IsFailed() {
					// REVIEW: ignore? Or should fResult be returned?
					// Related topics: moderation, censorship, or filter keywords from the response
					slog.Error("error occurred during invoking of OnCompletionStreamResponse filters", "error", fResult.Error)
				}
			}
		})

		utils.WriteEventStreamHeadersForHTTP(writer)
		// NOTICE: from now on, there should not have any explicit error get returned
		// since the status code will be written by above call. If there is any error
		// it should be written as a chunk in the stream response.
		pipeCompletionsStream(request.Context(), listenerFilters, reversedFilters, llmRequest, streamResp, writer)

		return resp, openai.SkipStreamResponse
	}
}

func pipeCompletionsStream(ctx context.Context, _ filters.RequestFilters, reversedFilters filters.RequestFilters, request object.LLMRequest, streamResp object.LLMStreamResponse, writer http.ResponseWriter) {
	rMeta := metadata.RequestMetadataFromCtx(ctx)

	handleChunk := func(chunk object.LLMChunkResponse) error {
		event, err := chunk.ToServerSentEvent()
		if err != nil {
			slog.Error("failed to convert chunk body to server sent event payload", "error", err)
			return err
		}

		err = event.MarshalTo(writer)
		if err != nil {
			slog.Error("failed to write SSE event into http.ResponseWriter", "error", err)
			return err
		}

		return nil
	}

	for {
		chunk, err := streamResp.NextChunk()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				slog.Error("failed to get next chunk from stream response", slog.Any("error", err))
				return
			}

			// EOF, send last chunk
			if err := handleChunk(chunk); err != nil {
				// Ignore, terminate stream reading
				return
			}

			// Then terminate the stream
			break
		}

		if chunk.IsEmpty() {
			continue
		}
		if chunk.IsUsage() && !lo.IsNil(chunk.GetUsage()) {
			rMeta.LLMUpstreamTokensUsage = mo.Some(lo.Must(object.AsLLMTokensUsage(chunk.GetUsage())))
		}
		if chunk.IsFirst() {
			rMeta.UpstreamFirstValidChunkAt = time.Now()
			rMeta.UpstreamResponseModel = chunk.GetModel()
		}

		if err := handleChunk(chunk); err != nil {
			// Ignore, terminate stream reading
			return
		}
	}
}
