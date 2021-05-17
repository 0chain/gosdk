package mocks

import (
	"net/http"
	"testing"
)

type miner struct {
	mapHandler map[string]http.Handler
}

var m *miner

func (m *miner) getMapHandler() map[string]http.Handler {
	if m.mapHandler == nil {
		m.mapHandler = map[string]http.Handler{
			"/": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		}
	}
	return m.mapHandler
}

func SetMinerHandler(t *testing.T, path string, handler http.HandlerFunc) {
	if s == nil {
		t.Error("miner http server is not initialized")
	}
	m.mapHandler[path] = handler
}

func NewMinerHTTPServer(t *testing.T) (url string, close func()) {
	m = &miner{}
	url, close, _ = NewHTTPServer(t, m.getMapHandler())
	return url, close
}
