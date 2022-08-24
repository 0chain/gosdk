package marker

import (
	"fmt"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/sys"
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

// ValidateWithOtherRM will validate rm1 assuming rm is valid. It checks parameters equality and validity of signature
func (rm *ReadMarker) ValidateWithOtherRM(rm1 *ReadMarker) error {
	if rm.ClientPublicKey != rm1.ClientPublicKey {
		return errors.New("validate_rm", fmt.Sprintf("client public key %s is not same as %s", rm1.ClientPublicKey, rm.ClientPublicKey))
	}

	signatureHash := rm1.GetHash()

	signOK, err := sys.Verify(rm1.Signature, signatureHash)
	if err != nil {
		return errors.New("validate_rm", err.Error())
	}

	if !signOK {
		return errors.New("validate_rm", "signature is not valid")
	}
	return nil
}
