package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/protobuf/types/known/anypb"
	"knoway.dev/api/filters/v1alpha1"
	filters2 "knoway.dev/pkg/clusters/filters"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/protoutils"
	"knoway.dev/pkg/types/openai"
	"knoway.dev/pkg/utils"
)

func NewResponseUnmarshallerWithConfig(cfg *anypb.Any) (filters2.ClusterFilter, error) {
	c, err := protoutils.FromAny[*v1alpha1.OpenAIResponseUnmarshallerConfig](cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}

	return &responseUnmarshaller{
		cfg: c,
	}, nil
}

var _ filters2.ClusterFilterResponseUnmarshaller = (*responseUnmarshaller)(nil)

type responseUnmarshaller struct {
	cfg *v1alpha1.OpenAIResponseUnmarshallerConfig
	filters2.ClusterFilter
}

var _ object.LLMResponse = (*OpenAIChatCompletionResponse)(nil)

type OpenAIChatCompletionResponse struct {
	Model string                `json:"model"`
	Usage *object.Usage         `json:"usage,omitempty"`
	Error *openai.ErrorResponse `json:"error,omitempty"`

	responseBody     json.RawMessage
	unmarshalledBody map[string]any
	outgoingResponse *http.Response
}

func (r *OpenAIChatCompletionResponse) UnmarshalJSON(bs []byte) error {
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
		respErr, err := utils.FromMap[openai.Error](respErrMap)
		if err != nil {
			return fmt.Errorf("failed to unmarshal error: %w", err)
		}

		r.Error = &openai.ErrorResponse{
			FromUpstream: true,
			ErrorBody:    respErr,
		}
	}

	return nil
}

func (r *OpenAIChatCompletionResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.responseBody)
}

func (r *OpenAIChatCompletionResponse) IsStream() bool {
	return false
}

func (r *OpenAIChatCompletionResponse) GetRequestID() string {
	return ""
}

func (r *OpenAIChatCompletionResponse) GetModel() string {
	return r.Model
}

func (r *OpenAIChatCompletionResponse) GetUsage() *object.Usage {
	return nil
}

func (r *OpenAIChatCompletionResponse) GetOutgoingResponse() *http.Response {
	return r.outgoingResponse
}

func (r *OpenAIChatCompletionResponse) GetError() error {
	if r.Error != nil {
		return r.Error
	}

	return nil
}

func (f *responseUnmarshaller) UnmarshalResponseBody(ctx context.Context, req object.LLMRequest, rawResponse *http.Response, buffer *bytes.Buffer, pre object.LLMResponse) (object.LLMResponse, error) {
	contentType := rawResponse.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		return nil, fmt.Errorf("unexpected content type %s on status %d", contentType, rawResponse.StatusCode)
	}

	resp := new(OpenAIChatCompletionResponse)

	err := resp.UnmarshalJSON(buffer.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.Error != nil {
		resp.Error.Status = rawResponse.StatusCode
	}

	resp.outgoingResponse = rawResponse

	return resp, nil
}
