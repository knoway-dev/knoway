package object

import "encoding/json"

type LLMError interface {
	json.Marshaler
	json.Unmarshaler
	error

	GetCode() string
	GetMessage() string
	GetStatus() int
}

func IsLLMError(err error) bool {
	// Assert with interface, cannot use errors.As
	_, ok := err.(LLMError) //nolint:errorlint
	return ok
}

func AsLLMError(err error) LLMError {
	if IsLLMError(err) {
		llmError, _ := err.(LLMError) //nolint:errorlint
		return llmError
	}

	return nil
}
