package gw

import (
	"context"
	"log/slog"
	"net"
	"net/http"

	"google.golang.org/protobuf/types/known/anypb"

	v1alpha2 "knoway.dev/api/filters/v1alpha1"
	"knoway.dev/api/listeners/v1alpha1"
	v1alpha3 "knoway.dev/api/route/v1alpha1"
	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/listener"
	"knoway.dev/pkg/listener/manager"
	"knoway.dev/pkg/registry/route"
)

func StartGateway(ctx context.Context, lifecycle bootkit.LifeCycle) error {
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

	lifecycle.Append(bootkit.LifeCycleHook{
		OnStart: func(ctx context.Context) error {
			slog.Info("Starting server...", "addr", ln.Addr().String())

			err := server.Serve(ln)
			if err != nil && err != http.ErrServerClosed {
				return err
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			slog.Info("Stopping server...")

			err = server.Shutdown(ctx)
			if err != nil {
				return err
			}

			slog.Info("Server stopped gracefully.")

			return nil
		},
	})

	return nil
}
