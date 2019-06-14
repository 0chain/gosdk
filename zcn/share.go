package zcn

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"path"

	"github.com/0chain/gosdk/encryption"
	"github.com/0chain/gosdk/util"
)

type authTicket struct {
	ClientID     string `json:"client_id"`
	OwnerID      string `json:"owner_id"`
	AllocationID string `json:"allocation_id"`
	FilePathHash string `json:"file_path_hash"`
	FileName     string `json:"file_name"`
	Expiration   int64  `json:"expiration"`
	Timestamp    int64  `json:"timestamp"`
	Signature    string `json:"signature"`
}

func (rm *authTicket) GetHashData() string {
	hashData := fmt.Sprintf("%v:%v:%v:%v:%v:%v:%v", rm.AllocationID, rm.ClientID, rm.OwnerID, rm.FilePathHash, rm.FileName, rm.Expiration, rm.Timestamp)
	return hashData
}

func (rm *authTicket) Sign(privateKey string) error {
	var err error
	hash := encryption.Hash(rm.GetHashData())
	rm.Signature, err = encryption.Sign(privateKey, hash)
	return err
}

func (obj *Allocation) GetShareAuthToken(remotePath string, clientID string) string {
	_, filename := path.Split(remotePath)
	at := &authTicket{}
	at.AllocationID = obj.allocationId
	at.OwnerID = obj.client.Id
	at.ClientID = clientID
	at.FileName = filename
	at.FilePathHash = encryption.Hash(obj.allocationId + ":" + remotePath)
	timestamp := util.Now()
	at.Expiration = timestamp + 7776000
	at.Timestamp = timestamp
	err := at.Sign(obj.client.PrivateKey)
	if err != nil {
		fmt.Println("Signing authticket failed", err)
		return ""
	}
	atBytes, err := json.Marshal(at)
	if err != nil {
		fmt.Println("Marshalling authticket failed", err)
		return ""
	}
	sEnc := b64.StdEncoding.EncodeToString(atBytes)
	return sEnc
}
