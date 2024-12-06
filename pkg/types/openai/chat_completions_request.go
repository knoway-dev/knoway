package openai

import (
	"bytes"
	"io"
	"net/http"

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

	object.BaseLLMRequest
}

func NewChatCompletionRequest(httpRequest *http.Request) (*ChatCompletionsRequest, error) {
	buffer, parsed, err := utils.ReadAsJSONWithClose(httpRequest.Body)
	if err != nil {
		return nil, NewErrorInvalidBody()
	}

	req := &ChatCompletionsRequest{
		Model:  utils.GetByJSONPath[string](parsed, "{ .model }"),
		Stream: utils.GetByJSONPath[bool](parsed, "{ .stream }"),
		StreamOptions: StreamOptions{
			IncludeUsage: utils.GetByJSONPath[bool](parsed, "{ .stream_options.include_usage }"),
		},
		bodyParsed:      parsed,
		bodyBuffer:      buffer,
		incomingRequest: httpRequest,
	}

	if req.Stream && !req.StreamOptions.IncludeUsage {
		var err error

		req.bodyBuffer, req.bodyParsed, err = modifyBufferBodyAndParsed(req.bodyBuffer, NewAdd("/stream_options", StreamOptions{IncludeUsage: true}))
		if err != nil {
			return nil, err
		}

		req.StreamOptions.IncludeUsage = true
	}

	return req, nil
}

func (r *ChatCompletionsRequest) IsStream() bool {
	return r.Stream
}

func (r *ChatCompletionsRequest) GetModel() string {
	return r.Model
}

func (r *ChatCompletionsRequest) SetModel(model string) error {
	var err error

	r.bodyBuffer, r.bodyParsed, err = modifyBufferBodyAndParsed(r.bodyBuffer, NewReplace("/model", model))
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
