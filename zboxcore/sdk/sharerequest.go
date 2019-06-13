package sdk

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"

	"0chain.net/clientsdk/zboxcore/client"

	"0chain.net/clientsdk/core/common"

	"0chain.net/clientsdk/zboxcore/blockchain"
	"0chain.net/clientsdk/zboxcore/marker"
)

type ShareRequest struct {
	allocationID       string
	blobbers           []*blockchain.StorageNode
	remotefilepathhash string
	remotefilepath     string
	authToken          *marker.AuthTicket
	ctx                context.Context
}

func (req *ShareRequest) GetAuthTicket(clientID string) (string, error) {
	listReq := &ListRequest{remotefilepath: req.remotefilepath, allocationID: req.allocationID, blobbers: req.blobbers, ctx: req.ctx}
	_, selected, _ := listReq.getFileConsensusFromBlobbers()
	if selected == nil {
		return "", common.NewError("invalid_parameters", "Could not get file meta data from blobbers")
	}
	at := &marker.AuthTicket{}
	at.AllocationID = req.allocationID
	at.OwnerID = client.GetClientID()
	at.ClientID = clientID
	at.FileName = selected.Name
	at.FilePathHash = selected.PathHash
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
