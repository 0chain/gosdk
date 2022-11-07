package marker

import (
	"fmt"
	"strconv"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/client"
)

type FileOp int

const (
	Upload FileOp = iota
	Update
	Rename
	Copy
	Move
	Delete
	NewDir
)

type InodeMeta struct {
	MetaData    map[string]int64 `json:"meta_data"`
	LatestInode Inode            `json:"latest_inode"`
}

// Owner needs to send file id while commiting new file/dir.
// Only the latest file id is required to be signed. For example, if there is upload
// request with remotepath `/a/b/c.txt` and /a does not exist then /a, /a/b and /a/b/c.txt
// all will get file id with file id for /a/b/c.txt being the greatest.
// If before uploading latestFileID was n then /a, /a//usr/share/code/resources/app/out/vs/code/electron-sandbox/workbench/workbench.htmlb and /a/b/c.txt will have n+1,
// n+2 and n+3 respectively.
// Onwer is required to sign only n+3.
type Inode struct {
	// LatestFileID is the latest number used to provide ID for a file/dir.
	// Once assigned file id will not be used with other files. This field will
	// remain immutable for a file. Every other meta data of a file is mutable.
	LatestFileID int64 `json:"latest_file_id"`
	// OwnerSignature is the signature of LatestFileID signed by the owner.
	OwnerSignature string `json:"owner_signature"`
}

func (in *Inode) Sign() error {
	if in.LatestFileID <= 0 {
		return fmt.Errorf("invalid latest file id")
	}
	hash := encryption.Hash(strconv.FormatInt(in.LatestFileID, 10))
	sign, err := client.Sign(hash)
	if err != nil {
		return err
	}
	in.OwnerSignature = sign
	return nil
}

func (in Inode) VerifySignature() error {
	hash := encryption.Hash(strconv.FormatInt(in.LatestFileID, 10))
	isValid, err := sys.Verify(in.OwnerSignature, hash)

	if err != nil {
		return err
	}

	if !isValid {
		return fmt.Errorf("invalid_signature")
	}
	return nil
}

// GetInode Verify signature and return latest inode
func (in Inode) GetLatestInode() (int64, error) {
	err := in.VerifySignature()
	if err != nil {
		return 0, err
	}
	return in.LatestFileID, nil
}

type WriteMarker struct {
	AllocationRoot         string `json:"allocation_root"`
	PreviousAllocationRoot string `json:"prev_allocation_root"`
	FileMetaRoot           string `json:"file_meta_root"`
	AllocationID           string `json:"allocation_id"`
	Size                   int64  `json:"size"`
	BlobberID              string `json:"blobber_id"`
	Timestamp              int64  `json:"timestamp"`
	ClientID               string `json:"client_id"`
	Signature              string `json:"signature"`

	// file info
	FileID    int64  `json:"file_id"`
	Operation FileOp `json:"operation"`
}

func (wm *WriteMarker) GetHashData() string {
	sigData := fmt.Sprintf(
		"%s:%s:%s:%s:%s:%s:%d:%d:%d:%d",
		wm.AllocationRoot, wm.PreviousAllocationRoot,
		wm.FileMetaRoot, wm.AllocationID,
		wm.BlobberID, wm.ClientID, wm.Size,
		wm.Timestamp, wm.FileID, wm.Operation)
	return sigData
}

func (wm *WriteMarker) GetHash() string {
	sigData := wm.GetHashData()
	return encryption.Hash(sigData)
}

func (wm *WriteMarker) Sign() error {
	var err error
	wm.Signature, err = client.Sign(wm.GetHash())
	return err
}

func (wm *WriteMarker) VerifySignature(clientPublicKey string) error {
	hashData := wm.GetHashData()
	signatureHash := encryption.Hash(hashData)
	sigOK, err := client.VerifySignature(wm.Signature, signatureHash)
	if err != nil {
		return errors.New("write_marker_validation_failed", "Error during verifying signature. "+err.Error())
	}
	if !sigOK {
		return errors.New("write_marker_validation_failed", "Write marker signature is not valid")
	}
	return nil
}
