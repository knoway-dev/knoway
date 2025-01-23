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

var _ object.LLMRequest = (*ImageGenerationsRequest)(nil)

type ImageGenerationsRequest struct {
	Model string `json:"model,omitempty"`

	bodyParsed      map[string]any
	bodyBuffer      *bytes.Buffer
	incomingRequest *http.Request
}

func NewImageGenerationsRequest(httpRequest *http.Request) (*ImageGenerationsRequest, error) {
	buffer, parsed, err := utils.ReadAsJSONWithClose(httpRequest.Body)
	if err != nil {
		return nil, NewErrorInvalidBody()
	}

	req := &ImageGenerationsRequest{
		Model:           utils.GetByJSONPath[string](parsed, "{ .model }"),
		bodyParsed:      parsed,
		bodyBuffer:      buffer,
		incomingRequest: httpRequest,
	}

	return req, nil
}

func (r *ImageGenerationsRequest) MarshalJSON() ([]byte, error) {
	return r.bodyBuffer.Bytes(), nil
}

func (r *ImageGenerationsRequest) IsStream() bool {
	return false
}

func (r *ImageGenerationsRequest) GetModel() string {
	return r.Model
}

func (r *ImageGenerationsRequest) SetModel(model string) error {
	var err error

	r.bodyBuffer, r.bodyParsed, err = modifyBufferBodyAndParsed(r.bodyBuffer, nil, NewReplace("/model", model))
	if err != nil {
		return err
	}

	r.Model = model

	return nil
}

func (r *ImageGenerationsRequest) SetDefaultParams(params map[string]*structpb.Value) error {
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

func (r *ImageGenerationsRequest) SetOverrideParams(params map[string]*structpb.Value) error {
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

func (r *ImageGenerationsRequest) GetRequestType() object.RequestType {
	return object.RequestTypeImageGeneration
}
