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

var _ object.LLMRequest = (*ChatCompletionsRequest)(nil)

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type ChatCompletionsRequest struct {
	Model         string        `json:"model,omitempty"`
	Stream        bool          `json:"stream,omitempty"`
	StreamOptions StreamOptions `json:"stream_options,omitempty"`

	// TODO: add more fields

	bodyParsed      map[string]any
	bodyBuffer      *bytes.Buffer
	incomingRequest *http.Request
}

func NewChatCompletionRequest(httpRequest *http.Request) (*ChatCompletionsRequest, error) {
	buffer, parsed, err := utils.ReadAsJSONWithClose(httpRequest.Body)
	if err != nil {
		return nil, NewErrorInvalidBody()
	}

	return &ChatCompletionsRequest{
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

func (r *ChatCompletionsRequest) IsStream() bool {
	return r.Stream
}

func (r *ChatCompletionsRequest) GetModel() string {
	return r.Model
}

func (r *ChatCompletionsRequest) SetModel(model string) error {
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

func (r *ChatCompletionsRequest) GetBody() io.Reader {
	return bytes.NewBuffer(r.bodyBuffer.Bytes())
}

func (r *ChatCompletionsRequest) GetBodyBuffer() *bytes.Buffer {
	return r.bodyBuffer
}

func (r *ChatCompletionsRequest) GetBodyParsed() map[string]any {
	return r.bodyParsed
}

func (r *ChatCompletionsRequest) GetIncomingRequest() *http.Request {
	return r.incomingRequest
}
