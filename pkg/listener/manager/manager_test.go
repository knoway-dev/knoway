package manager

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	v1alpha2 "knoway.dev/api/filters/v1alpha1"
	"knoway.dev/api/listeners/v1alpha1"
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/filters/auth"
	"knoway.dev/pkg/listener"
)

func TestNewWithConfigs(t *testing.T) {
	tests := []struct {
		name    string
		cfg     proto.Message
		want    listener.Listener
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &v1alpha1.ChatCompletionListener{
				Filters: []*v1alpha1.ListenerFilter{
					{
						Name: "api-key-auth",
						Config: func() *anypb.Any {
							c, err := anypb.New(&v1alpha2.APIKeyAuthConfig{
								AuthServer: nil,
							})
							if err != nil {
								panic(err)
							}
							return c
						}(),
					},
				},
			},
			want: &OpenAIChatCompletionListener{
				cfg: &v1alpha1.ChatCompletionListener{
					Filters: []*v1alpha1.ListenerFilter{
						{
							Name: "api-key-auth",
							Config: func() *anypb.Any {
								c, err := anypb.New(&v1alpha2.APIKeyAuthConfig{
									AuthServer: nil,
								})
								if err != nil {
									panic(err)
								}
								return c
							}(),
						},
					},
				},
				filters: []filters.RequestFilter{
					&auth.AuthFilter{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWithConfigs(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWithConfigs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, got, "NewWithConfigs()")
		})
	}
}
