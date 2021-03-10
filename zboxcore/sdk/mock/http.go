package mock

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func NewHTTPServer(t *testing.T, mapPathHandler map[string]http.Handler) (url string, close func()) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mapPathHandler[r.URL.Path] != nil {
			mapPathHandler[r.URL.Path].ServeHTTP(w, r)
			return
		}
	}))
	return server.URL, server.Close
}