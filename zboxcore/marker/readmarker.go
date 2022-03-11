package marker

import (
	"fmt"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/zboxcore/client"
)

type ReadMarker struct {
	ClientID        string           `json:"client_id"`
	ClientPublicKey string           `json:"client_public_key"`
	BlobberID       string           `json:"blobber_id"`
	AllocationID    string           `json:"allocation_id"`
	OwnerID         string           `json:"owner_id"`
	Timestamp       common.Timestamp `json:"timestamp"`
	ReadCounter     int64            `json:"counter"`
	Signature       string           `json:"signature"`
}

func (rm *ReadMarker) GetHash() string {
	sigData := fmt.Sprintf("%v:%v:%v:%v:%v:%v:%v", rm.AllocationID,
		rm.BlobberID, rm.ClientID, rm.ClientPublicKey, rm.OwnerID,
		rm.ReadCounter, rm.Timestamp)
	return encryption.Hash(sigData)
}

func (rm *ReadMarker) Sign() error {
	var err error
	rm.Signature, err = client.Sign(rm.GetHash())
	return err
}
