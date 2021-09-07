package blobber

import (
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterHandlers(s *mux.Router) {
	s.HandleFunc("/v1/file/upload/{allocation}", uploadAndUpdateFile).Methods(http.MethodPut, http.MethodPost)
}
