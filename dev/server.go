// Providers tools for local development - do not use.
package dev

import (
	"net/http/httptest"

	"github.com/0chain/gosdk/dev/blobber"
	"github.com/0chain/gosdk/dev/mock"
	"github.com/gorilla/mux"
)

// Server a local dev server to mock server APIs
type Server struct {
	*httptest.Server
	*mux.Router
}

// NewServer create a local dev server
func NewServer() *Server {
	router := mux.NewRouter()
	s := &Server{
		Router: router,
		Server: httptest.NewServer(router),
	}

	return s
}

// NewBlobberServer create a local dev blobber server
func NewBlobberServer(m mock.ResponseMap) *Server {
	s := NewServer()

	blobber.RegisterHandlers(s.Router, m)

	return s
}
