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

	SetApiKey(key string)
	SetUser(user string)
	SetAllowModels(allowModels []string)
	GetUser() string
	GetApiKey() string
	GetAllowModels() []string
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

var _ LLMRequest = (*BaseLLMRequest)(nil)

type BaseLLMRequest struct {
	Model string `json:"model,omitempty"`

	// auth info
	ApiKey      string   `json:"api_key,omitempty"`
	AllowModels []string `json:"allow_models,omitempty"`
	User        string   `json:"user,omitempty"`
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

func (r *BaseLLMRequest) SetAllowModels(allowModels []string) {
	if r == nil {
		return
	}
	r.AllowModels = allowModels
}

func (r *BaseLLMRequest) SetUser(user string) {
	if r == nil {
		return
	}
	r.User = user
}

func (r *BaseLLMRequest) SetApiKey(key string) {
	if r == nil {
		return
	}
	r.ApiKey = key
}

func (r *BaseLLMRequest) GetAllowModels() []string {
	if r == nil {
		return nil
	}
	return r.AllowModels
}

func (r *BaseLLMRequest) GetUser() string {
	if r == nil {
		return ""
	}
	return r.User
}

func (r *BaseLLMRequest) GetApiKey() string {
	if r == nil {
		return ""
	}
	return r.ApiKey
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
