package blobber

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func uploadAndUpdateFile(w http.ResponseWriter, req *http.Request) {
	uploadMeta := req.FormValue("uploadMeta")

	var form *UploadFormData
	err := json.Unmarshal([]byte(uploadMeta), &form)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(&UploadResult{
		Filename:   form.Filename,
		Hash:       form.ChunkHash,
		MerkleRoot: form.MerkleRoot,
	})

}

func getReference(w http.ResponseWriter, req *http.Request) {

	var vars = mux.Vars(req)

	rootRefs := `{"meta_data":{"chunk_size":0,"created_at":"0001-01-01T00:00:00Z","hash":"","lookup_hash":"","name":"/","num_of_blocks":0,"path":"/","path_hash":"","size":0,"type":"d","updated_at":"0001-01-01T00:00:00Z"},"Ref":{"ID":0,"Type":"d","AllocationID":"` + vars["allocation"] + `","LookupHash":"","Name":"/","Path":"/","Hash":"","NumBlocks":0,"PathHash":"","ParentPath":"","PathLevel":1,"CustomMeta":"","ContentHash":"","Size":0,"MerkleRoot":"","ActualFileSize":0,"ActualFileHash":"","MimeType":"","WriteMarker":"","ThumbnailSize":0,"ThumbnailHash":"","ActualThumbnailSize":0,"ActualThumbnailHash":"","EncryptedKey":"","Attributes":null,"Children":null,"OnCloud":false,"CommitMetaTxns":null,"CreatedAt":"0001-01-01T00:00:00Z","UpdatedAt":"0001-01-01T00:00:00Z","DeletedAt":"0001-01-01T00:00:00Z","ChunkSize":0},"latest_write_marker":null}`

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	//nolint
	w.Write([]byte(rootRefs))
}

func commitWrite(w http.ResponseWriter, req *http.Request) {

	//	var vars = mux.Vars(req)

	writeMarker := &WriteMarker{}
	err := json.Unmarshal([]byte(req.FormValue("write_marker")), writeMarker)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := &CommitResult{}
	result.AllocationRoot = writeMarker.AllocationRoot
	result.Success = true
	result.WriteMarker = writeMarker

	json.NewEncoder(w).Encode(result)
}
