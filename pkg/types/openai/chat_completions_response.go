package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"knoway.dev/pkg/object"
	"knoway.dev/pkg/utils"
)

var _ object.LLMResponse = (*ChatCompletionsResponse)(nil)

type ChatCompletionsResponse struct {
	Status int                   `json:"status"`
	Model  string                `json:"model"`
	Usage  *ChatCompletionsUsage `json:"usage,omitempty"`
	Error  *ErrorResponse        `json:"error,omitempty"`
	Stream bool                  `json:"stream"`

	request          object.LLMRequest
	responseBody     json.RawMessage
	bodyParsed       map[string]any
	outgoingResponse *http.Response
}

func NewChatCompletionResponse(request object.LLMRequest, response *http.Response, reader *bufio.Reader) (*ChatCompletionsResponse, error) {
	resp := new(ChatCompletionsResponse)

	buffer := new(bytes.Buffer)

	_, err := buffer.ReadFrom(reader)
	if err != nil {
		return nil, err
	}

	err = resp.processBytes(buffer.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	resp.request = request
	resp.outgoingResponse = response

	return resp, nil
}

func (r *ChatCompletionsResponse) processBytes(bs []byte, response *http.Response) error {
	if r == nil {
		return nil
	}

	r.responseBody = bs
	r.Status = response.StatusCode

	var body map[string]any

	err := json.Unmarshal(bs, &body)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	r.bodyParsed = body

	r.Model = utils.GetByJSONPath[string](body, "{ .model }")
	usageMap := utils.GetByJSONPath[map[string]any](body, "{ .usage }")

	r.Usage, err = utils.FromMap[ChatCompletionsUsage](usageMap)
	if err != nil {
		return fmt.Errorf("failed to unmarshal usage: %w", err)
	}

	errorResponse, err := unmarshalErrorResponseFromParsedBody(body, response, bs)
	if err != nil {
		return err
	}
	if errorResponse != nil {
		r.Error = errorResponse
	}

	return nil
}

func (r *ChatCompletionsResponse) MarshalJSON() ([]byte, error) {
	return r.responseBody, nil
}

func (r *ChatCompletionsResponse) IsStream() bool {
	return false
}

func (r *ChatCompletionsResponse) GetRequestID() string {
	// TODO: implement
	return ""
}

func (r *ChatCompletionsResponse) GetModel() string {
	return r.Model
}

func (r *ChatCompletionsResponse) SetModel(model string) error {
	if r.Error == nil {
		var err error

		r.responseBody, r.bodyParsed, err = modifyBytesBodyAndParsed(r.responseBody, NewReplace("/model", model))
		if err != nil {
			return err
		}
	}

	r.Model = model

	return nil
}

func (r *ChatCompletionsResponse) GetUsage() object.LLMUsage {
	return r.Usage
}

func (r *ChatCompletionsResponse) GetError() object.LLMError {
	if r.Error != nil {
		return r.Error
	}

	return nil
}
