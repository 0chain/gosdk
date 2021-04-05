package mocks

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type Blobber struct {
	ID         string
	URL        string
	server     *httptest.Server
	mapHandler map[string]http.Handler
}

func (b *Blobber) Close() {
	if b.server != nil {
		b.server.Close()
	}
}

func (b *Blobber) getMapHandler(t *testing.T) map[string]http.Handler {
	if b.mapHandler == nil {
		b.mapHandler = map[string]http.Handler{
			"/": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }),
		}
	}
	return b.mapHandler
}

func (b *Blobber) SetHandler(t *testing.T, path string, handler http.HandlerFunc) {
	b.mapHandler[path] = handler
}

func (b *Blobber) ResetHandler(t *testing.T) {
	b.Close()
	b.mapHandler = map[string]http.Handler{
		"/": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }),
	}
	_, _, nServer := NewHTTPServer(t, b.mapHandler, true)
	nServer.Listener.Close()
	port := strings.ReplaceAll(b.URL, "http://127.0.0.1", "")
	listener, err := net.Listen("tcp", port)
	assert.NoErrorf(t, err, "Error net.Listen() Unexpected error %v", err)
	nServer.Listener = listener
	nServer.Start()
	b.server = nServer
}

func NewBlobberHTTPServer(t *testing.T) (blobber *Blobber) {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	id := hex.EncodeToString(bytes)
	b := &Blobber{
		ID: id,
	}
	url, _, server := NewHTTPServer(t, b.getMapHandler(t))
	b.URL = url
	b.server = server
	return b
}
