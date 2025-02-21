package gateway

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	v1alpha2 "knoway.dev/api/filters/v1alpha1"
	"knoway.dev/api/listeners/v1alpha1"
	"knoway.dev/config"
	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/listener"
	"knoway.dev/pkg/listener/manager/chat"
	"knoway.dev/pkg/listener/manager/image"
)

func createListenerFilter(name string, config *anypb.Any) *v1alpha1.ListenerFilter {
	return &v1alpha1.ListenerFilter{
		Name:   name,
		Config: config,
	}
}

func createAuthFilter(url string, timeout int64) *v1alpha1.ListenerFilter {
	c, err := anypb.New(&v1alpha2.APIKeyAuthConfig{
		AuthServer: &v1alpha2.APIKeyAuthConfig_AuthServer{
			Url:     url,
			Timeout: durationpb.New(time.Duration(timeout) * time.Second),
		},
	})
	if err != nil {
		return nil
	}

	return createListenerFilter("api-key-auth", c)
}

func createRateLimitFilter() *v1alpha1.ListenerFilter {
	c, err := anypb.New(&v1alpha2.RateLimitConfig{})
	if err != nil {
		return nil
	}

	return createListenerFilter("rate-limit", c)
}

func createStatsFilter(url string, timeout int64) *v1alpha1.ListenerFilter {
	c, err := anypb.New(&v1alpha2.UsageStatsConfig{
		StatsServer: &v1alpha2.UsageStatsConfig_StatsServer{
			Url:     url,
			Timeout: durationpb.New(time.Duration(timeout) * time.Second),
		},
	})
	if err != nil {
		return nil
	}

	return createListenerFilter("stats-usage", c)
}

func newChatListenerConfig(cfg config.GatewayConfig) *v1alpha1.ChatCompletionListener {
	baseListenConfig := &v1alpha1.ChatCompletionListener{
		Name:    "openai-chat",
		Filters: []*v1alpha1.ListenerFilter{},
	}

	if cfg.Log.AccessLog != nil {
		baseListenConfig.AccessLog = &v1alpha1.Log{
			Enable: cfg.Log.AccessLog.Enabled,
		}
	}

	if cfg.AuthServer.URL != "" {
		if filter := createAuthFilter(cfg.AuthServer.URL, cfg.AuthServer.Timeout); filter != nil {
			baseListenConfig.Filters = append(baseListenConfig.Filters, filter)
		}
	}

	if cfg.Policy.RateLimit.Enabled {
		if filter := createRateLimitFilter(); filter != nil {
			baseListenConfig.Filters = append(baseListenConfig.Filters, filter)
		}
	}

	if cfg.StatsServer.URL != "" {
		if filter := createStatsFilter(cfg.StatsServer.URL, cfg.StatsServer.Timeout); filter != nil {
			baseListenConfig.Filters = append(baseListenConfig.Filters, filter)
		}
	}

	return baseListenConfig
}

func newImageListenerConfig(cfg config.GatewayConfig) *v1alpha1.ImageListener {
	baseListenConfig := &v1alpha1.ImageListener{
		Name:    "openai-image",
		Filters: []*v1alpha1.ListenerFilter{},
	}

	if cfg.Log.AccessLog != nil {
		baseListenConfig.AccessLog = &v1alpha1.Log{
			Enable: cfg.Log.AccessLog.Enabled,
		}
	}

	if cfg.AuthServer.URL != "" {
		if filter := createAuthFilter(cfg.AuthServer.URL, cfg.AuthServer.Timeout); filter != nil {
			baseListenConfig.Filters = append(baseListenConfig.Filters, filter)
		}
	}

	if cfg.Policy.RateLimit.Enabled {
		if filter := createRateLimitFilter(); filter != nil {
			baseListenConfig.Filters = append(baseListenConfig.Filters, filter)
		}
	}

	if cfg.StatsServer.URL != "" {
		if filter := createStatsFilter(cfg.StatsServer.URL, cfg.StatsServer.Timeout); filter != nil {
			baseListenConfig.Filters = append(baseListenConfig.Filters, filter)
		}
	}

	return baseListenConfig
}

func StartGateway(_ context.Context, lifecycle bootkit.LifeCycle, listenerAddr string, cfg config.GatewayConfig) error {
	if listenerAddr == "" {
		listenerAddr = ":8080"
	}

	server, err := listener.NewMux().
		Register(chat.NewOpenAIChatListenerConfigs(newChatListenerConfig(cfg), lifecycle)).
		Register(image.NewOpenAIImageListenerConfigs(newImageListenerConfig(cfg), lifecycle)).
		BuildServer(&http.Server{
			Addr:        listenerAddr,
			ReadTimeout: time.Minute,
		})
	if err != nil {
		return err
	}

	ln, err := net.Listen("tcp", listenerAddr)
	if err != nil {
		return err
	}

	lifecycle.Append(bootkit.LifeCycleHook{
		OnStart: func(ctx context.Context) error {
			slog.Info("Starting gateway ...", "addr", ln.Addr().String())

			if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
				return err
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			slog.Info("Stopping gateway ...")

			if err := server.Shutdown(ctx); err != nil {
				return err
			}

			slog.Info("Gateway stopped gracefully.")
			return nil
		},
	})

	return nil
}
