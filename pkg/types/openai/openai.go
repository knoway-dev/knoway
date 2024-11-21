package openai

import (
	"bytes"
	"encoding/json"
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

	bodyParsed         map[string]any
	bodyBuffer         *bytes.Buffer
	rawIncomingRequest *http.Request
}

func NewChatCompletionRequest(buffer *bytes.Buffer, parsed map[string]any, rawIncomingRequest *http.Request) *ChatCompletionRequest {
	return &ChatCompletionRequest{
		Model:              utils.GetByJSONPath[string](parsed, "{ .model }"),
		Stream:             utils.GetByJSONPath[bool](parsed, "{ .stream }"),
		bodyParsed:         parsed,
		bodyBuffer:         buffer,
		rawIncomingRequest: rawIncomingRequest,
	}
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
	return r.rawIncomingRequest
}
