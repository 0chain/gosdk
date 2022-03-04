package mock

import "net/http"

type ResponseMap map[string]Response

type Response struct {
	StatusCode int
	Body       []byte
}

// WithResponse mock respone
func WithResponse(m ResponseMap) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer w.Header().Set("Content-Type", "application/json")

		if m != nil {
			key := r.Method + ":" + r.URL.Path
			resp, ok := m[key]

			if ok {

				w.WriteHeader(resp.StatusCode)
				if resp.Body != nil {
					w.Write(resp.Body) //nolint: errcheck
				}

				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
	}
}
