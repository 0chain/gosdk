package blobber

import (
	"encoding/json"
	"net/http"
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
