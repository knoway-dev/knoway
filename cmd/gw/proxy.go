package gw

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"google.golang.org/protobuf/types/known/anypb"
	v1alpha2 "knoway.dev/api/filters/v1alpha1"
	"knoway.dev/api/listeners/v1alpha1"
	v1alpha3 "knoway.dev/api/route/v1alpha1"
	"knoway.dev/pkg/listener/manager"
	"knoway.dev/pkg/registry/route"
)

func StartProxy() {
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

	baseListenConfig := &v1alpha1.ChatCompletionListener{
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
	}
	r := mux.NewRouter()
	l, err := manager.NewWithConfigs(baseListenConfig)
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}
	err = l.RegisterRoutes(r)
	if err != nil {
		log.Fatalf("Failed to register routes: %v", err)
	}

	modelsListen, err := manager.NewModelsManagerWithConfigs(baseListenConfig)
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}
	err = modelsListen.RegisterRoutes(r)
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
