package object

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/samber/lo"
)

type LLMErrorCode string

const (
	LLMErrorCodeModelNotFoundOrNotAccessible LLMErrorCode = "model_not_found"
	LLMErrorCodeInsufficientQuota            LLMErrorCode = "insufficient_quota"
	LLMErrorCodeMissingAPIKey                LLMErrorCode = "missing_api_key"
	LLMErrorCodeIncorrectAPIKey              LLMErrorCode = "incorrect_api_key"
	LLMErrorCodeMissingModel                 LLMErrorCode = "missing_model"
)

var _ LLMError = (*BaseLLMError)(nil)

type BaseError struct {
	Code    *LLMErrorCode `json:"code"`
	Message string        `json:"message"`
}

type BaseLLMError struct {
	Status    int        `json:"-"`
	ErrorBody *BaseError `json:"error"`
}

func (e *BaseLLMError) Error() string {
	return e.ErrorBody.Message
}

func (e *BaseLLMError) GetCode() string {
	return string(lo.FromPtrOr(e.ErrorBody.Code, ""))
}

func (e *BaseLLMError) GetMessage() string {
	return e.ErrorBody.Message
}

func (e *BaseLLMError) GetStatus() int {
	return e.Status
}

func (e *BaseLLMError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"error": map[string]any{
			"code":    e.ErrorBody.Code,
			"message": e.ErrorBody.Message,
		},
	})
}

func (e *BaseLLMError) UnmarshalJSON(data []byte) error {
	e.ErrorBody = new(BaseError)

	var err error
	var parsed map[string]any

	err = json.Unmarshal(data, &parsed)
	if err != nil {
		return err
	}

	errorMapRaw, ok := parsed["error"]
	if !ok {
		return nil
	}

	errorMap, ok := errorMapRaw.(map[string]any)
	if !ok {
		return nil
	}

	codeStr, ok := errorMap["code"].(string)
	if ok {
		e.ErrorBody.Code = lo.ToPtr(LLMErrorCode(codeStr))
	}

	msgStr, ok := errorMap["message"].(string)
	if ok {
		e.ErrorBody.Message = msgStr
	}

	return nil
}

func NewErrorModelNotFoundOrNotAccessible(model string) *BaseLLMError {
	return &BaseLLMError{
		Status: http.StatusNotFound,
		ErrorBody: &BaseError{
			Code:    lo.ToPtr(LLMErrorCodeModelNotFoundOrNotAccessible),
			Message: fmt.Sprintf("The model `%s` does not exist or you do not have access to it.", model),
		},
	}
}

func NewErrorInsufficientQuota() *BaseLLMError {
	return &BaseLLMError{
		Status: http.StatusPaymentRequired,
		ErrorBody: &BaseError{
			Code:    lo.ToPtr(LLMErrorCodeInsufficientQuota),
			Message: "You exceeded your current quota.",
		},
	}
}

func NewErrorMissingAPIKey() *BaseLLMError {
	return &BaseLLMError{
		Status: http.StatusUnauthorized,
		ErrorBody: &BaseError{
			Code:    lo.ToPtr(LLMErrorCodeMissingAPIKey),
			Message: "Missing API key",
		},
	}
}

func NewErrorIncorrectAPIKey() *BaseLLMError {
	return &BaseLLMError{
		Status: http.StatusUnauthorized,
		ErrorBody: &BaseError{
			Code:    lo.ToPtr(LLMErrorCodeIncorrectAPIKey),
			Message: "Incorrect API key provided",
		},
	}
}

func NewErrorMissingModel() *BaseLLMError {
	return &BaseLLMError{
		Status: http.StatusNotFound,
		ErrorBody: &BaseError{
			Code:    lo.ToPtr(LLMErrorCodeMissingModel),
			Message: "Missing required parameter: '" + "model" + "'.",
		},
	}
}
