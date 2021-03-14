package mock

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"testing"
)

type Blobber struct {
	ID         string
	URL        string
	mapHandler map[string]http.Handler
}

func (b *Blobber) getMapHandler() map[string]http.Handler {
	if b.mapHandler == nil {
		b.mapHandler = map[string]http.Handler{
			"/": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		}
	}
	return b.mapHandler
}

func (b *Blobber) SetHandler(t *testing.T, path string, handler http.HandlerFunc) {
	b.mapHandler[path] = handler
}

func NewBlobberHTTPServer(t *testing.T) (blobber *Blobber, close func()) {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	id := hex.EncodeToString(bytes)
	b := &Blobber{
		ID: id,
	}
	url, close := NewHTTPServer(t, b.getMapHandler())
	b.URL = url
	return b, close
}
