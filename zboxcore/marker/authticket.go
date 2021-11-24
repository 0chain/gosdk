package marker

import (
	"fmt"

	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/zboxcore/client"
)

type AuthTicket struct {
	ClientID        string `json:"client_id"`
	OwnerID         string `json:"owner_id"`
	AllocationID    string `json:"allocation_id"`
	FilePathHash    string `json:"file_path_hash"`
	ActualFileHash  string `json:"actual_file_hash"`
	FileName        string `json:"file_name"`
	RefType         string `json:"reference_type"`
	Expiration      int64  `json:"expiration"`
	Available       int64  `json:"available_after"`
	Timestamp       int64  `json:"timestamp"`
	ReEncryptionKey string `json:"re_encryption_key,omitempty"`
	Encrypted       bool   `json:"encrypted"`
	Signature       string `json:"signature"`
}

func (rm *AuthTicket) GetHashData() string {
	hashData := fmt.Sprintf("%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v", rm.AllocationID, rm.ClientID, rm.OwnerID, rm.FilePathHash, rm.FileName, rm.RefType, rm.ReEncryptionKey, rm.Expiration, rm.Available, rm.Timestamp, rm.ActualFileHash, rm.Encrypted)
	return hashData
}

func (rm *AuthTicket) Sign() error {
	var err error
	hash := encryption.Hash(rm.GetHashData())
	rm.Signature, err = client.Sign(hash)
	return err
}
