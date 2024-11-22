package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONPathExecute(t *testing.T) {
	type testCase struct {
		name     string
		payload  map[string]any
		template string
		expected any
	}

	testCases := []testCase{
		{
			name: "model",
			payload: map[string]any{
				"model": "gpt-4o",
			},
			template: "{ .model }",
			expected: "gpt-4o",
		},
		{
			name: "message role",
			payload: map[string]any{
				"model": "gpt-4o",
				"messages": []any{
					map[string]any{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
			template: "{ .messages[0].role }",
			expected: "user",
		},
		{
			name: "message content",
			payload: map[string]any{
				"model": "gpt-4o",
				"messages": []any{
					map[string]any{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
			template: "{ .messages[0].content }",
			expected: "Hello",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, GetByJSONPath[string](tc.payload, tc.template))
		})
	}
}
