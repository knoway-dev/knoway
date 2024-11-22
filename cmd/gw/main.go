package main

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"google.golang.org/protobuf/types/known/anypb"
	v1alpha2 "knoway.dev/api/filters/v1alpha1"
	"knoway.dev/api/listeners/v1alpha1"
	v1alpha3 "knoway.dev/api/route/v1alpha1"
	"knoway.dev/pkg/constants"
	"knoway.dev/pkg/listener/manager"
	"knoway.dev/pkg/registry/cluster"
	"knoway.dev/pkg/registry/route"
)

func main() {
	rConfig := &v1alpha3.Route{
		Name: "default",
		Matches: []*v1alpha3.Match{
			{
				Model: &v1alpha3.StringMatch{
					Match: &v1alpha3.StringMatch_Exact{
						Exact: "some",
					},
				},
			},
		},
		ClusterName: "openai/gpt-3.5-turbo",
		Filters:     nil,
	}
	if err := route.RegisterRouteWithConfig(rConfig); err != nil {
		panic(err)
	}

	// init cluster register
	// todo Share controller, no static file
	cluster.InitClusterRegister(constants.DefaultClusterConfigPath)

	//cConfig := &v1alpha4.Cluster{
	//	Name:              "openai/gpt-3.5-turbo",
	//	LoadBalancePolicy: v1alpha4.LoadBalancePolicy_ROUND_ROBIN,
	//	Upstream: &v1alpha4.Upstream{
	//		Url:    "https://api.openai.com/v1/chat/completions",
	//		Method: v1alpha4.Upstream_POST,
	//		Headers: []*v1alpha4.Upstream_Header{
	//			{
	//				Key:   "Authorization",
	//				Value: "Bearer sk-proj-",
	//			},
	//		},
	//	},
	//	TlsConfig: nil,
	//	Filters: []*v1alpha4.ClusterFilter{
	//		{
	//			Name: "openai-request-unmarshaller",
	//			Config: func() *anypb.Any {
	//				c, err := anypb.New(&v1alpha2.OpenAIRequestMarshallerConfig{})
	//				if err != nil {
	//					panic(err)
	//				}
	//				return c
	//			}(),
	//		},
	//		{
	//			Name: "model-name-mapping",
	//			Config: func() *anypb.Any {
	//				c, err := anypb.New(&v1alpha2.OpenAIModelNameRewriteConfig{
	//					ModelName: "gpt-3.5-turbo",
	//				})
	//				if err != nil {
	//					panic(err)
	//				}
	//				return c
	//			}(),
	//		},
	//		{
	//			Name: "openai-response-unmarshaller",
	//			Config: func() *anypb.Any {
	//				c, err := anypb.New(&v1alpha2.OpenAIResponseUnmarshallerConfig{})
	//				if err != nil {
	//					panic(err)
	//				}
	//				return c
	//			}(),
	//		},
	//	},
	//}
	//
	//err := RegisterClusterWithConfig("openai/gpt-3.5-turbo", cConfig)
	//if err != nil {
	//	panic(err)
	//}

	r := mux.NewRouter()
	l, err := manager.NewWithConfigs(&v1alpha1.ChatCompletionListener{
		Name: "openai",
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
	})
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}

	err = l.RegisterRoutes(r)
	if err != nil {
		log.Fatalf("Failed to register routes: %v", err)
	}

	http.Handle("/", r)
	slog.Info("Starting server on :8080")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
