package observation

import (
	"strings"

	"go.opentelemetry.io/otel/attribute"
)

// AttributeKey is a generic key that can be used for both metrics labels and tracing attributes.
type AttributeKey string

// AsAttribute converts the AttributeKey to an OpenTelemetry attribute.Key.
func (a AttributeKey) AsAttribute() attribute.Key {
	return attribute.Key(a)
}

// AsLabelKey converts the AttributeKey to an OpenTelemetry label key.
func (a AttributeKey) AsLabelKey() string {
	return formatLabel(string(a))
}

func formatLabel(label string) string {
	return strings.ReplaceAll(label, ".", "_")
}

var (
	LLMRequestType    = AttributeKey("llm.request.type")
	LLMRequestStream  = AttributeKey("llm.request.stream")
	LLMRequestModel   = AttributeKey("llm.request.model")
	LLMRequestHeaders = AttributeKey("llm.request.headers")

	LLMResponseModel        = AttributeKey("llm.response.model")
	LLMResponseCode         = AttributeKey("llm.response.code")
	LLMResponseErrorMessage = AttributeKey("llm.response.error_message")
	LLMResponseHeaders      = AttributeKey("llm.response.headers")
	LLMResponseDuration     = AttributeKey("llm.response.duration")

	LLMTokenType             = AttributeKey("llm.usage.token_type")
	LLMUsageTotalTokens      = AttributeKey("llm.usage.total_tokens")
	LLMUsageCompletionTokens = AttributeKey("llm.usage.completion_tokens")
	LLMUsagePromptTokens     = AttributeKey("llm.usage.prompt_tokens")

	KnowayAuthInfoAPIKey = AttributeKey("knoway.auth.apikey")
	KnowayAuthInfoUser   = AttributeKey("knoway.auth.user")
)

type LLMTokenTypeEnum string

const (
	PromptTokenType     LLMTokenTypeEnum = "prompt"
	CompletionTokenType LLMTokenTypeEnum = "completion"
)
