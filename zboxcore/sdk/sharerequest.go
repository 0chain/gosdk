package sdk

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"

	"github.com/0chain/gosdk/zboxcore/fileref"

	"github.com/0chain/gosdk/zboxcore/client"

	"github.com/0chain/gosdk/core/common"

	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/marker"
)

type ShareRequest struct {
	allocationID   string
	blobbers       []*blockchain.StorageNode
	remotefilepath string
	remotefilename string
	authToken      *marker.AuthTicket
	refType        string
	ctx            context.Context
}

func (req *ShareRequest) GetAuthTicket(clientID string) (string, error) {
	at := &marker.AuthTicket{}
	at.AllocationID = req.allocationID
	at.OwnerID = client.GetClientID()
	at.ClientID = clientID
	at.FileName = req.remotefilename
	at.FilePathHash = fileref.GetReferenceLookup(req.allocationID, req.remotefilepath)
	at.RefType = req.refType
	timestamp := common.Now()
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
