package openai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	jsonpatch "github.com/evanphx/json-patch/v5"

	"knoway.dev/pkg/utils"
)

func modifyBufferBodyAndParsed(buffer *bytes.Buffer, applyOpt *jsonpatch.ApplyOptions, patches ...*JSONPatchOperationObject) (*bytes.Buffer, map[string]any, error) {
	patch, err := jsonpatch.DecodePatch(NewPatches(patches...))
	if err != nil {
		return nil, nil, err
	}
	if applyOpt == nil {
		applyOpt = jsonpatch.NewApplyOptions()
	}

	patched, err := patch.ApplyWithOptions(buffer.Bytes(), applyOpt)
	if err != nil {
		return nil, nil, err
	}

	buffer = bytes.NewBuffer(patched)

	var newParsed map[string]any

	err = json.Unmarshal(patched, &newParsed)
	if err != nil {
		return nil, nil, err
	}

	return buffer, newParsed, nil
}

func modifyBytesBodyAndParsed(bytes []byte, patches ...*JSONPatchOperationObject) ([]byte, map[string]any, error) {
	patch, err := jsonpatch.DecodePatch(NewPatches(patches...))
	if err != nil {
		return nil, nil, err
	}

	patched, err := patch.Apply(bytes)
	if err != nil {
		return nil, nil, err
	}

	var newParsed map[string]any

	err = json.Unmarshal(patched, &newParsed)
	if err != nil {
		return nil, nil, err
	}

	return patched, newParsed, nil
}

func unmarshalErrorResponseFromParsedBody(body map[string]any, response *http.Response, bs []byte) (*ErrorResponse, error) {
	// For general cases, errors will be returned as a map with "error" property
	respErrMap := utils.GetByJSONPath[map[string]any](body, "{ .error }")
	// For OpenRouter, endpoint not found errors will be returned as a string with "error" property
	errorStringMap := utils.GetByJSONPath[string](body, "{ .error }")
	// For vLLM, {"object":"error", "message":"error message"} would be returned when error occurs
	objectString := utils.GetByJSONPath[string](body, "{ .object }")

	if len(respErrMap) > 0 {
		respErr, err := utils.FromMap[ErrorResponse](respErrMap)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal error: %w", err)
		}

		respErr.Status = response.StatusCode
		respErr.FromUpstream = true

		return respErr, nil
	} else if errorStringMap != "" {
		slog.Error("unknown unexpected error response returned",
			slog.String("body", string(bs)),
			slog.String("uri", response.Request.RequestURI),
			slog.String("url", response.Request.URL.String()),
		)

		respErr := &ErrorResponse{
			Status: response.StatusCode,
			ErrorBody: &Error{
				Message: "upstream error: " + errorStringMap,
			},
			Cause: errors.New("unknown error"),
		}

		respErr.FromUpstream = true

		return respErr, nil
	} else if response.StatusCode >= 400 && response.StatusCode < 600 {
		// TODO: should split vLLM, OpenRouter, and OpenAI into different dedicated
		// types of implementations to object types to handle different responses
		// instead of this messy if-else block
		if objectString == "error" {
			if body["message"] != "" {
				respErr := &ErrorResponse{
					Status: response.StatusCode,
					ErrorBody: &Error{
						Code:    utils.GetByJSONPath[*string](body, "{ .code }"),
						Message: utils.GetByJSONPath[string](body, "{ .message }"),
						Param:   utils.GetByJSONPath[*string](body, "{ .param }"),
						Type:    utils.GetByJSONPath[string](body, "{ .type }"),
					},
				}

				respErr.FromUpstream = true

				return respErr, nil
			}
		} else {
			slog.Error("unknown unexpected error response with unknown body structure returned",
				slog.String("body", string(bs)),
				slog.String("uri", response.Request.RequestURI),
				slog.String("url", response.Request.URL.String()),
			)

			respErr := &ErrorResponse{
				Status: response.StatusCode,
				ErrorBody: &Error{
					Message: "upstream unknown error: " + response.Status,
				},
				Cause: errors.New("unknown error"),
			}

			respErr.FromUpstream = true

			return respErr, nil
		}
	}

	return nil, nil
}
