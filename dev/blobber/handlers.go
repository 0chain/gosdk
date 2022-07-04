package blobber

import (
	"encoding/json"
	"net/http"

	"github.com/0chain/gosdk/dev/blobber/model"
	"github.com/gorilla/mux"
)

func uploadAndUpdateFile(w http.ResponseWriter, req *http.Request) {
	uploadMeta := req.FormValue("uploadMeta")

	var form *model.UploadFormData
	err := json.Unmarshal([]byte(uploadMeta), &form)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//nolint: errcheck
	json.NewEncoder(w).Encode(&model.UploadResult{
		Filename:   form.Filename,
		Hash:       form.ChunkHash,
		MerkleRoot: form.MerkleRoot,
	})

}

func getReference(w http.ResponseWriter, req *http.Request) {

	var vars = mux.Vars(req)

	alloctionID := vars["allocation"]

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	result, ok := referencePathResults[alloctionID]

	if ok {
		buf, _ := json.Marshal(result)
		//nolint: errcheck
		w.Write(buf)
		return
	}

	rootRefs := `{"meta_data":{"chunk_size":0,"created_at":"0001-01-01T00:00:00Z","hash":"","lookup_hash":"","name":"/","num_of_blocks":0,"path":"/","path_hash":"","size":0,"type":"d","updated_at":"0001-01-01T00:00:00Z"},"Ref":{"ID":0,"Type":"d","AllocationID":"` + vars["allocation"] + `","LookupHash":"","Name":"/","Path":"/","Hash":"","NumBlocks":0,"PathHash":"","ParentPath":"","PathLevel":1,"CustomMeta":"","ContentHash":"","Size":0,"MerkleRoot":"","ActualFileSize":0,"ActualFileHash":"","MimeType":"","WriteMarker":"","ThumbnailSize":0,"ThumbnailHash":"","ActualThumbnailSize":0,"ActualThumbnailHash":"","EncryptedKey":"","Children":null,"OnCloud":false,"CommitMetaTxns":null,"CreatedAt":"0001-01-01T00:00:00Z","UpdatedAt":"0001-01-01T00:00:00Z","DeletedAt":"0001-01-01T00:00:00Z","ChunkSize":0},"latest_write_marker":null}`

	//nolint: errcheck
	w.Write([]byte(rootRefs))
}

func commitWrite(w http.ResponseWriter, req *http.Request) {

	//	var vars = mux.Vars(req)

	writeMarker := &model.WriteMarker{}
	err := json.Unmarshal([]byte(req.FormValue("write_marker")), writeMarker)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := &model.CommitResult{}
	result.AllocationRoot = writeMarker.AllocationRoot
	result.Success = true
	result.WriteMarker = writeMarker

	json.NewEncoder(w).Encode(result)
}

func mockRespone(w http.ResponseWriter, req *http.Request, statusCode int, respBody []byte) {
	w.Header().Set("Content-Type", "application/json")
	if respBody != nil {
		w.Write(respBody)
	}

	w.WriteHeader(statusCode)
}
