package object

import "encoding/json"

type LLMRequest interface {
	GetModel() string
}

type LLMResponse interface {
	IsStream() bool
	GetRequestID() string
	GetModel() string
	GetUsage() *Usage
	json.Unmarshaler
	json.Marshaler
}

type Usage struct {
	CompletionTokens int `json:"completion_tokens,omitempty"`
	PromptTokens     int `json:"prompt_tokens,omitempty"`
}
