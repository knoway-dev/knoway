package chat

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/samber/lo"
	"github.com/samber/mo"
	goopenai "github.com/sashabaranov/go-openai"

	"knoway.dev/pkg/metadata"
	"knoway.dev/pkg/utils"

	v1alpha4 "knoway.dev/api/clusters/v1alpha1"
	"knoway.dev/pkg/clusters"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/types/openai"
)

func ClustersToOpenAIModels(clusters []*v1alpha4.Cluster) []goopenai.Model {
	res := make([]goopenai.Model, 0)
	for _, c := range clusters {
		res = append(res, ClusterToOpenAIModel(c))
	}

	return res
}

func ClusterToOpenAIModel(c *v1alpha4.Cluster) goopenai.Model {
	// from https://platform.openai.com/docs/api-reference/models/object
	return goopenai.Model{
		CreatedAt: c.GetCreated(),
		ID:        c.GetName(),
		// The object type, which is always "model".
		Object:  "model",
		OwnedBy: string(clusters.MapClusterProviderToBackendProvider(c.GetProvider())),
		// todo
		Permission: nil,
		Root:       "",
		Parent:     "",
	}
}

func (l *OpenAIChatListener) pipeCompletionsStream(ctx context.Context, request object.LLMRequest, streamResp object.LLMStreamResponse, writer http.ResponseWriter) {
	rMeta := metadata.RequestMetadataFromCtx(ctx)

	handleChunk := func(chunk object.LLMChunkResponse) error {
		for _, f := range l.reversedFilters.OnCompletionStreamResponseFilters() {
			fResult := f.OnCompletionStreamResponse(ctx, request, streamResp, chunk)
			if fResult.IsFailed() {
				// REVIEW: ignore? Or should fResult be returned?
				// Related topics: moderation, censorship, or filter keywords from the response
				slog.Error("error occurred during invoking of OnCompletionStreamResponse filters", "error", fResult.Error)
			}
		}

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

func (l *OpenAIChatListener) clusterDoCompletionsRequest(ctx context.Context, c clusters.Cluster, writer http.ResponseWriter, request *http.Request, llmRequest object.LLMRequest) (object.LLMResponse, error) {
	rMeta := metadata.RequestMetadataFromCtx(ctx)

	resp, err := c.DoUpstreamRequest(ctx, llmRequest)
	if err != nil {
		// Cluster will ensure that error will always be LLMError
		return nil, err
	}

	// For non-streaming responses, usage should be set here
	if !resp.IsStream() && !lo.IsNil(resp.GetUsage()) {
		rMeta.LLMUpstreamTokensUsage = mo.Some(lo.Must(object.AsLLMTokensUsage(resp.GetUsage())))
	}

	if resp.GetError() != nil || !resp.IsStream() {
		err := c.DoUpstreamResponseComplete(request.Context(), llmRequest, resp)
		if err != nil {
			// Cluster will ensure that error will always be LLMError
			return resp, err
		}

		if resp.GetError() != nil {
			return resp, resp.GetError()
		}

		return resp, nil
	}

	streamResp, ok := resp.(object.LLMStreamResponse)
	if !ok {
		return resp, openai.NewErrorInternalError().WithCausef("failed to cast %T to object.LLMStreamResponse", resp)
	}

	utils.WriteEventStreamHeadersForHTTP(writer)
	// NOTICE: from now on, there should not have any explicit error get returned
	// since the status code will be written by above call. If there is any error
	// it should be written as a chunk in the stream response.
	l.pipeCompletionsStream(request.Context(), llmRequest, streamResp, writer)

	// For streaming responses, usage should be set after the stream is done
	if !lo.IsNil(resp.GetUsage()) {
		rMeta.LLMUpstreamTokensUsage = mo.Some(lo.Must(object.AsLLMTokensUsage(resp.GetUsage())))
	}

	// REVIEW: better way to compose the in and out actions?
	err = c.DoUpstreamResponseComplete(request.Context(), llmRequest, resp)
	if err != nil {
		slog.Error("failed to call DoUpstreamResponseComplete", slog.Any("error", err))

		// Ignore, we shouldn't return any error here. Since the stream is already written
		// to the client, if any error occurred here, it should be logged and ignored.
		return resp, openai.SkipStreamResponse
	}

	return resp, openai.SkipStreamResponse
}
