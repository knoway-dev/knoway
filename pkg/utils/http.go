package utils

import (
	"encoding/json"
	"net/http"
)

func SafeFlush(writer any) {
	f, ok := writer.(http.Flusher)
	if ok && f != nil {
		f.Flush()
	}
}

func WriteJSONForHTTP(status int, resp any, writer http.ResponseWriter) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)

	SafeFlush(writer)

	bs, _ := json.Marshal(resp)
	_, _ = writer.Write(bs)
}

func WriteEventStreamHeadersForHTTP(writer http.ResponseWriter) {
	writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	writer.Header().Set("Transfer-Encoding", "chunked")

	SafeFlush(writer)
}
