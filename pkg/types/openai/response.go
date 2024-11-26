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
	Model  string         `json:"model"`
	Usage  *object.Usage  `json:"usage,omitempty"`
	Error  *ErrorResponse `json:"error,omitempty"`
	Stream bool           `json:"stream"`

	// TODO: add more fields

	request          object.LLMRequest
	responseBody     json.RawMessage
	unmarshalledBody map[string]any
	outgoingResponse *http.Response
}

func NewChatCompletionResponse(request object.LLMRequest, response *http.Response, reader *bufio.Reader) (*ChatCompletionsResponse, error) {
	resp := new(ChatCompletionsResponse)

	buffer := new(bytes.Buffer)

	_, err := buffer.ReadFrom(reader)
	if err != nil {
		return nil, err
	}

	err = resp.processBytes(buffer.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.Error != nil {
		resp.Error.Status = response.StatusCode
	}

	resp.request = request
	resp.outgoingResponse = response

	return resp, nil
}

func (r *ChatCompletionsResponse) processBytes(bs []byte) error {
	r.responseBody = bs

	var body map[string]any

	err := json.Unmarshal(bs, &body)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	r.unmarshalledBody = body

	r.Model = utils.GetByJSONPath[string](body, "{ .model }")
	usageMap := utils.GetByJSONPath[map[string]any](body, "{ .usage }")
	respErrMap := utils.GetByJSONPath[map[string]any](body, "{ .error }")

	usage, err := utils.FromMap[object.Usage](usageMap)
	if err != nil {
		return fmt.Errorf("failed to unmarshal usage: %w", err)
	}

	r.Usage = usage

	if len(respErrMap) > 0 {
		respErr, err := utils.FromMap[ResponseError](respErrMap)
		if err != nil {
			return fmt.Errorf("failed to unmarshal error: %w", err)
		}

		code := fmt.Sprintf("%d", respErr.Code)
		rErr := &Error{
			Code:    &code,
			Message: respErr.Message,
			Param:   respErr.Param,
			Type:    respErr.Type,
		}

		r.Error = &ErrorResponse{
			FromUpstream: true,
			ErrorBody:    rErr,
		}
	}

	return nil
}

func (r *ChatCompletionsResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.responseBody)
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

func (r *ChatCompletionsResponse) GetUsage() *object.Usage {
	return r.Usage
}

func (r *ChatCompletionsResponse) GetOutgoingResponse() *http.Response {
	return r.outgoingResponse
}

func (r *ChatCompletionsResponse) GetError() error {
	if r.Error != nil {
		return r.Error
	}

	return nil
}
