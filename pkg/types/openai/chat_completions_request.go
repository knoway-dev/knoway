package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type CompletionsRequest struct {
	// REVIEW: do we need a dedicated struct for completions?
	// TODO: consider to drop the support of completions or have dedicated struct for completions
	// or otherwise the prompt filtering and content moderation will be a bit tricky to implement
	*ChatCompletionsRequest
}

func NewCompletionsRequest(httpRequest *http.Request) (*CompletionsRequest, error) {
	req, err := NewChatCompletionRequest(httpRequest)
	if err != nil {
		return nil, err
	}

	return &CompletionsRequest{
		ChatCompletionsRequest: req,
	}, nil
}

type ChatCompletionsRequest struct {
	Model         string        `json:"model,omitempty"`
	Stream        bool          `json:"stream,omitempty"`
	StreamOptions StreamOptions `json:"stream_options,omitempty"`

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

		req.bodyBuffer, req.bodyParsed, err = modifyBufferBodyAndParsed(req.bodyBuffer, nil, NewAdd("/stream_options", StreamOptions{IncludeUsage: true}))
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

	r.bodyBuffer, r.bodyParsed, err = modifyBufferBodyAndParsed(r.bodyBuffer, nil, NewReplace("/model", model))
	if err != nil {
		return err
	}

	r.Model = model

	return nil
}

func parseValue(value string) interface{} {
	var res interface{}
	err := json.Unmarshal([]byte(value), &res)
	if err != nil {
		return value
	}

	switch v := res.(type) {
	case float64:
		if float64(int(v)) == v {
			return int(v)
		}
		return v
	case int, int32, int64, bool, string:
		return v
	default:
		return res
	}
}

func (r *ChatCompletionsRequest) SetDefaultParams(params map[string]string) error {
	for key, value := range params {
		if _, exists := r.bodyParsed[key]; exists {
			continue
		}

		parsedValue := parseValue(value)
		var err error
		r.bodyBuffer, r.bodyParsed, err = modifyBufferBodyAndParsed(r.bodyBuffer, nil, NewAdd(fmt.Sprintf("/%s", key), &parsedValue))
		if err != nil {
			return fmt.Errorf("failed to add key %s: %w", key, err)
		}
	}

	return nil
}

func (r *ChatCompletionsRequest) SetOverrideParams(params map[string]string) error {
	applyOpt := jsonpatch.NewApplyOptions()
	applyOpt.EnsurePathExistsOnAdd = true

	for k, v := range params {
		parsedValue := parseValue(v)
		var err error
		r.bodyBuffer, r.bodyParsed, err = modifyBufferBodyAndParsed(r.bodyBuffer, applyOpt, NewAdd(fmt.Sprintf("/%s", k), &parsedValue))
		if err != nil {
			return err
		}
	}

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
