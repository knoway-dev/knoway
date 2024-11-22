package openai

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestSetModel(t *testing.T) {
	httpRequest, err := http.NewRequest("POST", "http://example.com", bytes.NewBuffer([]byte(`
{
    "model": "some",
    "messages": [
        {
            "role": "user",
            "content": "hi"
        }
    ]
}
`)))
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}

	request, err := NewChatCompletionRequest(httpRequest)
	if err != nil {
		t.Fatalf("Failed to create ChatCompletionRequest: %v", err)
	}

	newModel := "gpt-4"
	err = request.SetModel(newModel)
	if err != nil {
		t.Fatalf("SetModel returned an error: %v", err)
	}

	if request.GetModel() != newModel {
		t.Errorf("Expected model to be %s, got %s", newModel, request.GetModel())
	}

	// Verify the body buffer has been updated
	var body map[string]any
	if err := json.Unmarshal(request.GetBodyBuffer().Bytes(), &body); err != nil {
		t.Fatalf("Failed to unmarshal body: %v", err)
	}

	if body["model"] != newModel {
		t.Errorf("Expected model in body to be %s, got %v", newModel, body["model"])
	}
	messages := []map[string]any{
		{
			"role":    "user",
			"content": "hi",
		},
	}
	newMessages, ok := body["messages"].([]interface{})
	if !ok || len(newMessages) != len(messages) {
		t.Errorf("Expected messages in body to be %v, got %v", messages, newMessages)
	} else {
		for i, msg := range messages {
			if msg["role"] != newMessages[i].(map[string]interface{})["role"] || msg["content"] != newMessages[i].(map[string]interface{})["content"] {
				t.Errorf("Expected message %d to be %v, got %v", i, msg, newMessages[i])
			}
		}
	}

}
