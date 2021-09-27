// package dev providers tools for local development
package dev

import (
	"net/http/httptest"

	"github.com/0chain/gosdk/dev/blobber"
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
func NewBlobberServer() *Server {
	s := NewServer()

	blobber.RegisterHandlers(s.Router)

	return s
}
