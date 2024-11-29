package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

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
