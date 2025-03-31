package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"net/http"
	"sync"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/samber/lo"
	_ "golang.org/x/image/webp"

	"knoway.dev/api/clusters/v1alpha1"
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

func NewImageGenerationsImage(data map[string]any) *ImageGenerationsImage {
	base64JSON := utils.GetByJSONPath[string](data, "{ .b64_json }")
	url := utils.GetByJSONPath[string](data, "{ .url }")
	revisedPrompt := utils.GetByJSONPath[string](data, "{ .revised_prompt }")

	if base64JSON == "" && url == "" {
		return nil
	}

	return &ImageGenerationsImage{
		Base64JSON:    base64JSON,
		URL:           url,
		RevisedPrompt: revisedPrompt,
	}
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
	httpClient     *http.Client
	meteringPolicy *v1alpha1.ClusterMeteringPolicy
}

type NewImageGenerationsResponseOption func(*newImageGenerationsResponseOptions)

func NewImageGenerationsResponseWithHTTPClient(client *http.Client) NewImageGenerationsResponseOption {
	return func(o *newImageGenerationsResponseOptions) {
		o.httpClient = client
	}
}

func NewImageGenerationsResponseWithUsage(usage *v1alpha1.ClusterMeteringPolicy) NewImageGenerationsResponseOption {
	return func(o *newImageGenerationsResponseOptions) {
		o.meteringPolicy = usage
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

	err = resp.processBytes(ctx, buffer.Bytes(), request, response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	resp.request = request
	resp.outgoingResponse = response

	return resp, nil
}

func (r *ImageGenerationsResponse) processBytes(ctx context.Context, bs []byte, request object.LLMRequest, response *http.Response) error {
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
			imageObject := NewImageGenerationsImage(data)
			if imageObject == nil {
				continue
			}

			r.Images = append(r.Images, imageObject)
		}
	}

	err = r.resolveUsage(ctx, request)
	if err != nil {
		return err
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

func (r *ImageGenerationsResponse) resolveUsage(ctx context.Context, request object.LLMRequest) error {
	imageGenerationRequest, ok := request.(*ImageGenerationsRequest)
	if !ok {
		return errors.New("failed to cast request to ImageGenerationsRequest")
	}

	r.Usage.Images = lo.Map(r.Images, func(imageObject *ImageGenerationsImage, _ int) *ImageGenerationsUsageImage {
		var width uint64
		var height uint64

		if imageGenerationRequest.Size != nil {
			width = imageGenerationRequest.Size.Width
			height = imageGenerationRequest.Size.Height
		}

		return &ImageGenerationsUsageImage{
			Width:   width,
			Height:  height,
			Style:   lo.FromPtr(imageGenerationRequest.Style),
			Quality: lo.FromPtr(imageGenerationRequest.Quality),
		}
	})

	if r.options.meteringPolicy != nil {
		switch r.options.meteringPolicy.GetSizeFrom() {
		case
			v1alpha1.ClusterMeteringPolicy_SIZE_FROM_OUTPUT,
			v1alpha1.ClusterMeteringPolicy_SIZE_FROM_GREATEST:
			err := r.batchResolveImages(ctx)
			if err != nil {
				return err
			}

			r.Usage.Images = lo.Map(r.Usage.Images, func(image *ImageGenerationsUsageImage, index int) *ImageGenerationsUsageImage {
				switch r.options.meteringPolicy.GetSizeFrom() {
				case v1alpha1.ClusterMeteringPolicy_SIZE_FROM_OUTPUT:
					image.Width = uint64(r.Images[index].ImageConfig.Width)
					image.Height = uint64(r.Images[index].ImageConfig.Height)
				case v1alpha1.ClusterMeteringPolicy_SIZE_FROM_GREATEST:
					if imageGenerationRequest.Size != nil {
						requestResolution := imageGenerationRequest.Size.Width * imageGenerationRequest.Size.Height
						responseResolution := image.Width * image.Height

						if requestResolution > responseResolution {
							image.Width = imageGenerationRequest.Size.Width
							image.Height = imageGenerationRequest.Size.Height
						}
					}
				case
					v1alpha1.ClusterMeteringPolicy_SIZE_FROM_UNSPECIFIED,
					v1alpha1.ClusterMeteringPolicy_SIZE_FROM_INPUT:
					break
				default:
					break
				}

				return image
			})
		case v1alpha1.ClusterMeteringPolicy_SIZE_FROM_UNSPECIFIED:
			break
		case
			v1alpha1.ClusterMeteringPolicy_SIZE_FROM_INPUT:
			break
		default:
			break
		}
	}

	return nil
}

func (r *ImageGenerationsResponse) batchResolveImages(ctx context.Context) error {
	var wg sync.WaitGroup

	errResults := make([]error, len(r.Images))

	for index, imageObject := range r.Images {
		wg.Add(1)

		go func(ctx context.Context, index int, imageObject *ImageGenerationsImage) {
			err := imageObject.resolveImage(ctx, r.options.httpClient)
			if err != nil {
				errResults[index] = err
			}

			wg.Done()
		}(ctx, index, imageObject)
	}

	wg.Wait()

	return errors.Join(lo.Filter(errResults, func(err error, _ int) bool { return err != nil })...)
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
