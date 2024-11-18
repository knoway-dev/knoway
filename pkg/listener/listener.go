package listener

import "github.com/gorilla/mux"

type Listener interface {
	RegisterRoutes(mux *mux.Router) error
	Drain() error
}
