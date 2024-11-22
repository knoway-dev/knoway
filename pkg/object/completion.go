package object

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type LLMRequest interface {
	IsStream() bool
	GetModel() string
	GetBody() io.Reader
	GetBodyBuffer() *bytes.Buffer
	GetIncomingRequest() *http.Request

	SetModel(modelName string) error
}

type LLMResponse interface {
	json.Unmarshaler
	json.Marshaler

	IsStream() bool
	GetRequestID() string
	GetModel() string
	GetUsage() *Usage
	GetOutgoingResponse() *http.Response
	GetError() error
}

type Usage struct {
	CompletionTokens int `json:"completion_tokens,omitempty"`
	PromptTokens     int `json:"prompt_tokens,omitempty"`
}
