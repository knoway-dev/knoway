package openai

import (
	"bytes"
	"fmt"
	"net/http"

	jsonpatch "github.com/evanphx/json-patch/v5"
	structpb "github.com/golang/protobuf/ptypes/struct"

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

func (r *CompletionsRequest) GetRequestType() object.RequestType {
	return object.RequestTypeCompletions
}

type ChatCompletionsRequest struct {
	Model         string        `json:"model,omitempty"`
	Stream        bool          `json:"stream,omitempty"`
	StreamOptions StreamOptions `json:"stream_options,omitempty"`

	bodyParsed      map[string]any
	bodyBuffer      *bytes.Buffer
	incomingRequest *http.Request
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

func (r *ChatCompletionsRequest) MarshalJSON() ([]byte, error) {
	return r.bodyBuffer.Bytes(), nil
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

func (r *ChatCompletionsRequest) SetDefaultParams(params map[string]*structpb.Value) error {
	for k, v := range params {
		if _, exists := r.bodyParsed[k]; exists {
			continue
		}

		var err error

		r.bodyBuffer, r.bodyParsed, err = modifyBufferBodyAndParsed(r.bodyBuffer, nil, NewAdd("/"+k, &v))
		if err != nil {
			return fmt.Errorf("failed to add key %s: %w", k, err)
		}
	}

	changedModel := r.bodyParsed["model"]
	if model, ok := changedModel.(string); ok && r.Model != model {
		r.Model = model
	}

	return nil
}

func (r *ChatCompletionsRequest) SetOverrideParams(params map[string]*structpb.Value) error {
	applyOpt := jsonpatch.NewApplyOptions()
	applyOpt.EnsurePathExistsOnAdd = true

	for k, v := range params {
		var err error

		r.bodyBuffer, r.bodyParsed, err = modifyBufferBodyAndParsed(r.bodyBuffer, applyOpt, NewAdd("/"+k, &v))
		if err != nil {
			return err
		}
	}

	changedModel := r.bodyParsed["model"]
	if model, ok := changedModel.(string); ok && r.Model != model {
		r.Model = model
	}

	return nil
}

func (r *ChatCompletionsRequest) GetRequestType() object.RequestType {
	return object.RequestTypeChatCompletions
}

func (r *ChatCompletionsRequest) GetRequest() *http.Request {
	return r.incomingRequest
}
