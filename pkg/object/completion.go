package object

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"knoway.dev/pkg/types/sse"
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
	json.Marshaler

	IsStream() bool
	GetRequestID() string
	GetModel() string
	GetUsage() *Usage
	GetOutgoingResponse() *http.Response
	GetError() error
}

type LLMStreamResponse interface {
	LLMResponse

	IsEOF() bool
	NextChunk() (LLMChunkResponse, error)
}

type LLMChunkResponse interface {
	json.Marshaler

	IsEmpty() bool
	IsDone() bool
	GetResponse() LLMStreamResponse
	ToServerSentEvent() (*sse.Event, error)
}

type Usage struct {
	CompletionTokens int `json:"completion_tokens,omitempty"`
	PromptTokens     int `json:"prompt_tokens,omitempty"`
}

type RequestHeader struct {
	APIKey string `json:"APIKey"`
}
