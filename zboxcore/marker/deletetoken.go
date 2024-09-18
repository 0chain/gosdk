package marker

import (
	"fmt"

	"github.com/0chain/gosdk/core/client"
	"github.com/0chain/gosdk/core/encryption"
)

type DeleteToken struct {
	FilePathHash string `json:"file_path_hash"`
	FileRefHash  string `json:"file_ref_hash"`
	AllocationID string `json:"allocation_id"`
	Size         int64  `json:"size"`
	BlobberID    string `json:"blobber_id"`
	Timestamp    int64  `json:"timestamp"`
	ClientID     string `json:"client_id"`
	Signature    string `json:"signature"`
}

func (dt *DeleteToken) GetHash() string {
	sigData := fmt.Sprintf("%v:%v:%v:%v:%v:%v:%v", dt.FileRefHash, dt.FilePathHash, dt.AllocationID, dt.BlobberID, dt.ClientID, dt.Size, dt.Timestamp)
	return encryption.Hash(sigData)
}

func (dt *DeleteToken) Sign() error {
	var err error
	dt.Signature, err = client.Sign(dt.GetHash(), dt.ClientID)
	return err
}
