package mocks

import (
	"net/http"
	"testing"
)

type sharder struct {
	mapHandler map[string]http.Handler
}

var s *sharder

func (s *sharder) getMapHandler() map[string]http.Handler {
	if s.mapHandler == nil {
		s.mapHandler = map[string]http.Handler{
			"/": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		}
	}
	return s.mapHandler
}

func SetSharderHandler(t *testing.T, path string, handler http.HandlerFunc) {
	if s == nil {
		t.Error("sharder http server is not initialized")
	}
	s.mapHandler[path] = handler
}

func NewSharderHTTPServer(t *testing.T) (url string, close func()) {
	s = &sharder{}
	url, close, _ = NewHTTPServer(t, s.getMapHandler())
	return url, close
}
