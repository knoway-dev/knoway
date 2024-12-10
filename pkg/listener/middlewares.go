package listener

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/nekomeowww/fo"

	"knoway.dev/pkg/object"
	"knoway.dev/pkg/properties"
)

func WithProperties() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) (any, error) {
			return next(writer, request.WithContext(properties.NewPropertiesContext(request.Context())))
		}
	}
}

func WithOptions() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) (any, error) {
			if request.Method == http.MethodOptions {
				writer.WriteHeader(http.StatusNoContent)
				return nil, nil
			}
			return next(writer, request)
		}
	}
}

type CancellableRequestMap struct {
	mutex            sync.Mutex
	requestCancelMap map[*http.Request]context.CancelFunc
}

func NewCancellableRequestMap() *CancellableRequestMap {
	return &CancellableRequestMap{
		requestCancelMap: make(map[*http.Request]context.CancelFunc),
	}
}

func (l *CancellableRequestMap) Add(req *http.Request, cancel context.CancelFunc) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.requestCancelMap[req] = cancel
}

func (l *CancellableRequestMap) Remove(req *http.Request) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	delete(l.requestCancelMap, req)
}

func (l *CancellableRequestMap) CancelAll() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	for _, cancel := range l.requestCancelMap {
		cancel()
	}
}

func (l *CancellableRequestMap) CancelAllAfter(timeout time.Duration) {
	var wg sync.WaitGroup

	wg.Add(1)
	time.AfterFunc(timeout, func() {
		defer wg.Done()

		// Lock in callback function to prevent
		// lock acquisition order violation
		l.mutex.Lock()
		defer l.mutex.Unlock()

		for _, cancel := range l.requestCancelMap {
			cancel()
		}
	})
	wg.Wait()
}

func (l *CancellableRequestMap) CancelAllWithContext(ctx context.Context) {
	_ = fo.Invoke0(ctx, func() error {
		l.CancelAll()

		return nil
	})
}

func (l *CancellableRequestMap) CancelAllAfterWithContext(ctx context.Context, timeout time.Duration) {
	_ = fo.Invoke0(ctx, func() error {
		l.CancelAllAfter(timeout)

		return nil
	})
}

func CancellableInterceptor(cancellable *CancellableRequestMap) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) (any, error) {
			ctx, cancel := context.WithCancel(request.Context())
			cancellable.Add(request, cancel)
			defer cancellable.Remove(request)

			return next(writer, request.WithContext(ctx))
		}
	}
}

func RejectAfterDrainedInterceptor(d Drainable) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) (any, error) {
			if d.HasDrained() {
				return nil, object.NewErrorServiceUnavailable()
			}

			return next(writer, request)
		}
	}
}
