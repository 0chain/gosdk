package blobber

import (
	"net/http"

	"github.com/gorilla/mux"
)

var createDirResults = make(map[string]int)

func MockCreateDir(allocationTx, name string, statusCode int) {
	createDirResults[allocationTx+":"+name] = statusCode
}

func UnmockCreateDir(allocationTx, name string) {
	delete(createDirResults, allocationTx+":"+name)
}

func createDir(w http.ResponseWriter, req *http.Request) {
	var vars = mux.Vars(req)
	allocationTx := vars["allocation"]
	name := req.FormValue("name")

	statusCode, ok := createDirResults[allocationTx+":"+name]

	if ok {
		w.WriteHeader(statusCode)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

}
