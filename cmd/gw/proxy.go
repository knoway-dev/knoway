package gw

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	"google.golang.org/protobuf/types/known/anypb"

	v1alpha2 "knoway.dev/api/filters/v1alpha1"
	"knoway.dev/api/listeners/v1alpha1"
	v1alpha3 "knoway.dev/api/route/v1alpha1"
	"knoway.dev/pkg/listener"
	"knoway.dev/pkg/listener/manager"
	"knoway.dev/pkg/registry/route"
)

func StartProxy(stop chan struct{}) error {
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

	server, err := listener.NewMux().
		Register(manager.NewOpenAIChatCompletionsListenerWithConfigs(baseListenConfig)).
		Register(manager.NewOpenAIModelsListenerWithConfigs(baseListenConfig)).
		BuildServer(&http.Server{Addr: ":8080"})
	if err != nil {
		return err
	}

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}

	slog.Info("Starting server on :8080")

	go func() {
		err := server.Serve(ln)
		if err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed", "error", err)
		}
	}()

	// Wait for graceful shutdown
	// This could be replaced with a more sophisticated signal handling
	// mechanism if needed.
	<-stop

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10) // TODO: how long?
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		slog.Error("Server shutdown failed", "error", err)
	}

	slog.Info("Server stopped gracefully.")

	return nil
}
