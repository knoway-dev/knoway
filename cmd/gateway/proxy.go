package gateway

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/samber/lo"
	"gopkg.in/yaml.v3"

	"buf.build/go/protoyaml"

	"knoway.dev/api/listeners/v1alpha1"
	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/listener"
	"knoway.dev/pkg/listener/manager/chat"
	"knoway.dev/pkg/listener/manager/image"
)

func StartGateway(_ context.Context, lifecycle bootkit.LifeCycle, listenerAddr string, cfg map[string]map[string]interface{}) error {
	if listenerAddr == "" {
		listenerAddr = ":8080"
	}
	mux := listener.NewMux()
	exists := false
	if v, ok := cfg["chat"]; ok { //nolint: wsl
		bs, _ := yaml.Marshal(v)
		l := new(v1alpha1.ChatCompletionListener)
		lo.Must0(protoyaml.Unmarshal(bs, l))
		mux.Register(chat.NewOpenAIChatListenerConfigs(l, lifecycle))
		exists = true
	}
	if v, ok := cfg["image"]; ok {
		bs, _ := yaml.Marshal(v)
		l := new(v1alpha1.ImageListener)
		lo.Must0(protoyaml.Unmarshal(bs, l))
		mux.Register(image.NewOpenAIImageListenerConfigs(l, lifecycle))
		exists = true
	}
	if !exists {
		return errors.New("no listener found")
	}

	server, err := mux.BuildServer(&http.Server{Addr: listenerAddr, ReadTimeout: time.Minute})
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
