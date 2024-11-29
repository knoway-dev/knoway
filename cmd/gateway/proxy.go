package gateway

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	"knoway.dev/config"

	"google.golang.org/protobuf/types/known/anypb"

	v1alpha2 "knoway.dev/api/filters/v1alpha1"
	"knoway.dev/api/listeners/v1alpha1"
	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/listener"
	"knoway.dev/pkg/listener/manager"
)

func StartGateway(_ context.Context, lifecycle bootkit.LifeCycle, cfg config.GatewayConfig) error {
	if cfg.ListenerAddr == "" {
		cfg.ListenerAddr = ":8080"
	}

	baseListenConfig := &v1alpha1.ChatCompletionListener{
		Name:    "openai",
		Filters: []*v1alpha1.ListenerFilter{},
	}

	if cfg.AuthServer.Url != "" {
		baseListenConfig.Filters = append(baseListenConfig.Filters, &v1alpha1.ListenerFilter{
			Name: "api-key-auth",
			Config: func() *anypb.Any {
				c, err := anypb.New(&v1alpha2.APIKeyAuthConfig{
					AuthServer: &v1alpha2.APIKeyAuthConfig_AuthServer{
						Url: cfg.AuthServer.Url,
					},
				})
				if err != nil {
					return nil
				}

				return c
			}(),
		})
	}

	if cfg.StatsServer.Url != "" {
		baseListenConfig.Filters = append(baseListenConfig.Filters, &v1alpha1.ListenerFilter{
			Config: func() *anypb.Any {
				c, err := anypb.New(&v1alpha2.UsageStatsConfig{
					StatsServer: &v1alpha2.UsageStatsConfig_StatsServer{
						Url: cfg.StatsServer.Url,
					},
				})
				if err != nil {
					return nil
				}

				return c
			}(),
		})
	}

	server, err := listener.NewMux().
		Register(manager.NewOpenAIChatCompletionsListenerWithConfigs(baseListenConfig, lifecycle)).
		Register(manager.NewOpenAIModelsListenerWithConfigs(baseListenConfig, lifecycle)).
		BuildServer(&http.Server{Addr: cfg.ListenerAddr, ReadTimeout: time.Minute})
	if err != nil {
		return err
	}

	ln, err := net.Listen("tcp", cfg.ListenerAddr)
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
