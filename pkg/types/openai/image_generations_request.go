package openai

import (
	"bytes"
	"fmt"
	"net/http"

	jsonpatch "github.com/evanphx/json-patch/v5"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/samber/lo"

	"knoway.dev/pkg/object"
	"knoway.dev/pkg/utils"
)

const (
	// API Reference - OpenAI API
	// https://platform.openai.com/docs/api-reference/images/create
	// size, string or null, Optional, Defaults to 1024x1024
	defaultImageGenerationRequestSizeWidth  = 1024
	defaultImageGenerationRequestSizeHeight = 1024
)

type ImageGenerationsRequestSize struct {
	Width  uint64 `json:"width"`
	Height uint64 `json:"height"`
}

var _ object.LLMRequest = (*ImageGenerationsRequest)(nil)

type ImageGenerationsRequest struct {
	Model   string                       `json:"model,omitempty"`
	N       *uint64                      `json:"n,omitempty"`
	Quality *string                      `json:"quality,omitempty"`
	Style   *string                      `json:"style,omitempty"`
	Size    *ImageGenerationsRequestSize `json:"size,omitempty"`

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
		N:               utils.GetByJSONPath[*uint64](parsed, "{ .n }"),
		Quality:         utils.GetByJSONPath[*string](parsed, "{ .quality }"),
		Style:           utils.GetByJSONPath[*string](parsed, "{ .style }"),
		bodyParsed:      parsed,
		bodyBuffer:      buffer,
		incomingRequest: httpRequest,
	}

	size := utils.GetByJSONPath[*string](parsed, "{ .size }")
	if size == nil {
		req.Size = &ImageGenerationsRequestSize{
			Width:  defaultImageGenerationRequestSizeWidth,
			Height: defaultImageGenerationRequestSizeHeight,
		}
	} else {
		req.Size, err = parseImageGenerationsSizeString(size)
		if err != nil {
			return nil, err
		}
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

		if k == "size" {
			r.Size, err = parseImageGenerationsSizeString(lo.ToPtr(v.GetStringValue()))
			if err != nil {
				return err
			}
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

		if k == "size" {
			r.Size, err = parseImageGenerationsSizeString(lo.ToPtr(v.GetStringValue()))
			if err != nil {
				return err
			}
		}
	}

	changedModel := r.bodyParsed["model"]
	if model, ok := changedModel.(string); ok && r.Model != model {
		r.Model = model
	}

	return nil
}

func (r *ImageGenerationsRequest) RemoveParamKeys(keys []string) error {
	applyOpt := jsonpatch.NewApplyOptions()
	applyOpt.AllowMissingPathOnRemove = true

	for _, v := range keys {
		var err error

		r.bodyBuffer, r.bodyParsed, err = modifyBufferBodyAndParsed(r.bodyBuffer, applyOpt, NewRemove("/"+v))
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ImageGenerationsRequest) GetRequestType() object.RequestType {
	return object.RequestTypeImageGenerations
}
