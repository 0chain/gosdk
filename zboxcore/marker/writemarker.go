package marker

import (
	"fmt"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/client"
)

type WriteMarker struct {
	AllocationRoot         string `json:"allocation_root"`
	PreviousAllocationRoot string `json:"prev_allocation_root"`
	FileMetaRoot           string `json:"file_meta_root"`
	AllocationID           string `json:"allocation_id"`
	Size                   int64  `json:"size"`
	ChainSize              int64  `json:"chain_size"`
	ChainHash              string `json:"chain_hash"`
	ChainLength            int    `json:"chain_length"`
	BlobberID              string `json:"blobber_id"`
	Timestamp              int64  `json:"timestamp"`
	ClientID               string `json:"client_id"`
	Signature              string `json:"signature"`
}

func (wm *WriteMarker) GetHashData() string {
	var sigData string
	if wm.ChainHash != "" {
		sigData = fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s:%d:%d:%d",
			wm.AllocationRoot, wm.PreviousAllocationRoot,
			wm.FileMetaRoot, wm.ChainHash, wm.AllocationID, wm.BlobberID,
			wm.ClientID, wm.Size, wm.ChainSize, wm.Timestamp)
	} else {
		sigData = fmt.Sprintf("%s:%s:%s:%s:%s:%s:%d:%d",
			wm.AllocationRoot, wm.PreviousAllocationRoot,
			wm.FileMetaRoot, wm.AllocationID,
			wm.BlobberID, wm.ClientID, wm.Size,
			wm.Timestamp)
	}
	return sigData
}

func (wm *WriteMarker) GetHash() string {
	sigData := wm.GetHashData()
	return encryption.Hash(sigData)
}

func (wm *WriteMarker) Sign() error {
	var err error
	wm.Signature, err = client.Sign(wm.GetHash())
	return err
}

func (wm *WriteMarker) VerifySignature(clientPublicKey string) error {
	hashData := wm.GetHashData()
	signatureHash := encryption.Hash(hashData)
	sigOK, err := sys.VerifyWith(clientPublicKey, wm.Signature, signatureHash)
	if err != nil {
		return errors.New("write_marker_validation_failed", "Error during verifying signature. "+err.Error())
	}
	if !sigOK {
		return errors.New("write_marker_validation_failed", "Write marker signature is not valid")
	}
	return nil
}
