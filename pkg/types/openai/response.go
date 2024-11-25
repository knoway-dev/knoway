package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"knoway.dev/pkg/object"
	"knoway.dev/pkg/types/sse"
	"knoway.dev/pkg/utils"
)

var _ object.LLMResponse = (*ChatCompletionResponse)(nil)

type ChatCompletionResponse struct {
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

func NewChatCompletionResponse(request object.LLMRequest, response *http.Response, reader *bufio.Reader) (*ChatCompletionResponse, error) {
	resp := new(ChatCompletionResponse)

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

func (r *ChatCompletionResponse) processBytes(bs []byte) error {
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

var _ object.LLMChunkResponse = (*ChatCompletionStreamChunk)(nil)

type ChatCompletionStreamChunk struct {
	response         object.LLMStreamResponse
	responseBody     json.RawMessage
	unmarshalledBody map[string]any

	isEmpty bool
	isDone  bool
}

func NewChatCompletionStreamChunk(streamResp object.LLMStreamResponse, bs []byte) (*ChatCompletionStreamChunk, error) {
	resp := new(ChatCompletionStreamChunk)

	err := resp.processBytes(bs)
	if err != nil {
		return NewEmptyChatCompletionStreamChunk(streamResp), err
	}

	resp.response = streamResp

	return resp, nil
}

func NewEmptyChatCompletionStreamChunk(streamResp object.LLMStreamResponse) *ChatCompletionStreamChunk {
	resp := new(ChatCompletionStreamChunk)

	resp.isEmpty = true

	resp.response = streamResp

	return resp
}

func NewDoneChatCompletionStreamChunk(streamResp object.LLMStreamResponse) *ChatCompletionStreamChunk {
	resp := new(ChatCompletionStreamChunk)

	resp.isDone = true

	resp.response = streamResp

	return resp
}

func (r *ChatCompletionStreamChunk) IsEmpty() bool {
	return r.isEmpty
}

func (r *ChatCompletionStreamChunk) IsDone() bool {
	return r.isDone
}

func (r *ChatCompletionStreamChunk) processBytes(bs []byte) error {
	r.responseBody = bs

	var body map[string]any

	err := json.Unmarshal(bs, &body)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	r.unmarshalledBody = body

	return nil
}

func (r *ChatCompletionStreamChunk) GetResponse() object.LLMStreamResponse {
	return r.response
}

func (r *ChatCompletionStreamChunk) MarshalJSON() ([]byte, error) {
	if r.isDone {
		return []byte("[DONE]"), nil
	}

	return json.Marshal(r.unmarshalledBody)
}

func (r *ChatCompletionStreamChunk) ToServerSentEvent() (*sse.Event, error) {
	data, err := r.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return &sse.Event{
		Data: data,
	}, nil
}

// https://github.com/sashabaranov/go-openai/blob/74ed75f291f8f55d1104a541090d46c021169115/stream_reader.go#L13C1-L16C2
var (
	headerData  = []byte("data: ")
	errorPrefix = []byte(`data: {"error":`)
)

var _ object.LLMStreamResponse = (*ChatCompletionStreamResponse)(nil)

type ChatCompletionStreamResponse struct {
	Model string         `json:"model"`
	Usage *object.Usage  `json:"usage,omitempty"`
	Error *ErrorResponse `json:"error,omitempty"`

	// TODO: add more fields

	reader           *bufio.Reader
	request          object.LLMRequest
	outgoingResponse *http.Response

	// chunk processing specific
	hasErrorPrefix   bool
	errorEventBuffer *bytes.Buffer
	isDone           bool
}

func NewChatCompletionStreamResponse(request object.LLMRequest, response *http.Response, reader *bufio.Reader) (*ChatCompletionStreamResponse, error) {
	resp := new(ChatCompletionStreamResponse)

	resp.reader = reader
	resp.request = request
	resp.outgoingResponse = response
	resp.errorEventBuffer = new(bytes.Buffer)

	return resp, nil
}

func (r *ChatCompletionStreamResponse) MarshalJSON() ([]byte, error) {
	// TODO: implement
	return nil, nil
}

func (r *ChatCompletionStreamResponse) IsEOF() bool {
	return r.isDone
}

func (r *ChatCompletionStreamResponse) NextChunk() (object.LLMChunkResponse, error) {
	line, err := r.reader.ReadBytes('\n')
	if err != nil || r.hasErrorPrefix {
		// TODO: handle error
		return NewEmptyChatCompletionStreamChunk(r), err
	}

	noSpaceLine := bytes.TrimSpace(line)
	if bytes.HasPrefix(noSpaceLine, errorPrefix) {
		r.hasErrorPrefix = true
	}

	if !bytes.HasPrefix(noSpaceLine, headerData) || r.hasErrorPrefix {
		if r.hasErrorPrefix {
			noSpaceLine = bytes.TrimPrefix(noSpaceLine, headerData)
		}

		_, writeErr := r.errorEventBuffer.Write(noSpaceLine)
		if writeErr != nil {
			return NewEmptyChatCompletionStreamChunk(r), writeErr
		}

		// TODO: Empty message handling
		return NewEmptyChatCompletionStreamChunk(r), nil
	}

	noPrefixLine := bytes.TrimPrefix(noSpaceLine, headerData)
	if string(noPrefixLine) == "[DONE]" {
		r.isDone = true
		return NewDoneChatCompletionStreamChunk(r), io.EOF
	}

	return NewChatCompletionStreamChunk(r, noPrefixLine)
}

func (r *ChatCompletionStreamResponse) IsStream() bool {
	return true
}

func (r *ChatCompletionStreamResponse) GetRequestID() string {
	// TODO: implement
	return ""
}

func (r *ChatCompletionStreamResponse) GetModel() string {
	return r.Model
}

func (r *ChatCompletionStreamResponse) GetUsage() *object.Usage {
	return r.Usage
}

func (r *ChatCompletionStreamResponse) GetOutgoingResponse() *http.Response {
	return r.outgoingResponse
}

func (r *ChatCompletionStreamResponse) GetError() error {
	if r.Error != nil {
		return r.Error
	}

	return nil
}