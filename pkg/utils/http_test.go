package utils

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Copied from: https://github.com/labstack/echo/blob/3b017855b4d331002e2b8b28e903679b875ae3e9/context_test.go#L76-L83
func BenchmarkRealIPForHeaderXForwardFor(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RealIPFromRequest(&http.Request{
			Header: http.Header{HeaderXForwardedFor: []string{"127.0.0.1, 127.0.1.1, "}},
		})
	}
}

// Copied from https://github.com/labstack/echo/blob/3b017855b4d331002e2b8b28e903679b875ae3e9/context_test.go#L1036-L1123
func TestRealIPFromRequest(t *testing.T) {
	tests := []struct {
		r *http.Request
		s string
	}{
		{
			&http.Request{
				Header: http.Header{HeaderXForwardedFor: []string{"127.0.0.1, 127.0.1.1, "}},
			},
			"127.0.0.1",
		},
		{
			&http.Request{
				Header: http.Header{HeaderXForwardedFor: []string{"127.0.0.1,127.0.1.1"}},
			},
			"127.0.0.1",
		},
		{
			&http.Request{
				Header: http.Header{HeaderXForwardedFor: []string{"127.0.0.1"}},
			},
			"127.0.0.1",
		},
		{
			&http.Request{
				Header: http.Header{HeaderXForwardedFor: []string{"[2001:db8:85a3:8d3:1319:8a2e:370:7348], 2001:db8::1, "}},
			},
			"2001:db8:85a3:8d3:1319:8a2e:370:7348",
		},
		{
			&http.Request{
				Header: http.Header{HeaderXForwardedFor: []string{"[2001:db8:85a3:8d3:1319:8a2e:370:7348],[2001:db8::1]"}},
			},
			"2001:db8:85a3:8d3:1319:8a2e:370:7348",
		},
		{
			&http.Request{
				Header: http.Header{HeaderXForwardedFor: []string{"2001:db8:85a3:8d3:1319:8a2e:370:7348"}},
			},
			"2001:db8:85a3:8d3:1319:8a2e:370:7348",
		},
		{
			&http.Request{
				Header: http.Header{
					"X-Real-Ip": []string{"192.168.0.1"},
				},
			},
			"192.168.0.1",
		},
		{
			&http.Request{
				Header: http.Header{
					"X-Real-Ip": []string{"[2001:db8::1]"},
				},
			},
			"2001:db8::1",
		},

		{
			&http.Request{
				RemoteAddr: "89.89.89.89:1654",
			},
			"89.89.89.89",
		},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.s, RealIPFromRequest(tt.r))
	}
}
