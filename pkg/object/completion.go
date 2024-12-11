package object

import (
	"encoding/json"

	structpb "github.com/golang/protobuf/ptypes/struct"

	"knoway.dev/pkg/types/sse"
)

type LLMRequest interface {
	IsStream() bool
	GetModel() string
	SetModel(modelName string) error

	SetOverrideParams(params map[string]*structpb.Value) error
	SetDefaultParams(params map[string]*structpb.Value) error
}

type LLMResponse interface {
	json.Marshaler

	IsStream() bool
	GetRequestID() string
	GetUsage() LLMUsage
	GetError() error

	GetModel() string
	SetModel(modelName string) error
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

	GetModel() string
	SetModel(modelName string) error

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

func (r *BaseLLMRequest) SetDefaultParams(params map[string]*structpb.Value) error {
	return nil
}

func (r *BaseLLMRequest) SetOverrideParams(params map[string]*structpb.Value) error {
	return nil
}

func (r *BaseLLMRequest) IsStream() bool {
	return false
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
