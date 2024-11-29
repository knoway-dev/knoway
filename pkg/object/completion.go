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

	SetAPIKey(key string)
	SetUser(user string)
	GetUser() string
	GetAPIKey() string
}

type LLMResponse interface {
	json.Marshaler

	IsStream() bool
	GetRequestID() string
	GetModel() string
	GetUsage() LLMUsage
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
	IsUsage() bool
	GetResponse() LLMStreamResponse
	ToServerSentEvent() (*sse.Event, error)
}

type LLMUsage interface {
	GetTotalTokens() uint64
	GetCompletionTokens() uint64
	GetPromptTokens() uint64
}

type RequestHeader struct {
	APIKey string `json:"APIKey"`
}

var _ LLMRequest = (*BaseLLMRequest)(nil)

type BaseLLMRequest struct {
	Model string `json:"model,omitempty"`

	// auth info
	APIKey string `json:"api_key,omitempty"`
	User   string `json:"user,omitempty"`
}

func (r *BaseLLMRequest) IsStream() bool {
	return false
}

func (r *BaseLLMRequest) GetBody() io.Reader {
	return nil
}

func (r *BaseLLMRequest) GetBodyBuffer() *bytes.Buffer {
	return nil
}

func (r *BaseLLMRequest) GetIncomingRequest() *http.Request {
	return nil
}

func (r *BaseLLMRequest) SetUser(user string) {
	if r == nil {
		return
	}

	r.User = user
}

func (r *BaseLLMRequest) SetAPIKey(key string) {
	if r == nil {
		return
	}

	r.APIKey = key
}

func (r *BaseLLMRequest) GetUser() string {
	if r == nil {
		return ""
	}

	return r.User
}

func (r *BaseLLMRequest) GetAPIKey() string {
	if r == nil {
		return ""
	}

	return r.APIKey
}

func (r *BaseLLMRequest) SetModel(model string) error {
	if r == nil {
		return nil
	}

	r.Model = model

	return nil
}

func (r *BaseLLMRequest) GetModel() string {
	if r == nil {
		return ""
	}

	return r.Model
}
