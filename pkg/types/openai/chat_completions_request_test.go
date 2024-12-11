package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
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

	err = json.Unmarshal(lo.Must(json.Marshal(request)), &body)
	require.NoError(t, err)
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

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, "/api/v1", bytes.NewReader(body))
	require.NoError(t, err)

	chatReq, err := NewChatCompletionRequest(req)
	require.NoError(t, err)

	params := map[string]*structpb.Value{
		"model":       structpb.NewStringValue("openai/gpt-4"),
		"stream":      structpb.NewBoolValue(true),
		"temperature": structpb.NewNumberValue(0.7),
		"max_tokens":  structpb.NewNumberValue(100),
	}

	err = chatReq.SetDefaultParams(params)
	require.NoError(t, err)

	assert.Equal(t, false, chatReq.bodyParsed["stream"])
	assert.Equal(t, "gpt-4", chatReq.bodyParsed["model"])
	assert.InDelta(t, 0.7, chatReq.bodyParsed["temperature"], 0.0001)
	assert.InDelta(t, 100.0, chatReq.bodyParsed["max_tokens"], 0.0001)
}

func TestSetOverrideParams(t *testing.T) {
	body := []byte(`{
		"model": "gpt-4",
		"stream": false,
		"temperature": 0.5,
		"max_tokens": 200
	}`)

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, "/api/v1", bytes.NewReader(body))
	require.NoError(t, err)

	chatReq, err := NewChatCompletionRequest(req)
	require.NoError(t, err)

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
	require.NoError(t, err)

	assert.Equal(t, "openai/gpt-4", chatReq.bodyParsed["model"])
	assert.InDelta(t, 0.7, chatReq.bodyParsed["temperature"], 0.0001)
	assert.InDelta(t, 100.0, chatReq.bodyParsed["max_tokens"], 0.0001)

	assert.Equal(t, true, chatReq.bodyParsed["stream"])
	assert.Equal(t, map[string]any{
		"include_usage": true,
	}, chatReq.bodyParsed["stream_options"])
}
