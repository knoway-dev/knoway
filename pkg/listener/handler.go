package listener

import (
	"net/http"
)

type HandlerFunc func(writer http.ResponseWriter, request *http.Request) (any, error)

type Middleware func(HandlerFunc) HandlerFunc

func WithMiddlewares(middlewares ...Middleware) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}

		return next
	}
}

func HTTPHandlerFunc(fn HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		_, _ = fn(writer, request)
	}
}
