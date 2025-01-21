package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"knoway.dev/pkg/object"
)

var _ object.LLMResponse = (*ImageGenerationsResponse)(nil)

type ImageGenerationsResponse struct {
	Status int            `json:"status"`
	Model  string         `json:"model"`
	Usage  *Usage         `json:"usage,omitempty"`
	Error  *ErrorResponse `json:"error,omitempty"`

	request          object.LLMRequest
	responseBody     json.RawMessage
	bodyParsed       map[string]any
	outgoingResponse *http.Response
}

func NewImageGenerationsResponse(request object.LLMRequest, response *http.Response, reader *bufio.Reader) (*ImageGenerationsResponse, error) {
	resp := new(ImageGenerationsResponse)

	buffer := new(bytes.Buffer)

	_, err := buffer.ReadFrom(reader)
	if err != nil {
		return nil, err
	}

	err = resp.processBytes(buffer.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	resp.request = request
	resp.outgoingResponse = response

	return resp, nil
}

func (r *ImageGenerationsResponse) processBytes(bs []byte, response *http.Response) error {
	if r == nil {
		return nil
	}

	r.responseBody = bs
	r.Status = response.StatusCode

	var body map[string]any

	err := json.Unmarshal(bs, &body)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	r.bodyParsed = body

	errorResponse, err := unmarshalErrorResponseFromParsedBody(body, response, bs)
	if err != nil {
		return err
	}
	if errorResponse != nil {
		r.Error = errorResponse
	}

	return nil
}

func (r *ImageGenerationsResponse) MarshalJSON() ([]byte, error) {
	return r.responseBody, nil
}

func (r *ImageGenerationsResponse) GetRequestID() string {
	// TODO: implement
	return ""
}

func (r *ImageGenerationsResponse) IsStream() bool {
	return false
}

func (r *ImageGenerationsResponse) GetModel() string {
	return r.Model
}

func (r *ImageGenerationsResponse) SetModel(model string) error {
	r.Model = model

	return nil
}

func (r *ImageGenerationsResponse) GetUsage() object.LLMUsage {
	return r.Usage
}

func (r *ImageGenerationsResponse) GetError() object.LLMError {
	if r.Error != nil {
		return r.Error
	}

	return nil
}
