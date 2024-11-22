package gw

import (
	"github.com/gorilla/mux"
	"google.golang.org/protobuf/types/known/anypb"
	"log/slog"
	"net/http"

	v1alpha2 "knoway.dev/api/filters/v1alpha1"
	"knoway.dev/api/listeners/v1alpha1"
	v1alpha3 "knoway.dev/api/route/v1alpha1"
	"knoway.dev/pkg/listener/manager"
	"knoway.dev/pkg/registry/route"
)

func StartProxy(stop chan struct{}) (err error) {
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
		return err
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
						return nil
					}
					return c
				}(),
			},
		},
	}
	r := mux.NewRouter()
	l, err := manager.NewWithConfigs(baseListenConfig)
	if err != nil {
		return err
	}
	if err = l.RegisterRoutes(r); err != nil {
		return err
	}

	modelsListen, err := manager.NewModelsManagerWithConfigs(baseListenConfig)
	if err != nil {
		return err
	}
	if err = modelsListen.RegisterRoutes(r); err != nil {
		return err
	}

	http.Handle("/", r)
	slog.Info("Starting server on :8080")

	server := &http.Server{Addr: ":8080"}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed: %v", err)
		}
	}()

	// Wait for graceful shutdown
	// This could be replaced with a more sophisticated signal handling
	// mechanism if needed.
	<-stop
	if err := server.Shutdown(nil); err != nil {
		slog.Error("Server shutdown failed: %v", err)
	}
	slog.Info("Server stopped gracefully.")
	return
}
