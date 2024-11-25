package gw

import (
	"google.golang.org/protobuf/types/known/anypb"

	v1alpha4 "knoway.dev/api/clusters/v1alpha1"
	v1alpha2 "knoway.dev/api/filters/v1alpha1"
)

var StaticClustersConfig = map[string]*v1alpha4.Cluster{
	"openai/gpt-3.5-turbo": {
		Name:              "openai/gpt-3.5-turbo",
		Provider:          "openai",
		LoadBalancePolicy: v1alpha4.LoadBalancePolicy_ROUND_ROBIN,
		Upstream: &v1alpha4.Upstream{
			Url:    "https://openrouter.ai/api/v1/chat/completions",
			Method: v1alpha4.Upstream_POST,
			Headers: []*v1alpha4.Upstream_Header{
				{
					Key:   "Authorization",
					Value: "Bearer sk-or-v1-",
				},
			},
		},
		TlsConfig: nil,
		Filters: []*v1alpha4.ClusterFilter{
			{
				Name: "openai-request-unmarshaller",
				Config: func() *anypb.Any {
					c, err := anypb.New(&v1alpha2.OpenAIRequestMarshallerConfig{})
					if err != nil {
						panic(err)
					}
					return c
				}(),
			},
			{
				Name: "model-name-mapping",
				Config: func() *anypb.Any {
					c, err := anypb.New(&v1alpha2.OpenAIModelNameRewriteConfig{
						ModelName: "gpt-3.5-turbo",
					})
					if err != nil {
						panic(err)
					}
					return c
				}(),
			},
			{
				Name: "openai-response-unmarshaller",
				Config: func() *anypb.Any {
					c, err := anypb.New(&v1alpha2.OpenAIResponseUnmarshallerConfig{})
					if err != nil {
						panic(err)
					}
					return c
				}(),
			},
		},
	},
}
