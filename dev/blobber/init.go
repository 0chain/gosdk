package blobber

import (
	"github.com/0chain/common/constants/endpoint/v1_endpoint/blobber_endpoint"
	"net/http"

	"github.com/0chain/gosdk/dev/mock"
	"github.com/gorilla/mux"
)

func RegisterHandlers(r *mux.Router, m mock.ResponseMap) {
	r.HandleFunc(blobber_endpoint.FileUpload.PathWithPathVariable(), uploadAndUpdateFile).Methods(http.MethodPut, http.MethodPost)
	r.HandleFunc(blobber_endpoint.FileReferencePath.PathWithPathVariable(), getReference).Methods(http.MethodGet)
	r.HandleFunc(blobber_endpoint.ConnectionCommit.PathWithPathVariable(), commitWrite).Methods(http.MethodPost)

	r.HandleFunc(blobber_endpoint.WriteMarkerLock.PathWithPathVariable(), mock.WithResponse(m)).Methods(http.MethodPost)
	r.HandleFunc(blobber_endpoint.WriteMarkerLock.PathWithPathVariable(), mock.WithResponse(m)).Methods(http.MethodDelete)
	r.HandleFunc(blobber_endpoint.HashnodeRoot.PathWithPathVariable(), mock.WithResponse(m)).Methods(http.MethodGet)

	r.NotFoundHandler = Handle404(m)
}

// Handle404 ...
func Handle404(m mock.ResponseMap) http.Handler {
	return http.HandlerFunc(mock.WithResponse(m))
}
