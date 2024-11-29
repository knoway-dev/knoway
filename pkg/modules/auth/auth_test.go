package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanAccessModel(t *testing.T) {
	tests := []struct {
		name         string
		allowModels  []string
		requestModel string
		want         bool
	}{
		{
			name:         "t1",
			allowModels:  []string{"*/*"},
			requestModel: "kebe/model-1",
			want:         true,
		},
		{
			name:         "t2",
			allowModels:  []string{"*/*"},
			requestModel: "gpt-1",
			want:         true,
		},
		{
			name:         "t3",
			allowModels:  []string{"kebe/*"},
			requestModel: "kebe/model-1",
			want:         true,
		},
		{
			name:         "t4",
			allowModels:  []string{"kebe/*"},
			requestModel: "gpt-1",
			want:         false,
		},
		{
			name:         "t4",
			allowModels:  []string{"kebe/*"},
			requestModel: "nicole/gpt-1",
			want:         false,
		},
		{
			name:         "only public ",
			allowModels:  []string{"*"},
			requestModel: "nicole/gpt-1",
			want:         false,
		},
		{
			name:         "* match public",
			allowModels:  []string{"*"},
			requestModel: "gpt-1",
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, CanAccessModel(tt.allowModels, tt.requestModel), "CanAccessModel(%v, %v)", tt.allowModels, tt.requestModel)
		})
	}
}
