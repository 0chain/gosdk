package sdk

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"

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
	blobbers          []*blockchain.StorageNode
	remotefilepath    string
	remotefilename    string
	authToken         *marker.AuthTicket
	refType           string
	ctx               context.Context
	expirationSeconds int64
}

func (req *ShareRequest) GetAuthTicketForEncryptedFile(clientID string, encPublicKey string) (string, error) {
	at := &marker.AuthTicket{}
	at.AllocationID = req.allocationID
	at.OwnerID = client.GetClientID()
	at.ClientID = clientID
	at.FileName = req.remotefilename
	at.FilePathHash = fileref.GetReferenceLookup(req.allocationID, req.remotefilepath)
	at.RefType = req.refType
	timestamp := int64(common.Now())

	fileRef, err := req.GetFileRef()
	if err != nil {
		return "", err
	}
	at.ActualFileHash = fileRef.ActualFileHash

	if req.expirationSeconds == 0 {
		// default expiration after 90 days
		at.Expiration = 0
	} else {
		at.Expiration = timestamp + req.expirationSeconds
	}
	at.Timestamp = timestamp
	at.Encrypted = true
	err = at.Sign()
	if err != nil {
		return "", err
	}
	if len(encPublicKey) > 0 {
		encscheme := encryption.NewEncryptionScheme()
		encscheme.Initialize(client.GetClient().Mnemonic)
		reKey, err := encscheme.GetReGenKey(encPublicKey, "filetype:audio")
		if err != nil {
			return "", err
		}
		at.ReEncryptionKey = reKey
	}
	err = at.Sign()
	if err != nil {
		return "", err
	}
	atBytes, err := json.Marshal(at)
	if err != nil {
		return "", err
	}
	sEnc := b64.StdEncoding.EncodeToString(atBytes)
	return sEnc, nil
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
	}
	_, fileRef, _ = listReq.getFileConsensusFromBlobbers()
	if fileRef == nil {
		return nil, errors.New("file_meta_error", "Error getting object meta data from blobbers")
	}
	return fileRef, nil
}

func (req *ShareRequest) GetAuthTicket(clientID string) (string, error) {
	at := &marker.AuthTicket{}
	at.AllocationID = req.allocationID
	at.OwnerID = client.GetClientID()
	at.ClientID = clientID
	at.FileName = req.remotefilename
	at.FilePathHash = fileref.GetReferenceLookup(req.allocationID, req.remotefilepath)
	at.RefType = req.refType

	fileRef, err := req.GetFileRef()
	if err != nil {
		return "", err
	}

	at.ActualFileHash = fileRef.ActualFileHash

	timestamp := int64(common.Now())
	if req.expirationSeconds == 0 {
		// default expiration after 90 days
		at.Expiration = 0
	} else {
		at.Expiration = timestamp + req.expirationSeconds
	}
	at.Timestamp = timestamp
	err = at.Sign()
	if err != nil {
		return "", err
	}
	atBytes, err := json.Marshal(at)
	if err != nil {
		return "", err
	}
	sEnc := b64.StdEncoding.EncodeToString(atBytes)
	return sEnc, nil
}
