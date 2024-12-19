package chat

import (
	"context"
	"fmt"
	"sync"

	"github.com/gorilla/mux"
	"google.golang.org/protobuf/proto"

	"knoway.dev/api/listeners/v1alpha1"
	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/constants"
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/listener"
	"knoway.dev/pkg/registry/config"
)

var _ listener.Listener = (*OpenAIChatListener)(nil)
var _ listener.Drainable = (*OpenAIChatListener)(nil)

type OpenAIChatListener struct {
	cfg         *v1alpha1.ChatCompletionListener
	filters     filters.RequestFilters
	cancellable *listener.CancellableRequestMap

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

	return l, nil
}

func (l *OpenAIChatListener) RegisterRoutes(mux *mux.Router) error {
	middlewares := listener.WithMiddlewares(
		listener.WithInitMetadata(),
		listener.WithLog(),
		listener.WithRequestTimer(),
		listener.WithResponseHandler(ResponseHandler()),
		listener.WithRecover(),
		listener.WithOptions(),
		listener.WithCancellableInterceptor(l.cancellable),
		listener.WithRejectAfterDrainedInterceptor(l),
	)

	mux.HandleFunc("/v1/chat/completions", listener.HTTPHandlerFunc(middlewares(l.onChatCompletionsRequestWithError)))
	mux.HandleFunc("/v1/completions", listener.HTTPHandlerFunc(middlewares(l.onCompletionsRequestWithError)))
	mux.HandleFunc("/v1/models", listener.HTTPHandlerFunc(middlewares(l.onListModelsRequestWithError)))

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
