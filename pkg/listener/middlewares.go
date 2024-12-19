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

func WithLog(pickRequestHeaders []string, pickUpstreamResponseHeaders []string) Middleware {
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
				slog.String("user_agent", request.UserAgent()),
				slog.String("referer", request.Referer()),
				slog.Any("headers", lo.PickByKeys(request.Header, pickRequestHeaders)),
				slog.Any("query", request.URL.Query()),
				slog.Int64("latency", rMeta.RespondAt.Sub(rMeta.RequestAt).Milliseconds()),
				slog.Duration("latency_human", rMeta.RespondAt.Sub(rMeta.RequestAt)),
				slog.String("auth_info_api_key_id", rMeta.AuthInfo.GetApiKeyId()),
				slog.String("auth_info_user_id", rMeta.AuthInfo.GetUserId()),
				slog.String("request_model", rMeta.RequestModel),
				slog.String("response_model", rMeta.ResponseModel),
				slog.String("upstream_provider", rMeta.UpstreamProvider),
				slog.String("upstream_request_model", rMeta.UpstreamRequestModel),
				slog.String("upstream_response_model", rMeta.UpstreamResponseModel),
				slog.Int("upstream_response_status_code", rMeta.UpstreamResponseStatusCode),
				slog.Any("upstream_response_header", lo.PickByKeys(rMeta.UpstreamResponseHeader.OrElse(http.Header{}), pickUpstreamResponseHeaders)),
				slog.Uint64("llm_usage_prompt_tokens", rMeta.LLMUpstreamUsage.OrElse(object.DefaultLLMUsage{}).GetPromptTokens()),
				slog.Uint64("llm_usage_completion_tokens", rMeta.LLMUpstreamUsage.OrElse(object.DefaultLLMUsage{}).GetCompletionTokens()),
				slog.Uint64("llm_usage_total_tokens", rMeta.LLMUpstreamUsage.OrElse(object.DefaultLLMUsage{}).GetTotalTokens()),
			}

			if !rMeta.UpstreamRespondAt.IsZero() {
				attrs = append(attrs,
					slog.Int64("upstream_latency", rMeta.UpstreamRespondAt.Sub(rMeta.UpstreamRequestAt).Milliseconds()),
					slog.Duration("upstream_latency_human", rMeta.UpstreamRespondAt.Sub(rMeta.UpstreamRequestAt)),
				)
			}
			if !rMeta.UpstreamFirstValidChunkAt.IsZero() {
				attrs = append(attrs,
					slog.Int64("upstream_first_chunk_latency", rMeta.UpstreamFirstValidChunkAt.Sub(rMeta.UpstreamRequestAt).Milliseconds()),
					slog.Duration("upstream_first_chunk_latency_human", rMeta.UpstreamFirstValidChunkAt.Sub(rMeta.UpstreamRequestAt)),
				)
			}

			slog.Info("", attrs...)

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

func WithRecoverWithError() Middleware {
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

func WithCancellable(cancellable *CancellableRequestMap) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) (any, error) {
			ctx, cancel := context.WithCancel(request.Context())
			cancellable.Add(request, cancel)
			defer cancellable.Remove(request)

			return next(writer, request.WithContext(ctx))
		}
	}
}

func WithRejectAfterDrainedWithError(d Drainable) Middleware {
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
			metadata.RequestMetadataFromCtx(request.Context()).RequestAt = time.Now()
			resp, err := next(writer, request)
			metadata.RequestMetadataFromCtx(request.Context()).RespondAt = time.Now()

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
