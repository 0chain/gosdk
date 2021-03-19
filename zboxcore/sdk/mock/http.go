package mock

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func NewHTTPServer(t *testing.T, mapPathHandler map[string]http.Handler, unstarted ...bool) (url string, close func(), server *httptest.Server) {
	var newServerFn = httptest.NewServer
	if len(unstarted) > 0 && unstarted[0] {
		newServerFn = httptest.NewUnstartedServer
	}
	server = newServerFn(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mapPathHandler[r.URL.Path] == nil {
			w.WriteHeader(500)
			w.Write([]byte("Internal Server Error!"))
			return
		}
		mapPathHandler[r.URL.Path].ServeHTTP(w, r)
	}))
	return server.URL, server.Close, server
}
