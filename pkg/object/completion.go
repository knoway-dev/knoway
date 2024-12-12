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
