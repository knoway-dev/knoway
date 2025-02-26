package object

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/samber/lo"

	"knoway.dev/pkg/utils"
)

type LLMErrorCode string

const (
	LLMErrorCodeModelNotFoundOrNotAccessible LLMErrorCode = "model_not_found"
	LLMErrorCodeModelAccessDenied            LLMErrorCode = "model_access_denied"
	LLMErrorCodeRateLimitExceeded            LLMErrorCode = "model_rate_limit_exceeded"
	LLMErrorCodeInsufficientQuota            LLMErrorCode = "insufficient_quota"
	LLMErrorCodeMissingAPIKey                LLMErrorCode = "missing_api_key"
	LLMErrorCodeIncorrectAPIKey              LLMErrorCode = "incorrect_api_key"
	LLMErrorCodeMissingModel                 LLMErrorCode = "missing_model"
	LLMErrorCodeServiceUnavailable           LLMErrorCode = "service_unavailable"
	LLMErrorCodeInternalError                LLMErrorCode = "internal_error"
	LLMErrorCodeBadGateway                   LLMErrorCode = "bad_gateway"
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

func NewErrorModelAccessDenied(model string) *BaseLLMError {
	return &BaseLLMError{
		Status: http.StatusForbidden,
		ErrorBody: &BaseError{
			Code:    lo.ToPtr(LLMErrorCodeModelAccessDenied),
			Message: fmt.Sprintf("You do not have access to the model `%s`.", model),
		},
	}
}

func NewErrorRateLimitExceeded() *BaseLLMError {
	return &BaseLLMError{
		Status: http.StatusTooManyRequests,
		ErrorBody: &BaseError{
			Code:    lo.ToPtr(LLMErrorCodeRateLimitExceeded),
			Message: "You have exceeded the rate limit. Please try again later.",
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

func NewErrorIncorrectAPIKey(apiKey string) *BaseLLMError {
	return &BaseLLMError{
		Status: http.StatusUnauthorized,
		ErrorBody: &BaseError{
			Code:    lo.ToPtr(LLMErrorCodeIncorrectAPIKey),
			Message: "Incorrect API key provided: " + apiKey,
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

func NewErrorInternalError(internalErrs ...error) *BaseLLMError {
	internalErrs = append(internalErrs, errors.New("internal error"))

	return &BaseLLMError{
		Status: http.StatusInternalServerError,
		ErrorBody: &BaseError{
			Code:    lo.ToPtr(LLMErrorCodeInternalError),
			Message: lo.Must(lo.Coalesce(internalErrs...)).Error(),
		},
	}
}

func NewErrorBadGateway(upstreamErr error) *BaseLLMError {
	return &BaseLLMError{
		Status: http.StatusBadGateway,
		ErrorBody: &BaseError{
			Code:    lo.ToPtr(LLMErrorCodeBadGateway),
			Message: lo.Must(lo.Coalesce(upstreamErr, errors.New("bad gateway"))).Error(),
		},
	}
}

func NewErrorServiceUnavailable() *BaseLLMError {
	return &BaseLLMError{
		Status: http.StatusServiceUnavailable,
		ErrorBody: &BaseError{
			Code:    lo.ToPtr(LLMErrorCodeServiceUnavailable),
			Message: "service unavailable",
		},
	}
}

func LLMErrorOrInternalError(anyErrs ...error) LLMError {
	anyErrs = lo.Filter(anyErrs, utils.FilterNonNil)

	for _, err := range anyErrs {
		if !IsLLMError(err) {
			continue
		}

		return AsLLMError(err)
	}

	return NewErrorInternalError(anyErrs...)
}
