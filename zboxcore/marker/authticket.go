package marker

import (
	"fmt"

	"0chain.net/clientsdk/core/encryption"
	"0chain.net/clientsdk/zboxcore/client"
)

type AuthTicket struct {
	ClientID     string `json:"client_id"`
	OwnerID      string `json:"owner_id"`
	AllocationID string `json:"allocation_id"`
	FilePathHash string `json:"file_path_hash"`
	FileName     string `json:"file_name"`
	Expiration   int64  `json:"expiration"`
	Timestamp    int64  `json:"timestamp"`
	Signature    string `json:"signature"`
}

func (rm *AuthTicket) GetHashData() string {
	hashData := fmt.Sprintf("%v:%v:%v:%v:%v:%v:%v", rm.AllocationID, rm.ClientID, rm.OwnerID, rm.FilePathHash, rm.FileName, rm.Expiration, rm.Timestamp)
	return hashData
}

func (rm *AuthTicket) Sign() error {
	var err error
	hash := encryption.Hash(rm.GetHashData())
	rm.Signature, err = client.Sign(hash)
	return err
}
