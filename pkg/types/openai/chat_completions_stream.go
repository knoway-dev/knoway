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

var _ object.LLMChunkResponse = (*ChatCompletionStreamChunk)(nil)

type ChatCompletionStreamChunk struct {
	Model string `json:"model"`
	Usage *Usage `json:"usage,omitempty"`

	response     object.LLMStreamResponse
	responseBody json.RawMessage
	bodyParsed   map[string]any

	isEmpty bool
	isDone  bool
	isUsage bool
}

func NewChatCompletionStreamChunk(streamResp object.LLMStreamResponse, bs []byte, model string) (*ChatCompletionStreamChunk, error) {
	resp := new(ChatCompletionStreamChunk)

	err := resp.processBytes(bs)
	if err != nil {
		return NewEmptyChatCompletionStreamChunk(streamResp), err
	}

	resp.response = streamResp

	err = resp.SetModel(model)
	if err != nil {
		return NewEmptyChatCompletionStreamChunk(streamResp), err
	}

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

func NewUsageChatCompletionStreamChunk(streamResp object.LLMStreamResponse, bs []byte, model string) (*ChatCompletionStreamChunk, error) {
	resp := new(ChatCompletionStreamChunk)

	err := resp.processBytes(bs)
	if err != nil {
		return NewEmptyChatCompletionStreamChunk(streamResp), err
	}

	usageMap := utils.GetByJSONPath[map[string]any](resp.bodyParsed, "{ .usage }")

	resp.Usage, err = utils.FromMap[Usage](usageMap)
	if err != nil {
		return NewEmptyChatCompletionStreamChunk(streamResp), err
	}

	resp.isUsage = true
	resp.response = streamResp

	err = resp.SetModel(model)
	if err != nil {
		return NewEmptyChatCompletionStreamChunk(streamResp), err
	}

	return resp, nil
}

func (r *ChatCompletionStreamChunk) IsEmpty() bool {
	return r.isEmpty
}

func (r *ChatCompletionStreamChunk) IsDone() bool {
	return r.isDone
}

func (r *ChatCompletionStreamChunk) IsUsage() bool {
	return r.isUsage
}

func (r *ChatCompletionStreamChunk) GetModel() string {
	return r.Model
}

func (r *ChatCompletionStreamChunk) SetModel(model string) error {
	var err error

	r.responseBody, r.bodyParsed, err = modifyBytesBodyAndParsed(r.responseBody, NewReplace("/model", model))
	if err != nil {
		return err
	}

	r.Model = model

	return nil
}

func (r *ChatCompletionStreamChunk) processBytes(bs []byte) error {
	r.responseBody = bs

	var body map[string]any

	err := json.Unmarshal(bs, &body)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	r.bodyParsed = body

	return nil
}

func (r *ChatCompletionStreamChunk) GetResponse() object.LLMStreamResponse {
	return r.response
}

func (r *ChatCompletionStreamChunk) MarshalJSON() ([]byte, error) {
	if r.isDone {
		return []byte("[DONE]"), nil
	}

	return json.Marshal(r.bodyParsed)
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
	headerData            = []byte("data: ")
	errorPrefix           = []byte(`data: {"error":`)
	usageCompletionTokens = []byte(`"completion_tokens":`)
)

var _ object.LLMStreamResponse = (*ChatCompletionStreamResponse)(nil)

type ChatCompletionStreamResponse struct {
	Model string         `json:"model"`
	Usage *Usage         `json:"usage,omitempty"`
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
	if bytes.Contains(noPrefixLine, usageCompletionTokens) {
		chunk, err := NewUsageChatCompletionStreamChunk(r, noPrefixLine, r.GetModel())
		if err != nil {
			return chunk, err
		}

		r.Usage = chunk.Usage
	}

	return NewChatCompletionStreamChunk(r, noPrefixLine, r.GetModel())
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

func (r *ChatCompletionStreamResponse) SetModel(model string) error {
	r.Model = model

	return nil
}

func (r *ChatCompletionStreamResponse) GetUsage() object.LLMUsage {
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
