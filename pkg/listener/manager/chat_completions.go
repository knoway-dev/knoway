package manager

import (
	"context"
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

func NewOpenAIChatCompletionsListenerWithConfigs(cfg proto.Message) (listener.Listener, error) {
	c, ok := cfg.(*v1alpha1.ChatCompletionListener)
	if !ok {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}

	l := &OpenAIChatCompletionsListener{
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

type OpenAIChatCompletionsListener struct {
	cfg               *v1alpha1.ChatCompletionListener
	filters           []filters.RequestFilter
	listener.Listener // TODO: implement the interface
}

func (l *OpenAIChatCompletionsListener) UnmarshalLLMRequest(
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

func (l *OpenAIChatCompletionsListener) handleChatCompletionsChunks(resp object.LLMResponse, writer http.ResponseWriter) error {
	streamResp, ok := resp.(object.LLMStreamResponse)
	if !ok {
		return openai.NewErrorInternalError().WithCausef("failed to cast %T to object.LLMStreamResponse", resp)
	}

	utils.WriteEventStreamHeadersForHTTP(writer)

	for {
		chunk, err := streamResp.NextChunk()
		if err != nil {
			if errors.Is(err, io.EOF) {
				event, err := chunk.ToServerSentEvent()
				if err != nil {
					return openai.NewErrorInternalError().WithCause(err)
				}

				err = event.MarshalTo(writer)
				if err != nil {
					slog.Error("failed to marshal event", "error", err)
				}

				break
			}

			return openai.NewErrorInternalError().WithCause(err)
		}

		if chunk.IsEmpty() {
			continue
		}

		// TODO: request filter
		event, err := chunk.ToServerSentEvent()
		if err != nil {
			return openai.NewErrorInternalError().WithCause(err)
		}

		err = event.MarshalTo(writer)
		if err != nil {
			slog.Error("failed to marshal event", "error", err)
		}
	}

	return nil
}

func (l *OpenAIChatCompletionsListener) onChatCompletionsRequestWithError(writer http.ResponseWriter, request *http.Request) (any, error) {
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

	err = l.handleChatCompletionsChunks(resp, writer)
	if err != nil {
		return nil, err
	}

	return nil, openai.SkipResponse
}

func (l *OpenAIChatCompletionsListener) RegisterRoutes(mux *mux.Router) error {
	mux.HandleFunc("/v1/chat/completions", openai.WrapHandlerForOpenAIError(l.onChatCompletionsRequestWithError))

	return nil
}
