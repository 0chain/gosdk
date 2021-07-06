package sdk

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/common/errors"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/encryption"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/marker"
)

type ShareRequest struct {
	allocationID   string
	allocationTx   string
	blobbers       []*blockchain.StorageNode
	remotefilepath string
	remotefilename string
	authToken      *marker.AuthTicket
	refType        string
	ctx            context.Context
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
	at.Expiration = timestamp + 7776000
	at.Timestamp = timestamp
	err := at.Sign()
	if err != nil {
		return "", err
	}
	var fileRef *fileref.FileRef
	listReq := &ListRequest{
		remotefilepathhash: at.FilePathHash,
		allocationID:       req.allocationID,
		allocationTx:       req.allocationTx,
		blobbers:           req.blobbers,
		ctx:                req.ctx,
	}
	//listReq.authToken = at
	_, fileRef, _ = listReq.getFileConsensusFromBlobbers()
	if fileRef == nil {
		return "", errors.New("file_meta_error", "Error getting object meta data from blobbers")
	}
	if fileRef.Type == fileref.DIRECTORY || len(fileRef.EncryptedKey) == 0 {
		return req.GetAuthTicket(clientID)
	}
	var encscheme encryption.EncryptionScheme
	encscheme = encryption.NewEncryptionScheme()
	encscheme.Initialize(client.GetClient().Mnemonic)
	reKey, err := encscheme.GetReGenKey(encPublicKey, "filetype:audio")
	if err != nil {
		return "", err
	}
	at.ReEncryptionKey = reKey
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

func (req *ShareRequest) GetAuthTicket(clientID string) (string, error) {

	at := &marker.AuthTicket{}
	at.AllocationID = req.allocationID
	at.OwnerID = client.GetClientID()
	at.ClientID = clientID
	at.FileName = req.remotefilename
	at.FilePathHash = fileref.GetReferenceLookup(req.allocationID, req.remotefilepath)
	at.RefType = req.refType
	timestamp := int64(common.Now())
	at.Expiration = timestamp + 7776000
	at.Timestamp = timestamp
	err := at.Sign()
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
