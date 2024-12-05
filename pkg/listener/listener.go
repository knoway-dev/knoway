package listener

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

type Drainable interface {
	Drain(ctx context.Context) error
	HasDrained() bool
}

type Listener interface {
	Drainable

	RegisterRoutes(mux *mux.Router) error
}

type Mux struct {
	*mux.Router

	errors []error
	mutex  sync.Mutex
}

func NewMux() *Mux {
	return &Mux{
		Router: mux.NewRouter(),
		errors: make([]error, 0),
	}
}

func (r *Mux) Error() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return errors.Join(r.errors...)
}

func (r *Mux) Register(listener Listener, err error) *Mux {
	if err != nil {
		r.mutex.Lock()
		r.errors = append(r.errors, err)
		r.mutex.Unlock()

		return r
	}

	err = listener.RegisterRoutes(r.Router)
	if err != nil {
		r.mutex.Lock()
		r.errors = append(r.errors, err)
		r.mutex.Unlock()
	}

	return r
}

func (r *Mux) BuildServer(server *http.Server) (*http.Server, error) {
	if len(r.errors) > 0 {
		return nil, r.Error()
	}

	server.Handler = r.Router

	return server, nil
}
