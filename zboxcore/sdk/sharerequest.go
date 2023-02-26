package sdk

import (
	"context"

	"github.com/0chain/errors"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/encryption"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/marker"
)

type ShareRequest struct {
	allocationID      string
	allocationTx      string
	remotefilepath    string
	remotefilename    string
	refType           string
	expirationSeconds int64
	blobbers          []*blockchain.StorageNode
	ctx               context.Context
}

type ListShareResp struct {
	BlobberID                 int    `json:"blobber_id"`
	OwnerID                   string `json:"owner_id"`
	ClientID                  string `json:"client_id"`
	FilePathHash              string `json:"file_path_hash"`
	ReEncryptionKey           string `json:"re_encryption_key"`
	ClientEncryptionPublicKey string `json:"client_encryption_public_key"`
	Revoked                   bool   `json:"revoked"`
	ExpiryAt                  string `json:"expiry_at"`
	AvailableAt               string `json:"available_at"`
}

type ListShareOut struct {
	Shares    []ListShareResp `json:"shares"`
	BlobberID string          `json:"blobber_id"`
}

func (req *ShareRequest) GetFileRef() (*fileref.FileRef, error) {
	filePathHash := fileref.GetReferenceLookup(req.allocationID, req.remotefilepath)

	var fileRef *fileref.FileRef
	listReq := &ListRequest{
		remotefilepathhash: filePathHash,
		allocationID:       req.allocationID,
		allocationTx:       req.allocationTx,
		blobbers:           req.blobbers,
		ctx:                req.ctx,
		Consensus:          Consensus{},
	}
	_, fileRef, _ = listReq.getFileConsensusFromBlobbers()
	if fileRef == nil {
		return nil, errors.New("file_meta_error", "Error getting object meta data from blobbers")
	}
	return fileRef, nil
}

func (req *ShareRequest) getAuthTicket(clientID, encPublicKey string) (*marker.AuthTicket, error) {
	fRef, err := req.GetFileRef()
	if err != nil {
		return nil, err
	}

	at := &marker.AuthTicket{
		AllocationID:   req.allocationID,
		OwnerID:        client.GetClientID(),
		ClientID:       clientID,
		FileName:       req.remotefilename,
		FilePathHash:   fileref.GetReferenceLookup(req.allocationID, req.remotefilepath),
		RefType:        req.refType,
		ActualFileHash: fRef.ActualFileHash,
	}

	at.Timestamp = int64(common.Now())

	if req.expirationSeconds > 0 {
		at.Expiration = at.Timestamp + req.expirationSeconds
	}

	if encPublicKey != "" { // file is encrypted
		encScheme := encryption.NewEncryptionScheme()
		if _, err := encScheme.Initialize((client.GetClient().Mnemonic)); err != nil {
			return nil, err
		}

		reKey, err := encScheme.GetReGenKey(encPublicKey, "filetype:audio")
		if err != nil {
			return nil, err
		}

		at.ReEncryptionKey = reKey
		at.Encrypted = true
	}

	if err := at.Sign(); err != nil {
		return nil, err
	}

	return at, nil
}
