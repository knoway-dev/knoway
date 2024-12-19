package utils

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
)

func SafeFlush(writer any) {
	f, ok := writer.(http.Flusher)
	if ok && f != nil {
		f.Flush()
	}
}

func WriteJSONForHTTP(status int, resp any, writer http.ResponseWriter) {
	bs, _ := json.Marshal(resp) //nolint:errchkjson

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)

	SafeFlush(writer)

	_, _ = writer.Write(bs)
}

func WriteEventStreamHeadersForHTTP(writer http.ResponseWriter) {
	writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	writer.Header().Set("Transfer-Encoding", "chunked")
	writer.WriteHeader(http.StatusOK)

	SafeFlush(writer)
}

const (
	HeaderXForwardedFor = "X-Forwarded-For"
	HeaderXRealIP       = "X-Real-Ip"
)

// Copied from https://github.com/labstack/echo/blob/3b017855b4d331002e2b8b28e903679b875ae3e9/context.go#L297
func RealIPFromRequest(request *http.Request) string {
	// Fall back to legacy behavior
	if ip := request.Header.Get(HeaderXForwardedFor); ip != "" {
		i := strings.IndexAny(ip, ",")
		if i > 0 {
			forwardedIP := strings.TrimSpace(ip[:i])
			forwardedIP = strings.TrimPrefix(forwardedIP, "[")
			forwardedIP = strings.TrimSuffix(forwardedIP, "]")

			return forwardedIP
		}

		return ip
	}
	if ip := request.Header.Get(HeaderXRealIP); ip != "" {
		ip = strings.TrimPrefix(ip, "[")
		ip = strings.TrimSuffix(ip, "]")

		return ip
	}

	ra, _, _ := net.SplitHostPort(request.RemoteAddr)

	return ra
}
