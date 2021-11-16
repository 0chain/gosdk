package httpwasm

import (
	"net/http/httptest"

	"github.com/gorilla/mux"
)

type Server struct {
	*httptest.Server
	*mux.Router
}

func NewServer() *Server {
	router := mux.NewRouter()
	s := &Server{
		Router: router,
		Server: httptest.NewServer(router),
	}

	return s
}

// NewMinerServer create a local dev miner server
func NewMinerServer() *Server {
	s := NewServer()

	RegisterMinerHandlers(s.Router.PathPrefix("/miner01").Subrouter())
	RegisterMinerHandlers(s.Router.PathPrefix("/miner02").Subrouter())
	RegisterMinerHandlers(s.Router.PathPrefix("/miner03").Subrouter())

	return s
}

// NewSharderServer create a local dev sharder server
func NewSharderServer() *Server {
	s := NewServer()

	RegisterSharderHandlers(s.Router.PathPrefix("/sharder01").Subrouter())
	RegisterSharderHandlers(s.Router.PathPrefix("/sharder02").Subrouter())
	RegisterSharderHandlers(s.Router.PathPrefix("/sharder03").Subrouter())
	return s
}

// NewSharderServer create a local dev sharder server
func NewDefaultServer() *Server {
	s := NewServer()

	RegisterDefaultHandlers(s.Router.PathPrefix("/").Subrouter())

	return s
}
