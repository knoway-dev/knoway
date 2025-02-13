package openai

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestImageGenerationsSetDefaultParams(t *testing.T) {
	body := []byte(`{
		"model": "public/sd-3",
		"n": 3,
		"size": "1024x1792"
	}`)

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, "/api/v1", bytes.NewReader(body))
	require.NoError(t, err)

	chatReq, err := NewImageGenerationsRequest(req)
	require.NoError(t, err)

	params := map[string]*structpb.Value{
		"model":   structpb.NewStringValue("openai/dall-e-3"),
		"n":       structpb.NewNumberValue(1),
		"style":   structpb.NewStringValue("natural"),
		"quality": structpb.NewStringValue("hd"),
		"size":    structpb.NewStringValue("1792x1024"),
	}

	err = chatReq.SetDefaultParams(params)
	require.NoError(t, err)

	assert.Equal(t, "public/sd-3", chatReq.bodyParsed["model"])
	assert.InDelta(t, 3.0, chatReq.bodyParsed["n"], 0.0001)
	assert.Equal(t, "natural", chatReq.bodyParsed["style"])
	assert.Equal(t, "hd", chatReq.bodyParsed["quality"])
	assert.Equal(t, "1024x1792", chatReq.bodyParsed["size"])
}

func TestImageGenerationsSetOverrideParams(t *testing.T) {
	body := []byte(`{
		"model": "public/sd-3",
		"n": 3,
		"style": "natural"
	}`)

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, "/api/v1", bytes.NewReader(body))
	require.NoError(t, err)

	chatReq, err := NewImageGenerationsRequest(req)
	require.NoError(t, err)

	params := map[string]*structpb.Value{
		"model": structpb.NewStringValue("openai/dall-e-3"),
		"n":     structpb.NewNumberValue(1),
		"size":  structpb.NewStringValue("1792x1024"),
	}

	err = chatReq.SetOverrideParams(params)
	require.NoError(t, err)

	assert.Equal(t, "openai/dall-e-3", chatReq.bodyParsed["model"])
	assert.InDelta(t, 1.0, chatReq.bodyParsed["n"], 0.0001)
	assert.Equal(t, "natural", chatReq.bodyParsed["style"])
}
