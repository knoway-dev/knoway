package listener

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"knoway.dev/pkg/metadata"

	"github.com/nekomeowww/fo"
	"github.com/samber/lo"

	"knoway.dev/pkg/object"
	"knoway.dev/pkg/types/openai"
	"knoway.dev/pkg/utils"
)

func WithLog() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) (any, error) {
			resp, err := next(writer, request)

			rMeta := metadata.RequestMetadataFromCtx(request.Context())
			attrs := []any{
				slog.Int("status", rMeta.StatusCode),
				slog.String("remote_ip", utils.RealIPFromRequest(request)),
				slog.String("host", request.Host),
				slog.String("uri", request.RequestURI),
				slog.String("method", request.Method),
				slog.String("path", request.URL.Path),
				slog.String("protocol", request.Proto),
				slog.String("referer", request.Referer()),
				slog.String("user_agent", request.UserAgent()),
				slog.Int64("latency", time.Since(rMeta.RequestTime).Milliseconds()),
				slog.Duration("latency_human", time.Since(rMeta.RequestTime)),
				slog.Any("headers", lo.OmitByKeys(request.Header, []string{"Authorization"})),
				slog.Any("query", request.URL.Query()),
				slog.Any("cookies", lo.Map(request.Cookies(), func(item *http.Cookie, index int) string {
					return item.String()
				})),
			}

			if rMeta.ErrorMessage != "" {
				attrs = append(attrs, slog.String("error", rMeta.ErrorMessage))
			}

			slog.Info("request handled", attrs...)

			return resp, err
		}
	}
}

func WithInitMetadata() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) (any, error) {
			return next(writer, request.WithContext(metadata.InitMetadataContext(request)))
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

func WithRecover() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) (any, error) {
			defer func() {
				if r := recover(); r != nil {
					var url string
					if request != nil && request.URL != nil {
						url = request.URL.String()
					}

					stack := string(debug.Stack())

					slog.Error("Recovered from panic",
						slog.Any("panic", r),
						slog.String("url", url),
						slog.String("stack", stack),
					)

					internalErr := openai.NewErrorInternalError()

					utils.WriteJSONForHTTP(internalErr.Status, internalErr, writer)
				}
			}()

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

func WithCancellableInterceptor(cancellable *CancellableRequestMap) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) (any, error) {
			ctx, cancel := context.WithCancel(request.Context())
			cancellable.Add(request, cancel)
			defer cancellable.Remove(request)

			return next(writer, request.WithContext(ctx))
		}
	}
}

func WithRejectAfterDrainedInterceptor(d Drainable) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) (any, error) {
			if d.HasDrained() {
				return nil, object.NewErrorServiceUnavailable()
			}

			return next(writer, request)
		}
	}
}

func WithRequestTimer() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) (any, error) {
			metadata.RequestMetadataFromCtx(request.Context()).RequestTime = time.Now()
			resp, err := next(writer, request)
			metadata.RequestMetadataFromCtx(request.Context()).ResponseTime = time.Now()

			return resp, err
		}
	}
}

func WithResponseHandler(fn func(resp any, err error, writer http.ResponseWriter, request *http.Request)) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) (any, error) {
			resp, err := next(writer, request)
			fn(resp, err, writer, request)

			return nil, nil
		}
	}
}
