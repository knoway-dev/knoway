package openai

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorUnmarshalJSON(t *testing.T) {
	t.Run("OpenRouter", func(t *testing.T) {
		errorPayload := map[string]any{
			"error": map[string]any{
				"message": "Invalid credentials",
				"code":    401,
			},
		}

		errorJSON, err := json.Marshal(errorPayload)
		require.NoError(t, err)
		require.NotEmpty(t, errorJSON)

		var e ErrorResponse
		err = e.UnmarshalJSON(errorJSON)
		require.NoError(t, err)

		assert.Equal(t, "Invalid credentials", e.ErrorBody.Message)
		assert.Equal(t, "401", *e.ErrorBody.Code)
		assert.Empty(t, e.ErrorBody.Param)
		assert.Empty(t, e.ErrorBody.Type)
	})

	t.Run("OpenAI", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			errorPayload := map[string]any{
				"error": map[string]any{
					"message": "Incorrect API key provided: sk-abcd. You can find your API key at https://platform.openai.com/account/api-keys.",
					"type":    "invalid_request_error",
					"param":   nil,
					"code":    "invalid_api_key",
				},
			}

			errorJSON, err := json.Marshal(errorPayload)
			require.NoError(t, err)
			require.NotEmpty(t, errorJSON)

			var e ErrorResponse
			err = e.UnmarshalJSON(errorJSON)
			require.NoError(t, err)

			assert.Equal(t, "Incorrect API key provided: sk-abcd. You can find your API key at https://platform.openai.com/account/api-keys.", e.ErrorBody.Message)
			assert.Equal(t, "invalid_api_key", *e.ErrorBody.Code)
			assert.Empty(t, e.ErrorBody.Param)
			assert.Equal(t, "invalid_request_error", e.ErrorBody.Type)
		})
	})

	t.Run("Ollama", func(t *testing.T) {
		errorPayload := map[string]any{
			"error": map[string]any{
				"message": "model is required",
				"type":    "api_error",
				"param":   nil,
				"code":    nil,
			},
		}

		errorJSON, err := json.Marshal(errorPayload)
		require.NoError(t, err)
		require.NotEmpty(t, errorJSON)

		var e ErrorResponse
		err = e.UnmarshalJSON(errorJSON)
		require.NoError(t, err)

		assert.Equal(t, "model is required", e.ErrorBody.Message)
		assert.Nil(t, e.ErrorBody.Code)
		assert.Empty(t, e.ErrorBody.Param)
		assert.Equal(t, "api_error", e.ErrorBody.Type)
	})
}
