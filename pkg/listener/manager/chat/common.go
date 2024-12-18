package chat

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	goopenai "github.com/sashabaranov/go-openai"

	"knoway.dev/pkg/metadata"

	v1alpha4 "knoway.dev/api/clusters/v1alpha1"
	"knoway.dev/pkg/clusters"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/registry/cluster"
	registryroute "knoway.dev/pkg/registry/route"
	"knoway.dev/pkg/route"
	"knoway.dev/pkg/types/openai"
	"knoway.dev/pkg/utils"
)

func ClustersToOpenAIModels(clusters []*v1alpha4.Cluster) []goopenai.Model {
	res := make([]goopenai.Model, 0)
	for _, c := range clusters {
		res = append(res, ClusterToOpenAIModel(c))
	}

	return res
}

func ClusterToOpenAIModel(cluster *v1alpha4.Cluster) goopenai.Model {
	// from https://platform.openai.com/docs/api-reference/models/object
	return goopenai.Model{
		CreatedAt: cluster.GetCreated(),
		ID:        cluster.GetName(),
		// The object type, which is always "model".
		Object:  "model",
		OwnedBy: cluster.GetProvider(),
		// todo
		Permission: nil,
		Root:       "",
		Parent:     "",
	}
}

var (
	SkipResponse = errors.New("skip writing response") //nolint:errname,stylecheck
)

func (l *OpenAIChatListener) pipeCompletionsStream(ctx context.Context, request object.LLMRequest, resp object.LLMResponse, writer http.ResponseWriter) error {
	streamResp, ok := resp.(object.LLMStreamResponse)
	if !ok {
		return openai.NewErrorInternalError().WithCausef("failed to cast %T to object.LLMStreamResponse", resp)
	}

	utils.WriteEventStreamHeadersForHTTP(writer)

	marshalToWriter := func(chunk object.LLMChunkResponse) error {
		event, err := chunk.ToServerSentEvent()
		if err != nil {
			slog.Error("failed to convert chunk body to server sent event payload", "error", err)
			return openai.NewErrorInternalError().WithCause(err)
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
			if errors.Is(err, io.EOF) {
				for _, f := range l.filters.OnCompletionStreamResponseFilters() {
					fResult := f.OnCompletionStreamResponse(ctx, request, streamResp, chunk)
					if fResult.IsFailed() {
						slog.Error("error occurred during invoking of OnCompletionStreamResponse filters", "error", fResult.Error)
					}
				}
				if err = marshalToWriter(chunk); err != nil {
					// Ignore
					return nil //nolint:nilerr
				}

				break
			}

			return openai.NewErrorInternalError().WithCause(err)
		}

		if chunk.IsEmpty() {
			continue
		}
		if chunk.IsDone() {
			metadata.RequestMetadataFromCtx(ctx).UpstreamResponseTime = time.Now()
		}
		if chunk.IsFirst() {
			metadata.RequestMetadataFromCtx(ctx).UpstreamFirstChunkResponseTime = time.Now()
		}

		for _, f := range l.filters.OnCompletionStreamResponseFilters() {
			fResult := f.OnCompletionStreamResponse(ctx, request, streamResp, chunk)
			if fResult.IsFailed() {
				slog.Error("error occurred during invoking of OnCompletionStreamResponse filters", "error", fResult.Error)
			}
		}
		if err = marshalToWriter(chunk); err != nil {
			return err
		}
	}

	return nil
}

func (l *OpenAIChatListener) findRoute(ctx context.Context, llmRequest object.LLMRequest) (route.Route, string) {
	var r route.Route
	var clusterName string

	// TODO: do route
	registryroute.ForeachRoute(func(item route.Route) bool {
		if cn, ok := item.Match(ctx, llmRequest); ok {
			clusterName = cn
			r = item

			return false
		}

		return true
	})

	return r, clusterName
}

func (l *OpenAIChatListener) findCluster(ctx context.Context, llmRequest object.LLMRequest) (clusters.Cluster, bool) {
	r, clusterName := l.findRoute(ctx, llmRequest)
	if r == nil {
		return nil, false
	}

	c, ok := cluster.FindClusterByName(clusterName)
	if !ok {
		return nil, false
	}

	return c, true
}

func (l *OpenAIChatListener) clusterDoCompletionsRequest(ctx context.Context, c clusters.Cluster, writer http.ResponseWriter, request *http.Request, llmRequest object.LLMRequest) (object.LLMResponse, error) {
	resp, err := c.DoUpstreamRequest(ctx, llmRequest)
	if err != nil {
		return nil, openai.NewErrorInternalError().WithCause(err)
	}

	if resp.GetError() != nil || !resp.IsStream() {
		metadata.RequestMetadataFromCtx(ctx).UpstreamResponseTime = time.Now()

		err := c.DoUpstreamResponseComplete(request.Context(), llmRequest, resp)
		if err != nil {
			return nil, openai.NewErrorInternalError().WithCause(err)
		}

		if resp.GetError() != nil {
			return nil, resp.GetError()
		}

		return resp, nil
	}

	err = l.pipeCompletionsStream(request.Context(), llmRequest, resp, writer)
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
