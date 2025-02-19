package openai

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"knoway.dev/pkg/object"
)

func TestNewChatCompletionStreamChunk(t *testing.T) {
	mockResp := &ChatCompletionStreamResponse{}
	testCases := []struct {
		name      string
		input     []byte
		model     string
		wantEmpty bool
		wantError bool
	}{
		{
			name:      "Valid chunk",
			input:     []byte(`{"model":"gpt-4","content":"test"}`),
			wantEmpty: false,
			wantError: false,
			model:     "gpt-4",
		},
		{
			name:      "Invalid JSON",
			input:     []byte(`{invalid json}`),
			wantEmpty: true,
			wantError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			chunk, err := NewChatCompletionStreamChunk(mockResp, tc.input)
			if tc.wantError {
				require.Error(t, err)
				assert.True(t, chunk.IsEmpty())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.model, chunk.GetModel())
				assert.False(t, chunk.IsEmpty())
			}
		})
	}
}

func TestNewUsageChatCompletionStreamChunk(t *testing.T) {
	mockResp := &ChatCompletionStreamResponse{}
	validUsage := []byte(`{
		"model": "gpt-4",
		"usage": {
			"prompt_tokens": 10,
			"completion_tokens": 20,
			"total_tokens": 30
		}
	}`)

	chunk, err := NewUsageChatCompletionStreamChunk(mockResp, validUsage)
	require.NoError(t, err)
	assert.True(t, chunk.IsUsage())
	assert.Equal(t, "gpt-4", chunk.GetModel())
	assert.NotNil(t, chunk.Usage)
	assert.Equal(t, uint64(10), chunk.Usage.PromptTokens)
	assert.Equal(t, uint64(20), chunk.Usage.CompletionTokens)
	assert.Equal(t, uint64(30), chunk.Usage.TotalTokens)
}

func TestChatCompletionStreamResponse_NextChunk(t *testing.T) {
	testCases := []struct {
		name            string
		input           string
		wantDone        bool
		wantEmpty       bool
		wantUsage       bool
		wantError       bool
		wantErrorPrefix bool
		expectedError   error
		wantContent     string
		wantStreamModel string
	}{
		{
			name:            "Valid chunk",
			input:           "data: {\"model\":\"gpt-4\",\"content\":\"test\"}\n",
			wantContent:     "{\"content\":\"test\",\"model\":\"gpt-4\"}",
			wantStreamModel: "gpt-4",
		},
		{
			name:          "Done message",
			input:         "data: [DONE]\n",
			wantDone:      true,
			wantError:     true,
			expectedError: io.EOF,
		},
		{
			name:            "Usage message",
			input:           "data: {\"model\":\"gpt-4\",\"usage\":{\"prompt_tokens\":10,\"completion_tokens\":20,\"total_tokens\":30}}\n",
			wantUsage:       true,
			wantStreamModel: "gpt-4",
		},
		{
			name:      "Empty line",
			input:     "\n",
			wantEmpty: true,
		},
		{
			name:      "Invalid JSON",
			input:     "data: {invalid-json}\n",
			wantEmpty: true,
			wantError: true,
		},
		{
			name:            "Error prefix",
			input:           "data: {\"error\":{\"message\":\"error occurred\",\"type\":\"invalid_request_error\"}}\n",
			wantEmpty:       true,
			wantErrorPrefix: true,
		},
		{
			name:      "No data prefix",
			input:     "raw: {\"content\":\"test\"}\n",
			wantEmpty: true,
		},
		{
			name:            "Multiple lines with error",
			input:           "data: {\"error\":{\"message\":\"error occurred\"}}\ndata: additional error info\n",
			wantErrorPrefix: true,
			wantEmpty:       true,
		},
		{
			name:          "EOF condition",
			input:         "",
			wantEmpty:     true,
			wantError:     true,
			expectedError: io.EOF,
		},
		{
			name:            "Error buffer write failure simulation",
			input:           "data: {\"error\":\"test\"}\n",
			wantEmpty:       true,
			wantErrorPrefix: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := bufio.NewReaderSize(strings.NewReader(tc.input), 2048*2*2*2)

			stream, err := NewChatCompletionStreamResponse(&ChatCompletionsRequest{}, nil, reader)
			require.NoError(t, err)

			chunk, err := stream.NextChunk()
			if tc.wantError {
				require.Error(t, err)
				if tc.expectedError != nil {
					assert.Equal(t, tc.expectedError, err)
				}
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.wantDone, chunk.IsDone())
			assert.Equal(t, tc.wantEmpty, chunk.IsEmpty())
			assert.Equal(t, tc.wantUsage, chunk.IsUsage())
			assert.Equal(t, tc.wantErrorPrefix, stream.hasErrorPrefix)

			if tc.wantContent != "" {
				data, err := json.Marshal(chunk)
				require.NoError(t, err)
				assert.Equal(t, tc.wantContent, string(data))
			}

			if tc.wantStreamModel != "" {
				assert.Equal(t, tc.wantStreamModel, stream.GetModel())
			}

			// Test error buffer content when error prefix is detected
			if tc.wantErrorPrefix {
				assert.NotEmpty(t, stream.errorEventBuffer.String())
			}
		})
	}
}

