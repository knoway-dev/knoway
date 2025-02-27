package chat

import (
	"context"
	"fmt"
	"sync"

	"github.com/gorilla/mux"
	"github.com/samber/lo/mutable"
	"google.golang.org/protobuf/proto"

	"knoway.dev/api/listeners/v1alpha1"
	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/constants"
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/listener"
	"knoway.dev/pkg/registry/config"
	"knoway.dev/pkg/types/openai"
	"knoway.dev/pkg/utils"
)

var _ listener.Listener = (*OpenAIChatListener)(nil)
var _ listener.Drainable = (*OpenAIChatListener)(nil)

type OpenAIChatListener struct {
	cfg             *v1alpha1.ChatCompletionListener
	filters         filters.RequestFilters
	reversedFilters filters.RequestFilters
	cancellable     *listener.CancellableRequestMap

	mutex   sync.RWMutex
	drained bool
}

func NewOpenAIChatListenerConfigs(cfg proto.Message, lifecycle bootkit.LifeCycle) (listener.Listener, error) {
	c, ok := cfg.(*v1alpha1.ChatCompletionListener)
	if !ok {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}

	l := &OpenAIChatListener{
		cfg:         c,
		cancellable: listener.NewCancellableRequestMap(),
	}

	lifecycle.Append(bootkit.LifeCycleHook{
		OnStop: l.Drain,
	})

	for _, fc := range c.GetFilters() {
		f, err := config.NewRequestFilterWithConfig(fc.GetName(), fc.GetConfig(), lifecycle)
		if err != nil {
			return nil, err
		}

		l.filters = append(l.filters, f)
	}

	l.reversedFilters = utils.Clone(l.filters)
	mutable.Reverse(l.reversedFilters)

	return l, nil
}

func (l *OpenAIChatListener) RegisterRoutes(mux *mux.Router) error {
	middlewares := listener.WithMiddlewares(
		listener.WithCancellable(l.cancellable),
		listener.WithInitMetadata(),
		listener.WithAccessLog(l.cfg.GetAccessLog().GetEnable()),
		listener.WithRequestTimer(),
		listener.WithOptions(),
		listener.WithResponseHandler(openai.ResponseHandler()),
		listener.WithRecoverWithError(),
		listener.WithRejectAfterDrainedWithError(l),
	)

	mux.HandleFunc("/v1/chat/completions", listener.HTTPHandlerFunc(middlewares(listener.CommonListenerHandler(l.filters, l.reversedFilters, l.unmarshalChatCompletionsRequestToLLMRequest, l.clusterDoCompletionsRequest))))
	mux.HandleFunc("/v1/completions", listener.HTTPHandlerFunc(middlewares(listener.CommonListenerHandler(l.filters, l.reversedFilters, l.unmarshalCompletionsRequestToLLMRequest, l.clusterDoCompletionsRequest))))
	mux.HandleFunc("/v1/models", listener.HTTPHandlerFunc(middlewares(l.listModels)))

	return nil
}

func (l *OpenAIChatListener) HasDrained() bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	return l.drained
}

func (l *OpenAIChatListener) Drain(ctx context.Context) error {
	l.mutex.Lock()
	l.drained = true
	l.mutex.Unlock()

	l.cancellable.CancelAllAfterWithContext(ctx, constants.DefaultDrainWaitTime)

	return nil
}
