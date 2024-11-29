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
	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/listener"
	"knoway.dev/pkg/listener/manager"
)

type GatewayConfig struct {
	AuthServerAddress    string
	GatewayListenAddress string
}

func StartGateway(_ context.Context, cfg GatewayConfig, lifecycle bootkit.LifeCycle) error {
	baseListenConfig := &v1alpha1.ChatCompletionListener{
		Name:    "openai",
		Filters: []*v1alpha1.ListenerFilter{},
	}

	if cfg.AuthServerAddress != "" {
		baseListenConfig.Filters = append(baseListenConfig.Filters, &v1alpha1.ListenerFilter{
			Name: "api-key-auth",
			Config: func() *anypb.Any {
				c, err := anypb.New(&v1alpha2.APIKeyAuthConfig{
					AuthServer: &v1alpha2.APIKeyAuthConfig_AuthServer{
						Url: cfg.AuthServerAddress,
					},
				})
				if err != nil {
					return nil
				}

				return c
			}(),
		})
	}

	addr := cfg.GatewayListenAddress
	if addr == "" {
		addr = ":8080" // default address
	}

	server, err := listener.NewMux().
		Register(manager.NewOpenAIChatCompletionsListenerWithConfigs(baseListenConfig)).
		Register(manager.NewOpenAIModelsListenerWithConfigs(baseListenConfig)).
		BuildServer(&http.Server{Addr: addr, ReadTimeout: time.Minute})
	if err != nil {
		return err
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	lifecycle.Append(bootkit.LifeCycleHook{
		OnStart: func(ctx context.Context) error {
			slog.Info("Starting gateway ...", "addr", ln.Addr().String())

			err := server.Serve(ln)
			if err != nil && err != http.ErrServerClosed {
				return err
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			slog.Info("Stopping gateway ...")

			err = server.Shutdown(ctx)
			if err != nil {
				return err
			}

			slog.Info("Gateway stopped gracefully.")

			return nil
		},
	})

	return nil
}
