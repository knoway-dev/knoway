package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"net/http"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"

	"knoway.dev/pkg/object"
	"knoway.dev/pkg/utils"
)

var _ object.LLMResponse = (*ImageGenerationsResponse)(nil)

type ImageGenerationsImage struct {
	ImageConfig image.Config `json:"-"`
	ImageFormat string       `json:"-"`

	Base64JSON    string `json:"b64_json"`
	URL           string `json:"url"`
	RevisedPrompt string `json:"revised_prompt"`
}

func (i *ImageGenerationsImage) resolveImage(ctx context.Context, client *http.Client) error {
	switch {
	case i.Base64JSON != "":
		decodedBase64Payload, err := base64.StdEncoding.DecodeString(i.Base64JSON)
		if err != nil {
			return err
		}

		decodedImage, format, err := image.DecodeConfig(bytes.NewReader(decodedBase64Payload))
		if err != nil {
			return err
		}

		i.ImageConfig = decodedImage
		i.ImageFormat = format

		return nil
	case i.URL != "":
		httpClient := client
		if httpClient == nil {
			httpClient = http.DefaultClient
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, i.URL, nil)
		if err != nil {
			return err
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		content, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		decodedImage, format, err := image.DecodeConfig(bytes.NewReader(content))
		if err != nil {
			return err
		}

		i.ImageConfig = decodedImage
		i.ImageFormat = format

		return nil
	default:
		return nil
	}
}

type newImageGenerationsResponseOptions struct {
	httpClient *http.Client
}

type NewImageGenerationsResponseOption func(*newImageGenerationsResponseOptions)

func NewImageGenerationsResponseWithHTTPClient(client *http.Client) NewImageGenerationsResponseOption {
	return func(o *newImageGenerationsResponseOptions) {
		o.httpClient = client
	}
}

type ImageGenerationsResponse struct {
	Status int                      `json:"status"`
	Model  string                   `json:"model"`
	Usage  *ImageGenerationsUsage   `json:"usage,omitempty"`
	Error  *ErrorResponse           `json:"error,omitempty"`
	Images []*ImageGenerationsImage `json:"images"`

	request          object.LLMRequest
	responseBody     json.RawMessage
	bodyParsed       map[string]any
	outgoingResponse *http.Response
	options          *newImageGenerationsResponseOptions
}

func NewImageGenerationsResponse(ctx context.Context, request object.LLMRequest, response *http.Response, reader *bufio.Reader, opts ...NewImageGenerationsResponseOption) (*ImageGenerationsResponse, error) {
	options := &newImageGenerationsResponseOptions{
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(options)
	}

	resp := new(ImageGenerationsResponse)
	resp.options = options
	resp.Usage = new(ImageGenerationsUsage)

	buffer := new(bytes.Buffer)

	_, err := buffer.ReadFrom(reader)
	if err != nil {
		return nil, err
	}

	err = resp.processBytes(ctx, buffer.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	resp.request = request
	resp.outgoingResponse = response

	return resp, nil
}

func (r *ImageGenerationsResponse) processBytes(ctx context.Context, bs []byte, response *http.Response) error {
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
	dataArray := utils.GetByJSONPath[[]map[string]any](body, "{ .data }")
	r.Images = make([]*ImageGenerationsImage, 0, len(dataArray))

	if len(dataArray) > 0 {
		for _, data := range dataArray {
			base64JSON := utils.GetByJSONPath[string](data, "{ .b64_json }")
			url := utils.GetByJSONPath[string](data, "{ .url }")
			revisedPrompt := utils.GetByJSONPath[string](data, "{ .revised_prompt }")

			if base64JSON == "" && url == "" {
				continue
			}

			imageObject := &ImageGenerationsImage{
				Base64JSON:    base64JSON,
				URL:           url,
				RevisedPrompt: revisedPrompt,
			}

			err = imageObject.resolveImage(ctx, r.options.httpClient)
			if err != nil {
				return err
			}

			r.Images = append(r.Images, imageObject)
			r.Usage.Images = append(r.Usage.Images, ImageGenerationsUsageImage{
				Width:  uint64(imageObject.ImageConfig.Width),
				Height: uint64(imageObject.ImageConfig.Height),
			})
		}
	}

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
