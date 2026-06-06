package marker

import (
	"fmt"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/client"
)

type VersionMarker struct {
	ClientID      string `json:"client_id"`
	BlobberID     string `json:"blobber_id"`
	AllocationID  string `json:"allocation_id"`
	Version       int64  `json:"version"`
	Timestamp     int64  `json:"timestamp"`
	Signature     string `json:"signature"`
	IsRepair      bool   `json:"is_repair"`
	RepairVersion int64  `json:"repair_version"`
	RepairOffset  string `json:"repair_offset"`
}

func (vm *VersionMarker) GetHashData() string {
	return fmt.Sprintf("%s:%s:%s:%d:%d", vm.AllocationID, vm.ClientID, vm.BlobberID, vm.Version, vm.Timestamp)
}

func (vm *VersionMarker) GetHash() string {
	sigData := vm.GetHashData()
	return encryption.Hash(sigData)
}

func (vm *VersionMarker) Sign() error {
	var err error
	vm.Signature, err = client.Sign(vm.GetHash())
	return err
}

func (vm *VersionMarker) VerifySignature(clientPublicKey string) error {
	hashData := vm.GetHashData()
	signatureHash := encryption.Hash(hashData)
	// Verify against the EXPLICIT owner public key (passed by the caller via
	// client.GetClientPublicKey()), not the SDK's default/local verify key.
	// sys.Verify uses the locally-configured key, which for split-key (and
	// shared) wallets is not the allocation owner's aggregated key — so a valid
	// version marker fails verification → false "Allocation Broken". The
	// standard WriteMarker.VerifySignature already uses sys.VerifyWith.
	sigOK, err := sys.VerifyWith(clientPublicKey, vm.Signature, signatureHash)
	if err != nil {
		return errors.New("write_marker_validation_failed", "Error during verifying signature. "+err.Error())
	}
	if !sigOK {
		return errors.New("write_marker_validation_failed", "Write marker signature is not valid")
	}
	return nil
}
