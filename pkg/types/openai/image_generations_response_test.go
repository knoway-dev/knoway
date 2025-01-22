package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/nekomeowww/xo"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/test/bufconn"
)

func testServer(t *testing.T, handler http.Handler) (*http.Client, func(), func()) {
	t.Helper()

	const bufSize = 1024 * 1024
	listener := bufconn.Listen(bufSize)

	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: time.Second,
	}

	start := func() {
		err := server.Serve(listener)
		if !errors.Is(err, http.ErrServerClosed) {
			assert.NoError(t, err)
		}
	}

	stop := func() {
		if server == nil {
			return
		}

		err := server.Shutdown(context.Background())
		if !errors.Is(err, http.ErrServerClosed) {
			assert.NoError(t, err)
		}
	}

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
				return listener.Dial()
			},
			Dial: func(network, addr string) (net.Conn, error) {
				return listener.Dial()
			},
			DialTLSContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
				return listener.Dial()
			},
			DialTLS: func(network, addr string) (net.Conn, error) {
				return listener.Dial()
			},
		},
	}

	return client, start, stop
}

func TestNewImageGenerationsResponse(t *testing.T) {
	tests := []struct {
		name           string
		imageFile      string
		isURL          bool
		expectedWidth  int
		expectedHeight int
		expectedFormat string
		contentType    string
		urlPath        string
	}{
		{
			name:           "Base64-Jpeg",
			imageFile:      "./testdata/SampleJPGImage_100kbmb.jpg",
			isURL:          false,
			expectedWidth:  689,
			expectedHeight: 689,
			expectedFormat: "jpeg",
		},
		{
			name:           "Base64-Png",
			imageFile:      "./testdata/SamplePNGImage_100kbmb.png",
			isURL:          false,
			expectedWidth:  272,
			expectedHeight: 170,
			expectedFormat: "png",
		},
		{
			name:           "Base64-Webp",
			imageFile:      "./testdata/GoogleSampleWebpImage.webp",
			isURL:          false,
			expectedWidth:  1024,
			expectedHeight: 772,
			expectedFormat: "webp",
		},
		{
			name:           "Base64-Gif",
			imageFile:      "./testdata/SampleGIFImage_135kbmb.gif",
			isURL:          false,
			expectedWidth:  492,
			expectedHeight: 229,
			expectedFormat: "gif",
		},
		{
			name:           "URL-Jpeg",
			imageFile:      "./testdata/SampleJPGImage_100kbmb.jpg",
			isURL:          true,
			expectedWidth:  689,
			expectedHeight: 689,
			expectedFormat: "jpeg",
			contentType:    "image/jpeg",
			urlPath:        "https://example.com/jpeg.jpg",
		},
		{
			name:           "URL-Png",
			imageFile:      "./testdata/SamplePNGImage_100kbmb.png",
			isURL:          true,
			expectedWidth:  272,
			expectedHeight: 170,
			expectedFormat: "png",
			contentType:    "image/png",
			urlPath:        "https://example.com/png.png",
		},
		{
			name:           "URL-Gif",
			imageFile:      "./testdata/SampleGIFImage_135kbmb.gif",
			isURL:          true,
			expectedWidth:  492,
			expectedHeight: 229,
			expectedFormat: "gif",
			contentType:    "image/gif",
			urlPath:        "https://example.com/gif.gif",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imageContent, err := os.ReadFile(xo.RelativePathOf(tt.imageFile))
			require.NoError(t, err)

			var responseBody []byte
			var client *http.Client
			var start, stop func()

			if tt.isURL {
				client, start, stop = testServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", tt.contentType)
					w.WriteHeader(http.StatusOK)
					_, err := w.Write(imageContent)
					assert.NoError(t, err)
				}))
				go start()
				defer stop()

				responseBody, err = json.Marshal(map[string]any{
					"data": []map[string]any{
						{
							"url": tt.urlPath,
						},
					},
				})
			} else {
				base64Str := base64.StdEncoding.EncodeToString(imageContent)
				assert.Equal(t, imageContent, lo.Must(base64.StdEncoding.DecodeString(base64Str)))

				responseBody, err = json.Marshal(map[string]any{
					"data": []map[string]any{
						{
							"b64_json": base64Str,
						},
					},
				})
			}

			require.NoError(t, err)

			req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, "http://localhost/v1/images/generations", nil)
			require.NoError(t, err)

			resp := &http.Response{
				StatusCode: http.StatusOK,
				Request:    req,
				Body:       io.NopCloser(bytes.NewReader(responseBody)),
			}

			reader := bufio.NewReader(bytes.NewReader(responseBody))

			var imageGenerationsResp *ImageGenerationsResponse
			if tt.isURL {
				imageGenerationsResp, err = NewImageGenerationsResponse(context.Background(), &ImageGenerationsRequest{}, resp, reader, NewImageGenerationsResponseWithHTTPClient(client))
			} else {
				imageGenerationsResp, err = NewImageGenerationsResponse(context.Background(), &ImageGenerationsRequest{}, resp, reader)
			}

			require.NoError(t, err)

			assert.Len(t, imageGenerationsResp.Images, 1)
			if tt.isURL {
				assert.Empty(t, imageGenerationsResp.Images[0].Base64JSON)
				assert.Equal(t, tt.urlPath, imageGenerationsResp.Images[0].URL)
			} else {
				assert.NotEmpty(t, imageGenerationsResp.Images[0].Base64JSON)
				assert.Empty(t, imageGenerationsResp.Images[0].URL)
			}

			assert.Equal(t, tt.expectedFormat, imageGenerationsResp.Images[0].ImageFormat)
			assert.NotNil(t, imageGenerationsResp.Images[0].ImageConfig)
			assert.Equal(t, tt.expectedWidth, imageGenerationsResp.Images[0].ImageConfig.Width)
			assert.Equal(t, tt.expectedHeight, imageGenerationsResp.Images[0].ImageConfig.Height)
		})
	}
}
