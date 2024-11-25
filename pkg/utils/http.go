package utils

import "net/http"

func SafeFlush(writer any) {
	f, ok := writer.(http.Flusher)
	if ok && f != nil {
		f.Flush()
	}
}
