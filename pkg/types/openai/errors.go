package openai

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/samber/lo"
)

type Error struct {
	Code    *string `json:"code"`
	Message string  `json:"message"`
	Param   *string `json:"param"`
	Type    string  `json:"type"`
}

type ErrorResponse struct { //nolint:errname
	Status       int    `json:"-"`
	FromUpstream bool   `json:"-"`
	ErrorBody    *Error `json:"error"`
	Cause        error  `json:"-"`
}

func (e *ErrorResponse) Error() string {
	return e.ErrorBody.Message
}

func (e *ErrorResponse) appendCause(err error) *ErrorResponse {
	if e.ErrorBody == nil {
		e.ErrorBody = &Error{}
	}
	if e.ErrorBody.Message != "" {
		e.ErrorBody.Message = fmt.Sprintf("%s:%s", e.ErrorBody.Message, err.Error())
	} else {
		e.ErrorBody.Message = err.Error()
	}

	return e
}

func (e *ErrorResponse) WithCause(err error) *ErrorResponse {
	e.Cause = err
	e.appendCause(err) //nolint:errcheck

	return e
}

func (e *ErrorResponse) WithCausef(format string, args ...interface{}) *ErrorResponse {
	e.WithCause(fmt.Errorf(format, args...)) //nolint:errcheck

	return e
}

func NewErrorResponse(status int, err Error) *ErrorResponse {
	return &ErrorResponse{
		Status:    status,
		ErrorBody: &err,
	}
}

/*
Example:

	{
		"error": {
			"message": "You didn't provide an API key. You need to provide your API key in an Authorization header using Bearer auth
				(i.e. Authorization: Bearer YOUR_KEY), or as the password field (with blank username) if you're accessing the API from
				your browser and are prompted for a username and password. You can obtain an API key from
				https://platform.openai.com/account/api-keys.",
			"type": "invalid_request_error",
			"param": null,
			"code": null
		}
	}
*/
func NewErrorMissingAPIKey() *ErrorResponse {
	return NewErrorResponse(http.StatusUnauthorized, Error{
		Message: "" +
			"You didn't provide an API key. You need to provide your API key in an Authorization header using Bearer auth " +
			"(i.e. Authorization: Bearer YOUR_KEY), or as the password field (with blank username) if you're accessing the " +
			"API from your browser and are prompted for a username and password. You can obtain an API key from " +
			"https://platform.openai.com/account/api-keys.",
		Type: "invalid_request_error",
	})
}

/*
Example:

	{
	    "error": {
	        "message": "Incorrect API key provided: sk-abcd. You can find your API key at https://platform.openai.com/account/api-keys.",
	        "type": "invalid_request_error",
	        "param": null,
	        "code": "invalid_api_key"
	    }
	}
*/
func NewErrorIncorrectAPIKey() *ErrorResponse {
	return NewErrorResponse(http.StatusUnauthorized, Error{
		Message: "Incorrect API key provided: sk-abcd. You can find your API key at https://platform.openai.com/account/api-keys.",
		Type:    "invalid_request_error",
		Code:    lo.ToPtr("invalid_api_key"),
	})
}

func NewErrorBadRequest() *ErrorResponse {
	return NewErrorResponse(
		http.StatusBadRequest, Error{
			Message: "bad request",
			Type:    "invalid_request_error",
		},
	)
}

/*
Example:

	{
	    "error": {
	        "message": "you must provide a model parameter",
	        "type": "invalid_request_error",
	        "param": null,
	        "code": null
	    }
	}
*/
func NewErrorMissingModel() *ErrorResponse {
	return NewErrorResponse(
		http.StatusBadRequest, Error{
			Message: "you must provide a model parameter",
			Type:    "invalid_request_error",
		},
	)
}

/*
Example:

	{
	    "error": {
	        "message": "We could not parse the JSON body of your request.
				(HINT: This likely means you aren't using your HTTP
				library correctly. The OpenAI API expects a JSON
				payload, but what was sent was not valid JSON. If you
				have trouble figuring out how to fix this, please
				contact us through our help center at help.openai.com.)",
	        "type": "invalid_request_error",
	        "param": null,
	        "code": null
	    }
	}
*/
func NewErrorInvalidBody() *ErrorResponse {
	return NewErrorResponse(http.StatusBadRequest, Error{
		Message: "" +
			"We could not parse the JSON body of your request. " +
			"(HINT: This likely means you aren't using your HTTP " +
			"library correctly. The OpenAI API expects a JSON " +
			"payload, but what was sent was not valid JSON. If you " +
			"have trouble figuring out how to fix this, please " +
			"contact us through our help center at help.openai.com.)",
		Type: "invalid_request_error",
	})
}

/*
Example:

	{
	    "error": {
	        "message": "Invalid URL (POST /v1/chat/not-found)",
	        "type": "invalid_request_error",
	        "param": null,
	        "code": null
	    }
	}
*/
func NewErrorNotFound(method string, url string) *ErrorResponse {
	return NewErrorResponse(http.StatusNotFound, Error{
		Message: fmt.Sprintf("Invalid URL (%s %s)", strings.ToUpper(method), url),
		Type:    "invalid_request_error",
	})
}

/*
Example:

	{
	    "error": {
	        "message": "The model `abcd` does not exist or you do not have access to it.",
	        "type": "invalid_request_error",
	        "param": null,
	        "code": "model_not_found"
	    }
	}
*/
func NewErrorModelNotFoundOrNotAccessible(model string) *ErrorResponse {
	return NewErrorResponse(http.StatusNotFound, Error{
		Message: fmt.Sprintf("The model `%s` does not exist or you do not have access to it.", model),
		Type:    "invalid_request_error",
		Code:    lo.ToPtr("model_not_found"),
	})
}

/*
Example:

	{
	    "error": {
	        "message": "You exceeded your current quota, please check your plan and billing details.
				For more information on this error, read the docs:
				https://platform.openai.com/docs/guides/error-codes/api-errors.",
	        "type": "insufficient_quota",
	        "param": null,
	        "code": "insufficient_quota"
	    }
	}
*/
func NewErrorQuotaExceeded() *ErrorResponse {
	return NewErrorResponse(http.StatusTooManyRequests, Error{
		Message: "" +
			"You exceeded your current quota, please check your plan and billing details. " +
			"For more information on this error, read the docs: " +
			"https://platform.openai.com/docs/guides/error-codes/api-errors.",
		Type: "insufficient_quota",
		Code: lo.ToPtr("insufficient_quota"),
	})
}

/*
Example:

	{
	  "error": {
	    "message": "Missing required parameter: 'messages'.",
	    "type": "invalid_request_error",
	    "param": "messages",
	    "code": "missing_required_parameter"
	  }
	}
*/
func NewErrorMissingParameter(parameter string) *ErrorResponse {
	return NewErrorResponse(http.StatusBadRequest, Error{
		Message: "Missing required parameter: '" + parameter + "'.",
		Type:    "invalid_request_error",
		Param:   lo.ToPtr(parameter),
		Code:    lo.ToPtr("missing_required_parameter"),
	})
}

func NewErrorInternalError() *ErrorResponse {
	return NewErrorResponse(http.StatusInternalServerError, Error{
		Message: "internal error",
		Type:    "internal_error",
	})
}
