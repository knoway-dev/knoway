package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/samber/lo"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/utils"
)

type Error struct {
	Code    *string `json:"code"`
	Message string  `json:"message"`
	Param   *string `json:"param"`
	Type    string  `json:"type"`
}

type Event string

const (
	EventError Event = "error"
)

type ErrorEvent struct {
	Event Event `json:"event"`
	Error Error `json:"error"`
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

func (e *ErrorResponse) WithCause(err error) *ErrorResponse {
	e.Cause = err
	return e
}

var _ object.LLMRequest = (*ChatCompletionRequest)(nil)

type ChatCompletionRequest struct {
	Model  string `json:"model,omitempty"`
	Stream bool   `json:"stream,omitempty"`

	// TODO: add more fields

	bodyParsed      map[string]any
	bodyBuffer      *bytes.Buffer
	incomingRequest *http.Request
}

func NewChatCompletionRequest(httpRequest *http.Request) (*ChatCompletionRequest, error) {
	buffer, parsed, err := utils.ReadAsJSONWithClose(httpRequest.Body)
	if err != nil {
		return nil, NewErrorInvalidBody()
	}

	return &ChatCompletionRequest{
		Model:           utils.GetByJSONPath[string](parsed, "{ .model }"),
		Stream:          utils.GetByJSONPath[bool](parsed, "{ .stream }"),
		bodyParsed:      parsed,
		bodyBuffer:      buffer,
		incomingRequest: httpRequest,
	}, nil
}

func (r *ChatCompletionRequest) IsStream() bool {
	return r.Stream
}

func (r *ChatCompletionRequest) GetModel() string {
	return r.Model
}

func (r *ChatCompletionRequest) SetModel(model string) error {
	patch, err := jsonpatch.DecodePatch(lo.Must(json.Marshal([]map[string]any{
		{"op": "replace", "path": "/model", "value": model},
	})))
	if err != nil {
		return err
	}

	patched, err := patch.Apply(r.bodyBuffer.Bytes())
	if err != nil {
		return err
	}

	r.bodyBuffer = bytes.NewBuffer(patched)

	var m map[string]any

	err = json.Unmarshal(patched, &m)
	if err != nil {
		return err
	}

	r.Model = model

	return nil
}

func (r *ChatCompletionRequest) GetBody() io.Reader {
	return bytes.NewBuffer(r.bodyBuffer.Bytes())
}

func (r *ChatCompletionRequest) GetBodyBuffer() *bytes.Buffer {
	return r.bodyBuffer
}

func (r *ChatCompletionRequest) GetBodyParsed() map[string]any {
	return r.bodyParsed
}

func (r *ChatCompletionRequest) GetIncomingRequest() *http.Request {
	return r.incomingRequest
}

var _ object.LLMResponse = (*ChatCompletionResponse)(nil)

type ChatCompletionResponse struct {
	Model string         `json:"model"`
	Usage *object.Usage  `json:"usage,omitempty"`
	Error *ErrorResponse `json:"error,omitempty"`

	// TODO: add more fields

	responseBody     json.RawMessage
	unmarshalledBody map[string]any
	outgoingResponse *http.Response
}

func NewChatCompletionResponse(response *http.Response, buffer *bytes.Buffer) (*ChatCompletionResponse, error) {
	resp := new(ChatCompletionResponse)

	err := resp.UnmarshalJSON(buffer.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.Error != nil {
		resp.Error.Status = response.StatusCode
	}

	resp.outgoingResponse = response

	return resp, nil
}

func (r *ChatCompletionResponse) UnmarshalJSON(bs []byte) error {
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
		respErr, err := utils.FromMap[Error](respErrMap)
		if err != nil {
			return fmt.Errorf("failed to unmarshal error: %w", err)
		}

		r.Error = &ErrorResponse{
			FromUpstream: true,
			ErrorBody:    respErr,
		}
	}

	return nil
}

func (r *ChatCompletionResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.responseBody)
}

func (r *ChatCompletionResponse) IsStream() bool {
	// TODO: implement
	return false
}

func (r *ChatCompletionResponse) GetRequestID() string {
	// TODO: implement
	return ""
}

func (r *ChatCompletionResponse) GetModel() string {
	return r.Model
}

func (r *ChatCompletionResponse) GetUsage() *object.Usage {
	return r.Usage
}

func (r *ChatCompletionResponse) GetOutgoingResponse() *http.Response {
	return r.outgoingResponse
}

func (r *ChatCompletionResponse) GetError() error {
	if r.Error != nil {
		return r.Error
	}

	return nil
}