func TestChatCompletionStreamChunk_MarshalJSON(t *testing.T) {
	testCases := []struct {
		name     string
		chunk    *ChatCompletionStreamChunk
		expected string
	}{
		{
			name: "Regular chunk",
			chunk: &ChatCompletionStreamChunk{
				Model:      "gpt-4",
				bodyParsed: map[string]any{"model": "gpt-4", "content": "test"},
			},
			expected: `{"content":"test","model":"gpt-4"}`,
		},
		{
			name: "Done chunk",
			chunk: &ChatCompletionStreamChunk{
				isDone: true,
			},
			expected: `[DONE]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := tc.chunk.MarshalJSON()
			require.NoError(t, err)
			assert.Equal(t, tc.expected, string(data))
		})
	}
}

func TestChatCompletionStreamResponse_IsStream(t *testing.T) {
	stream := &ChatCompletionStreamResponse{}
	assert.True(t, stream.IsStream())
}

func TestChatCompletionStreamResponse_GetSetModel(t *testing.T) {
	stream := &ChatCompletionStreamResponse{}
	model := "gpt-4"

	err := stream.SetModel(model)
	require.NoError(t, err)
	assert.Equal(t, model, stream.GetModel())
}

func TestChatCompletionStreamChunk_SetModel(t *testing.T) {
	chunk := &ChatCompletionStreamChunk{
		responseBody: lo.Must(json.Marshal(map[string]any{"model": "gpt-3"})),
		bodyParsed:   map[string]any{"model": "gpt-3"},
	}
	model := "gpt-4"

	err := chunk.SetModel(model)
	require.NoError(t, err)
	assert.Equal(t, model, chunk.GetModel())
}

func TestChatCompletionStreamChunk_ToServerSentEvent(t *testing.T) {
	chunk := &ChatCompletionStreamChunk{
		Model: "gpt-4",
		bodyParsed: map[string]any{
			"model":   "gpt-4",
			"content": "test",
		},
	}

	event, err := chunk.ToServerSentEvent()
	require.NoError(t, err)
	assert.Equal(t, lo.Must(json.Marshal(chunk.bodyParsed)), event.Data)
}

func TestChatCompletionStreamResponse_GetUsage(t *testing.T) {
	stream := &ChatCompletionStreamResponse{
		Usage: &ChatCompletionsUsage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}

	usage := stream.GetUsage()
	assert.NotNil(t, usage)

	tokenUsage, ok := object.AsLLMTokensUsage(usage)
	require.True(t, ok)

	assert.Equal(t, uint64(10), tokenUsage.GetPromptTokens())
	assert.Equal(t, uint64(20), tokenUsage.GetCompletionTokens())
	assert.Equal(t, uint64(30), tokenUsage.GetTotalTokens())
}
