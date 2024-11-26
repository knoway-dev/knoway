package openai

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	jsonpatch "github.com/evanphx/json-patch/v5"

	"knoway.dev/pkg/object"
	"knoway.dev/pkg/utils"
)

var _ object.LLMRequest = (*ChatCompletionRequest)(nil)

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type ChatCompletionRequest struct {
	Model         string        `json:"model,omitempty"`
	Stream        bool          `json:"stream,omitempty"`
	StreamOptions StreamOptions `json:"stream_options,omitempty"`

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
		Model:  utils.GetByJSONPath[string](parsed, "{ .model }"),
		Stream: utils.GetByJSONPath[bool](parsed, "{ .stream }"),
		StreamOptions: StreamOptions{
			IncludeUsage: utils.GetByJSONPath[bool](parsed, "{ .stream_options.include_usage }"),
		},
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
	patch, err := jsonpatch.DecodePatch(NewPatches(
		NewReplace("/model", model),
	))
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
