package gateway

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"google.golang.org/protobuf/types/known/anypb"

	"google.golang.org/protobuf/proto"

	"knoway.dev/api/listeners/v1alpha1"
	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/listener"
	"knoway.dev/pkg/listener/manager/chat"
	"knoway.dev/pkg/listener/manager/image"
)

func StartGateway(_ context.Context, lifecycle bootkit.LifeCycle, listenerAddr string, cfg []*anypb.Any) error {
	if listenerAddr == "" {
		listenerAddr = ":8080"
	}
	if len(cfg) == 0 {
		return errors.New("no listener found")
	}
	mux := listener.NewMux()

	for _, c := range cfg {
		obj, err := anypb.UnmarshalNew(c, proto.UnmarshalOptions{})
		if err != nil {
			return err
		}

		switch obj.(type) {
		case *v1alpha1.ChatCompletionListener:
			mux.Register(chat.NewOpenAIChatListenerConfigs(obj, lifecycle))
		case *v1alpha1.ImageListener:
			mux.Register(image.NewOpenAIImageListenerConfigs(obj, lifecycle))
		default:
			return fmt.Errorf("%s is not a valid listener", c.GetTypeUrl())
		}
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
