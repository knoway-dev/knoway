package object

import (
	"encoding/json"

	structpb "github.com/golang/protobuf/ptypes/struct"

	"knoway.dev/pkg/types/sse"
)

type RequestType string

const (
	RequestTypeChatCompletions RequestType = "chat_completions"
	RequestTypeCompletions     RequestType = "completions"
)

type LLMRequest interface {
	IsStream() bool
	GetModel() string
	SetModel(modelName string) error

	SetOverrideParams(params map[string]*structpb.Value) error
	SetDefaultParams(params map[string]*structpb.Value) error

	GetRequestType() RequestType
}

type LLMResponse interface {
	json.Marshaler

	IsStream() bool
	GetRequestID() string
	GetUsage() LLMUsage
	GetError() LLMError

	GetModel() string
	SetModel(modelName string) error
}

func IsLLMResponse(r any) bool {
	_, ok := r.(LLMResponse)
	return ok
}

type LLMStreamResponse interface {
	LLMResponse

	IsEOF() bool
	NextChunk() (LLMChunkResponse, error)
}

func IsLLMStreamResponse(r any) bool {
	_, ok := r.(LLMStreamResponse)
	if ok {
		return true
	}

	llmResp, ok := r.(LLMStreamResponse)

	return ok && llmResp.IsStream()
}

type LLMChunkResponse interface {
	json.Marshaler

	IsFirst() bool
	IsEmpty() bool
	IsDone() bool
	IsUsage() bool
	GetResponse() LLMStreamResponse

	GetModel() string
	SetModel(modelName string) error
	GetUsage() LLMUsage

	ToServerSentEvent() (*sse.Event, error)
}

type LLMUsage interface {
	GetTotalTokens() uint64
	GetCompletionTokens() uint64
	GetPromptTokens() uint64
}

var _ LLMUsage = (*DefaultLLMUsage)(nil)

type DefaultLLMUsage struct{}

func (u DefaultLLMUsage) GetPromptTokens() uint64 {
	return 0
}

func (u DefaultLLMUsage) GetCompletionTokens() uint64 {
	return 0
}

func (u DefaultLLMUsage) GetTotalTokens() uint64 {
	return 0
}
