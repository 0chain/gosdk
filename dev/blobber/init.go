package blobber

import (
	"net/http"

	"github.com/0chain/gosdk/dev/mock"
	"github.com/gorilla/mux"
)

func RegisterHandlers(r *mux.Router, m mock.ResponseMap) {
	r.HandleFunc("/v1/file/upload/{allocation}", uploadAndUpdateFile).Methods(http.MethodPut, http.MethodPost)
	r.HandleFunc("/v1/file/referencepath/{allocation}", getReference).Methods(http.MethodGet)
	r.HandleFunc("/v1/connection/commit/{allocation}", commitWrite).Methods(http.MethodPost)

	r.HandleFunc("/v1/writemarker/lock/{allocation}", mock.WithResponse(m)).Methods(http.MethodPost)
	r.HandleFunc("/v1/writemarker/lock/{allocation}", mock.WithResponse(m)).Methods(http.MethodDelete)
}
