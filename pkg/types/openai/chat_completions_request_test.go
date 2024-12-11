package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetModel(t *testing.T) {
	httpRequest, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "http://example.com", bytes.NewBufferString(`
{
    "model": "some",
    "messages": [
        {
            "role": "user",
            "content": "hi"
        }
    ]
}
`))
	require.NoError(t, err)

	request, err := NewChatCompletionRequest(httpRequest)
	require.NoError(t, err)

	newModel := "gpt-4"
	err = request.SetModel(newModel)
	require.NoError(t, err)

	assert.Equal(t, newModel, request.GetModel())

	// Verify the body buffer has been updated
	var body map[string]any
	if err := json.Unmarshal(request.GetBodyBuffer().Bytes(), &body); err != nil {
		t.Fatalf("Failed to unmarshal body: %v", err)
	}

	assert.Equal(t, newModel, body["model"])

	messages := []map[string]any{
		{
			"role":    "user",
			"content": "hi",
		},
	}

	newMessages, ok := body["messages"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, len(messages), len(newMessages))

	for i, msg := range messages {
		newMessageMap, ok := newMessages[i].(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, msg["role"], newMessageMap["role"])
		assert.Equal(t, msg["content"], newMessageMap["content"])
	}
}

func TestSetDefaultParams(t *testing.T) {
	body := []byte(`{
		"model": "gpt-4",
		"stream": false
	}`)

	req, err := http.NewRequest("POST", "/api/v1", bytes.NewReader(body))
	assert.NoError(t, err, "failed to create HTTP request")

	chatReq, err := NewChatCompletionRequest(req)
	assert.NoError(t, err, "failed to create ChatCompletionsRequest")

	params := map[string]*structpb.Value{
		"model":       structpb.NewStringValue("openai/gpt-4"),
		"stream":      structpb.NewBoolValue(true),
		"temperature": structpb.NewNumberValue(0.7),
		"max_tokens":  structpb.NewNumberValue(100),
	}

	err = chatReq.SetDefaultParams(params)
	assert.NoError(t, err, "SetDefaultParams failed")

	bodyParsed := chatReq.GetBodyParsed()
	assert.Equal(t, false, bodyParsed["stream"], "stream")
	assert.Equal(t, "gpt-4", bodyParsed["model"], "model")
	assert.Equal(t, 0.7, bodyParsed["temperature"], "temperature")
	assert.Equal(t, 100.0, bodyParsed["max_tokens"], "max_tokens")
}

func TestSetOverrideParams(t *testing.T) {
	body := []byte(`{
		"model": "gpt-4",
		"stream": false,
		"temperature": 0.5,
		"max_tokens": 200
	}`)

	req, err := http.NewRequest("POST", "/api/v1", bytes.NewReader(body))
	assert.NoError(t, err, "failed to create HTTP request")

	chatReq, err := NewChatCompletionRequest(req)
	assert.NoError(t, err, "failed to create ChatCompletionsRequest")

	params := map[string]*structpb.Value{
		"model":       structpb.NewStringValue("openai/gpt-4"),
		"stream":      structpb.NewBoolValue(true),
		"temperature": structpb.NewNumberValue(0.7),
		"max_tokens":  structpb.NewNumberValue(100),
		"stream_options": structpb.NewStructValue(&structpb.Struct{
			Fields: map[string]*structpb.Value{
				"include_usage": structpb.NewBoolValue(true),
			},
		}),
	}

	err = chatReq.SetOverrideParams(params)
	assert.NoError(t, err, "SetOverrideParams failed")

	bodyParsed := chatReq.GetBodyParsed()
	assert.Equal(t, "openai/gpt-4", bodyParsed["model"], "model")
	assert.Equal(t, 0.7, bodyParsed["temperature"], "temperature")
	assert.Equal(t, 100.0, bodyParsed["max_tokens"], "max_tokens")

	assert.Equal(t, true, bodyParsed["stream"], "stream")
	assert.Equal(t, map[string]interface{}{
		"include_usage": true,
	}, bodyParsed["stream_options"], "stream_options")
}
