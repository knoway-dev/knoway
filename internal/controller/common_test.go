package controller

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"

	"knoway.dev/api/v1alpha1"
)

func TestProcessStruct_OpenAIChatParams(t *testing.T) {
	tests := []struct {
		name        string
		input       *v1alpha1.ModelParams
		expected    map[string]*structpb.Value
		expectError bool
	}{
		{
			name: "Valid ChatRequest with new params",
			input: &v1alpha1.ModelParams{
				OpenAI: &v1alpha1.OpenAIParam{
					CommonParams: v1alpha1.CommonParams{
						Model:       "gpt-3.5-turbo",
						Temperature: lo.ToPtr("0.7"),
					},
					MaxTokens:           lo.ToPtr(100),
					MaxCompletionTokens: lo.ToPtr(200),
					TopP:                lo.ToPtr("0.3"),
					Stream:              lo.ToPtr(true),
					StreamOptions: &v1alpha1.StreamOptions{
						IncludeUsage: lo.ToPtr(true),
					},
				},
			},
			expected: map[string]*structpb.Value{
				"model":                 structpb.NewStringValue("gpt-3.5-turbo"),
				"temperature":           structpb.NewNumberValue(0.7),
				"max_tokens":            structpb.NewNumberValue(100),
				"max_completion_tokens": structpb.NewNumberValue(200),
				"top_p":                 structpb.NewNumberValue(0.3),
				"stream":                structpb.NewBoolValue(true),
				"stream_options": structpb.NewStructValue(&structpb.Struct{
					Fields: map[string]*structpb.Value{
						"include_usage": structpb.NewBoolValue(true),
					},
				}),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := make(map[string]*structpb.Value)

			// Call the function under test
			err := parseModelParams(tt.input, params)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				// Validate the result
				assert.Equal(t, tt.expected, params)
			}
		})
	}
}
