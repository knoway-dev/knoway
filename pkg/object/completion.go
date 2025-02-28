package object

import (
	"context"
	"encoding/json"
	"net/http"

	structpb "github.com/golang/protobuf/ptypes/struct"

	"knoway.dev/pkg/types/sse"
)

type RequestType string

const (
	RequestTypeChatCompletions  RequestType = "chat_completions"
	RequestTypeCompletions      RequestType = "completions"
	RequestTypeImageGenerations RequestType = "image_generations"
)

type LLMRequest interface {
	IsStream() bool
	GetModel() string
	SetModel(modelName string) error

	SetOverrideParams(params map[string]*structpb.Value) error
	SetDefaultParams(params map[string]*structpb.Value) error

	GetRequestType() RequestType
	GetRawRequest() *http.Request
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
	WaitUntilEOF() <-chan LLMStreamResponse
	OnChunk(cb func(ctx context.Context, stream LLMStreamResponse, chunk LLMChunkResponse))
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
	isLLMUsage()
}

type LLMTokensUsage interface {
	LLMUsage

	GetTotalTokens() uint64
	GetCompletionTokens() uint64
	GetPromptTokens() uint64
}

func AsLLMTokensUsage(u LLMUsage) (LLMTokensUsage, bool) {
	t, ok := u.(LLMTokensUsage)
	return t, ok
}

var _ LLMUsage = (*IsLLMUsage)(nil)

type IsLLMUsage struct{}

func (IsLLMUsage) isLLMUsage() {}
