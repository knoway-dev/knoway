package manager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"google.golang.org/protobuf/proto"

	"knoway.dev/api/listeners/v1alpha1"
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/listener"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/registry/cluster"
	"knoway.dev/pkg/registry/config"
	"knoway.dev/pkg/registry/route"
	route2 "knoway.dev/pkg/route"
	"knoway.dev/pkg/types/openai"
	"knoway.dev/pkg/utils"
)

var (
	errSkipWriteResponse = errors.New("skip writing response")
)

func NewWithConfigs(cfg proto.Message) (listener.Listener, error) {
	c, ok := cfg.(*v1alpha1.ChatCompletionListener)
	if !ok {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}

	l := &OpenAIChatCompletionListener{
		cfg: c,
	}

	for _, fc := range c.Filters {
		f, err := config.NewRequestFilterWithConfig(fc.Name, fc.Config)
		if err != nil {
			return nil, err
		}

		l.filters = append(l.filters, f)
	}

	return l, nil
}

type OpenAIChatCompletionListener struct {
	cfg               *v1alpha1.ChatCompletionListener
	filters           []filters.RequestFilter
	listener.Listener // todo implement the interface
}

func (l *OpenAIChatCompletionListener) UnmarshalLLMRequest(
	ctx context.Context,
	request *http.Request,
) (object.LLMRequest, error) {
	llmRequest, err := openai.NewChatCompletionRequest(request)
	if err != nil {
		return nil, err
	}

	if llmRequest.GetModel() == "" {
		return nil, openai.NewErrorMissingModel()
	}

	return llmRequest, nil
}

func writeResponse(status int, resp any, writer http.ResponseWriter) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)

	bs, _ := json.Marshal(resp)
	_, _ = writer.Write(bs)
}

func prepareWriteEventStream(writer http.ResponseWriter) {
	writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	writer.Header().Set("Transfer-Encoding", "chunked")

	utils.SafeFlush(writer)
}

func (l *OpenAIChatCompletionListener) onChatCompletionRequestWithError(writer http.ResponseWriter, request *http.Request) (any, error) {
	req, err := l.UnmarshalLLMRequest(request.Context(), request)
	if err != nil {
		return nil, err
	}

	for _, f := range l.filters {
		res := f.OnCompletionRequest(request.Context(), req)
		if res.Type == filters.ListenerFilterResultTypeFailed {
			return nil, openai.NewErrorIncorrectAPIKey()
		}
	}

	var r route2.Route
	var clusterName string

	// TODO: do route
	route.ForeachRoute(func(item route2.Route) bool {
		if cn, ok := item.Match(request.Context(), req); ok {
			clusterName = cn
			r = item

			return false
		}

		return true
	})

	if r == nil {
		return nil, openai.NewErrorModelNotFoundOrNotAccessible(req.GetModel())
	}

	c, ok := cluster.FindClusterByName(clusterName)
	if !ok {
		return nil, openai.NewErrorModelNotFoundOrNotAccessible(req.GetModel())
	}

	resp, err := c.DoUpstreamRequest(request.Context(), req)
	if err != nil {
		return nil, openai.NewErrorInternalError().WithCause(err)
	}

	if resp.GetError() != nil {
		return nil, resp.GetError()
	}
	if !resp.IsStream() { //nolint:wsl
		return resp, nil
	}

	streamResp, ok := resp.(object.LLMStreamResponse)
	if !ok {
		return nil, openai.NewErrorInternalError().WithCausef("failed to cast %T to object.LLMStreamResponse", resp)
	}

	prepareWriteEventStream(writer)

	for {
		chunk, err := streamResp.NextChunk()
		if err != nil {
			if errors.Is(err, io.EOF) {
				event, err := chunk.ToServerSentEvent()
				if err != nil {
					return nil, openai.NewErrorInternalError().WithCause(err)
				}

				err = event.MarshalTo(writer)
				if err != nil {
					slog.Error("failed to marshal event", "error", err)
				}

				break
			}

			return nil, err
		}

		if chunk.IsEmpty() {
			continue
		}

		// TODO: request filter
		event, err := chunk.ToServerSentEvent()
		if err != nil {
			return nil, err
		}

		err = event.MarshalTo(writer)
		if err != nil {
			slog.Error("failed to marshal event", "error", err)
		}
	}

	return nil, errSkipWriteResponse
}

func (l *OpenAIChatCompletionListener) wrapErrorHandler(fn func(writer http.ResponseWriter, request *http.Request) (any, error)) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		resp, err := fn(writer, request)
		if err == nil {
			if resp != nil {
				writeResponse(http.StatusOK, resp, writer)
			}

			return
		}

		if errors.Is(err, errSkipWriteResponse) {
			return
		}

		var openAIError *openai.ErrorResponse
		if errors.As(err, &openAIError) && openAIError != nil {
			if openAIError.FromUpstream {
				slog.Error("upstream returned an error",
					"status", openAIError.Status,
					"code", openAIError.ErrorBody.Code,
					"message", openAIError.ErrorBody.Message,
					"type", openAIError.ErrorBody.Type,
				)
			} else {
				if openAIError.Status >= 500 {
					slog.Error("failed to handle request", "error", openAIError, "cause", openAIError.Cause)
				}
			}

			writeResponse(openAIError.Status, openAIError, writer)

			return
		}

		slog.Error("failed to handle request, unhandled error occurred", "error", err)
		writeResponse(http.StatusInternalServerError, openai.NewErrorInternalError(), writer)
	}
}

func (l *OpenAIChatCompletionListener) RegisterRoutes(mux *mux.Router) error {
	mux.HandleFunc("/v1/chat/completions", l.wrapErrorHandler(l.onChatCompletionRequestWithError))

	return nil
}
