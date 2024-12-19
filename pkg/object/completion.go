package object

import (
	"encoding/json"
	"net/http"

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
	GetRequest() *http.Request
}

func IsLLMRequest(r any) bool {
	_, ok := r.(LLMRequest)
	return ok
}

func IsLLMStreamRequest(r any) bool {
	llmReq, ok := r.(LLMRequest)
	return ok && llmReq.IsStream()
}

type LLMResponse interface {
	json.Marshaler

	IsStream() bool
	GetRequestID() string
	GetUsage() LLMUsage
	GetError() LLMError

	GetModel() string
	SetModel(modelName string) error

	GetResponse() *http.Response
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

	ToServerSentEvent() (*sse.Event, error)
}

func IsLLMChunkResponse(r any) bool {
	_, ok := r.(LLMChunkResponse)
	return ok
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
