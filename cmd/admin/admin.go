package admin

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	"knoway.dev/pkg/bootkit"

	"google.golang.org/protobuf/encoding/protojson"

	"knoway.dev/api/admin/v1alpha1"

	"github.com/gorilla/mux"
	"github.com/samber/lo"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"knoway.dev/pkg/listener"
	"knoway.dev/pkg/registry/cluster"
	"knoway.dev/pkg/registry/route"
)

type debugListener struct {
	staticListeners []*anypb.Any
}

func NewAdminListener(staticListeners []*anypb.Any) (listener.Listener, error) {
	return &debugListener{staticListeners: staticListeners}, nil
}

func (d *debugListener) Drain(ctx context.Context) error {
	return nil
}

func (d *debugListener) HasDrained() bool {
	return false
}

func sliceToAny[T proto.Message](s []T) []*anypb.Any {
	anys := make([]*anypb.Any, 0, len(s))

	for _, v := range s {
		a, err := anypb.New(v)
		if err != nil {
			slog.Error("failed to convert to any", "err", err)
			continue
		}

		anys = append(anys, a)
	}

	return anys
}

func (d *debugListener) configDump(writer http.ResponseWriter, request *http.Request) {
	clusters := cluster.DebugDumpAllClusters()
	routes := route.DebugDumpAllRoutes()
	listeners := d.staticListeners
	cd := &v1alpha1.ConfigDump{
		Clusters:  sliceToAny(clusters),
		Routes:    sliceToAny(routes),
		Listeners: listeners,
	}
	bs := lo.Must1(protojson.MarshalOptions{
		Multiline:         true,
		Indent:            "  ",
		AllowPartial:      false,
		UseProtoNames:     false,
		UseEnumNumbers:    false,
		EmitUnpopulated:   false,
		EmitDefaultValues: false,
		Resolver:          nil,
	}.Marshal(cd))
	_, _ = writer.Write(bs)
}

func (d *debugListener) RegisterRoutes(mux *mux.Router) error {
	mux.HandleFunc("/config_dump", d.configDump)
	return nil
}

func NewAdminServer(_ context.Context, staticListeners []*anypb.Any, addr string, lifecycle bootkit.LifeCycle) error {
	m := listener.NewMux()
	m.Register(NewAdminListener(staticListeners))

	server, err := m.BuildServer(&http.Server{Addr: addr, ReadTimeout: time.Minute})
	if err != nil {
		return err
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	lifecycle.Append(bootkit.LifeCycleHook{
		OnStart: func(ctx context.Context) error {
			slog.Info("Starting admin server ...", "addr", ln.Addr().String())

			if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
				return err
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			slog.Info("Stopping admin server ...")

			if err := server.Shutdown(ctx); err != nil {
				return err
			}

			slog.Info("Admin server stopped gracefully.")
			return nil
		},
	})

	return nil
}
