package sdk

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"sync"

	"github.com/0chain/errors"

	"github.com/0chain/gosdk/core/client"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/encryption"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/marker"
)

type ShareRequest struct {
	ClientId          string
	allocationID      string
	allocationTx      string
	sig               string
	remotefilepath    string
	remotefilename    string
	refType           string
	expirationSeconds int64
	blobbers          []*blockchain.StorageNode
	ctx               context.Context
	signingPrivateKey ed25519.PrivateKey
}

func (req *ShareRequest) GetFileRef() (*fileref.FileRef, error) {
	filePathHash := fileref.GetReferenceLookup(req.allocationID, req.remotefilepath)

	var fileRef *fileref.FileRef
	listReq := &ListRequest{
		ClientId:           req.ClientId,
		remotefilepathhash: filePathHash,
		allocationID:       req.allocationID,
		allocationTx:       req.allocationTx,
		sig:                req.sig,
		blobbers:           req.blobbers,
		ctx:                req.ctx,
		Consensus:          Consensus{RWMutex: &sync.RWMutex{}},
	}
	_, _, fileRef, _ = listReq.getFileConsensusFromBlobbers()
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
		OwnerID:        client.Id(req.ClientId),
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
		var entropy string
		if fRef.EncryptionVersion == SignatureV2 {
			if len(req.signingPrivateKey) == 0 {
				return nil, errors.New("wallet_error", "signing private key is empty")
			}
			entropy = hex.EncodeToString(req.signingPrivateKey)
		} else {
			entropy = client.Wallet().Mnemonic
		}
		if entropy == "" {
			return nil, errors.New("wallet_error", "wallet mnemonic is empty")
		}
		if _, err := encScheme.Initialize((entropy)); err != nil {
			return nil, err
		}

		reKey, err := encScheme.GetReGenKey(encPublicKey, "filetype:audio")
		if err != nil {
			return nil, err
		}

		at.ReEncryptionKey = reKey
		at.Encrypted = true
		at.EncryptionPublicKey = encPublicKey
	}

	if err := at.Sign(); err != nil {
		return nil, err
	}

	return at, nil
}
