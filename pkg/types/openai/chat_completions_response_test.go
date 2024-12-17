package openai

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsage(t *testing.T) {
	u := &Usage{
		TotalTokens:      1,
		CompletionTokens: 2,
		PromptTokens:     3,
	}

	assert.Equal(t, uint64(1), u.GetTotalTokens())
	assert.Equal(t, uint64(2), u.GetCompletionTokens())
	assert.Equal(t, uint64(3), u.GetPromptTokens())
}

func TestNewChatCompletionResponse(t *testing.T) {
	testCases := []struct {
		name          string
		responseBody  string
		statusCode    int
		expectError   bool
		expectedModel string
		expectedUsage *Usage
	}{
		{
			name: "Valid response",
			responseBody: `{
                "model": "gpt-4",
                "usage": {
                    "prompt_tokens": 10,
                    "completion_tokens": 20,
                    "total_tokens": 30
                }
            }`,
			statusCode:    200,
			expectedModel: "gpt-4",
			expectedUsage: &Usage{
				PromptTokens:     10,
				CompletionTokens: 20,
				TotalTokens:      30,
			},
		},
		{
			name:         "Invalid JSON",
			responseBody: `{invalid json}`,
			statusCode:   200,
			expectError:  true,
		},
		{
			name: "Error response with map",
			responseBody: `{
                "error": {
                    "message": "error occurred",
                    "type": "invalid_request_error"
                }
            }`,
			statusCode:  400,
			expectError: false,
		},
		{
			name: "Error response with string",
			responseBody: `{
                "error": "endpoint not found"
            }`,
			statusCode:  404,
			expectError: false,
		},
		{
			name:         "Error status without error body",
			responseBody: `{}`,
			statusCode:   500,
			expectError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, "http://localhost/v1/chat/completions", nil)
			require.NoError(t, err)

			resp := &http.Response{
				StatusCode: tc.statusCode,
				Request:    req,
				Body:       io.NopCloser(strings.NewReader(tc.responseBody)),
			}

			reader := bufio.NewReader(strings.NewReader(tc.responseBody))

			response, err := NewChatCompletionResponse(&ChatCompletionsRequest{}, resp, reader)
			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tc.expectedModel != "" {
				assert.Equal(t, tc.expectedModel, response.GetModel())
			}
			if tc.expectedUsage != nil {
				assert.Equal(t, tc.expectedUsage, response.GetUsage())
			}

			assert.Equal(t, tc.statusCode, response.Status)
		})
	}
}

func TestChatCompletionsResponse_Usage(t *testing.T) {
	usage := &Usage{
		TotalTokens:      30,
		CompletionTokens: 20,
		PromptTokens:     10,
		CompletionTokensDetails: &CompletionTokensDetails{
			AcceptedPredictionTokens: 5,
			AudioTokens:              2,
			ReasoningTokens:          8,
			RejectedPredictionTokens: 5,
		},
		PromptTokensDetails: &PromptTokensDetails{
			AudioTokens:  3,
			CachedTokens: 7,
		},
	}

	assert.Equal(t, uint64(30), usage.GetTotalTokens())
	assert.Equal(t, uint64(20), usage.GetCompletionTokens())
	assert.Equal(t, uint64(10), usage.GetPromptTokens())
}

func TestChatCompletionsResponse_SetModel(t *testing.T) {
	testCases := []struct {
		name        string
		initial     *ChatCompletionsResponse
		newModel    string
		expectError bool
	}{
		{
			name: "Valid model update",
			initial: &ChatCompletionsResponse{
				Model:        "gpt-3",
				responseBody: []byte(`{"model":"gpt-3"}`),
				bodyParsed:   map[string]any{"model": "gpt-3"},
			},
			newModel: "gpt-4",
		},
		{
			name: "Update with error present",
			initial: &ChatCompletionsResponse{
				Model: "gpt-3",
				Error: &ErrorResponse{},
			},
			newModel: "gpt-4",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.initial.SetModel(tc.newModel)
			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.newModel, tc.initial.GetModel())
		})
	}
}

func TestChatCompletionsResponse_MarshalJSON(t *testing.T) {
	responseBody := []byte(`{"model":"gpt-4"}`)
	response := &ChatCompletionsResponse{
		responseBody: responseBody,
	}

	data, err := response.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, responseBody, data)
}

func TestChatCompletionsResponse_IsStream(t *testing.T) {
	response := &ChatCompletionsResponse{}
	assert.False(t, response.IsStream())
}

func TestChatCompletionsResponse_GetError(t *testing.T) {
	testCases := []struct {
		name           string
		response       *ChatCompletionsResponse
		expectNilError bool
	}{
		{
			name: "With error",
			response: &ChatCompletionsResponse{
				Error: &ErrorResponse{
					Status: 400,
					ErrorBody: &Error{
						Message: "test error",
					},
				},
			},
			expectNilError: false,
		},
		{
			name:           "Without error",
			response:       &ChatCompletionsResponse{},
			expectNilError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.response.GetError()
			if tc.expectNilError {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}
