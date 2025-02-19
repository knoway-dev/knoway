package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"knoway.dev/pkg/object"
	"knoway.dev/pkg/types/sse"
	"knoway.dev/pkg/utils"
)

var _ object.LLMChunkResponse = (*ChatCompletionStreamChunk)(nil)

type ChatCompletionStreamChunk struct {
	Model string                `json:"model"`
	Usage *ChatCompletionsUsage `json:"usage,omitempty"`

	response     object.LLMStreamResponse
	responseBody json.RawMessage
	bodyParsed   map[string]any

	isEmpty bool
	isDone  bool
	isUsage bool
	isFirst bool
}

func NewChatCompletionStreamChunk(streamResp *ChatCompletionStreamResponse, bs []byte) (*ChatCompletionStreamChunk, error) {
	resp := new(ChatCompletionStreamChunk)

	err := resp.processBytes(bs)
	if err != nil {
		return NewEmptyChatCompletionStreamChunk(streamResp), err
	}

	model := utils.GetByJSONPath[string](resp.bodyParsed, "{ .model }")

	resp.response = streamResp
	resp.Model = model
	resp.isFirst = streamResp.IsFirst()

	if streamResp.GetModel() == "" {
		err = streamResp.SetModel(model)
		if err != nil {
			return NewEmptyChatCompletionStreamChunk(streamResp), err
		}
	}

	return resp, nil
}

func NewEmptyChatCompletionStreamChunk(streamResp *ChatCompletionStreamResponse) *ChatCompletionStreamChunk {
	resp := new(ChatCompletionStreamChunk)

	resp.isEmpty = true
	resp.response = streamResp
	resp.isFirst = streamResp.IsFirst()

	return resp
}

func NewUsageChatCompletionStreamChunk(streamResp *ChatCompletionStreamResponse, bs []byte) (*ChatCompletionStreamChunk, error) {
	resp := new(ChatCompletionStreamChunk)

	err := resp.processBytes(bs)
	if err != nil {
		return NewEmptyChatCompletionStreamChunk(streamResp), err
	}

	usageMap := utils.GetByJSONPath[map[string]any](resp.bodyParsed, "{ .usage }")
	model := utils.GetByJSONPath[string](resp.bodyParsed, "{ .model }")

	resp.Usage, err = utils.FromMap[ChatCompletionsUsage](usageMap)
	if err != nil {
		return NewEmptyChatCompletionStreamChunk(streamResp), err
	}

	resp.isUsage = true
	resp.response = streamResp
	resp.Model = model
	resp.isFirst = streamResp.IsFirst()

	if streamResp.GetModel() == "" {
		err = streamResp.SetModel(model)
		if err != nil {
			return NewEmptyChatCompletionStreamChunk(streamResp), err
		}
	}

	return resp, nil
}

func NewDoneChatCompletionStreamChunk(streamResp *ChatCompletionStreamResponse) *ChatCompletionStreamChunk {
	resp := new(ChatCompletionStreamChunk)

	resp.isDone = true
	resp.response = streamResp

	return resp
}

func (r *ChatCompletionStreamChunk) IsFirst() bool {
	return r.isFirst
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

func (r *ChatCompletionStreamChunk) GetUsage() object.LLMUsage {
	return r.Usage
}

// https://github.com/sashabaranov/go-openai/blob/74ed75f291f8f55d1104a541090d46c021169115/stream_reader.go#L13C1-L16C2
var (
	headerData            = []byte("data: ")
	errorPrefix           = []byte(`data: {"error":`)
	usageCompletionTokens = []byte(`"completion_tokens":`)
)

var _ object.LLMStreamResponse = (*ChatCompletionStreamResponse)(nil)

type ChatCompletionStreamResponse struct {
	Model string                `json:"model"`
	Usage *ChatCompletionsUsage `json:"usage,omitempty"`
	Error *ErrorResponse        `json:"error,omitempty"`

	reader           *bufio.Reader
	request          object.LLMRequest
	outgoingResponse *http.Response

	// chunk processing specific
	hasErrorPrefix   bool
	errorEventBuffer *bytes.Buffer
	isDone           bool
	chunkNum         int

	// Mutex for locking
	mu sync.Mutex
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
	// NOTICE: stream response should not be marshaled
	return json.Marshal(nil)
}

func (r *ChatCompletionStreamResponse) IsFirst() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.chunkNum == 1
}

func (r *ChatCompletionStreamResponse) IsEOF() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

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
		r.mu.Lock()
		r.hasErrorPrefix = true
		r.mu.Unlock()
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

	r.mu.Lock()
	r.chunkNum++
	r.mu.Unlock()

	noPrefixLine := bytes.TrimPrefix(noSpaceLine, headerData)
	if string(noPrefixLine) == "[DONE]" {
		r.mu.Lock()
		r.isDone = true
		r.mu.Unlock()

		return NewDoneChatCompletionStreamChunk(r), io.EOF
	}

	if bytes.Contains(noPrefixLine, usageCompletionTokens) {
		chunk, err := NewUsageChatCompletionStreamChunk(r, noPrefixLine)
		if err != nil {
			return chunk, err
		}

		r.mu.Lock()
		r.Usage = chunk.Usage
		r.mu.Unlock()

		return chunk, nil
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

func (r *ChatCompletionStreamResponse) SetModel(model string) error {
	r.Model = model

	return nil
}

func (r *ChatCompletionStreamResponse) GetUsage() object.LLMUsage {
	return r.Usage
}

func (r *ChatCompletionStreamResponse) GetError() object.LLMError {
	if r.Error != nil {
		return r.Error
	}

	return nil
}
