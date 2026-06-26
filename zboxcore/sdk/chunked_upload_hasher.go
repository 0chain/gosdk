package sdk

import (
	"crypto/md5"
	"encoding/hex"
	"hash"
	"os"
	"sync"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/util"
)

// skipActualFileHash, when GOSDK_SKIP_ACTUAL_FILE_HASH=1, skips the whole-file
// content MD5 (the S3 ETag / FileMeta.ActualHash). The blobber does NOT verify
// ActualHash (only the per-block DataHash is recomputed + enforced), so skipping
// it is safe and removes one of the two full-payload MD5 passes — ~half the
// upload hashing CPU (profiled at 52% of the gateway on PUT). The ETag becomes a
// fixed placeholder, so enable only where the real content-MD5 ETag isn't needed
// (FSAL streaming / bulk S3 PUT).
var skipActualFileHash = os.Getenv("GOSDK_SKIP_ACTUAL_FILE_HASH") == "1"

// Hasher interface to gather all hasher related functions.
// A hasher is used to calculate the hash of a file, fixed merkle tree, and validation merkle tree.
type Hasher interface {
	// GetFileHash get file hash
	GetFileHash() (string, error)
	// WriteToFile write bytes to file hasher
	WriteToFile(buf []byte) error

	GetFixedMerkleRoot() (string, error)
	// WriteToFixedMT write bytes to FMT hasher
	WriteToFixedMT(buf []byte) error

	GetValidationRoot() (string, error)
	// WriteToValidationMT write bytes Validation Tree hasher
	WriteToValidationMT(buf []byte) error
	// Finalize will let merkle tree know that tree is finalized with the content it has received
	Finalize() error
	// WriteToBlockHasher write bytes to block hasher
	WriteToBlockHasher(buf []byte) error
	// GetBlockHash get block hash
	GetBlockHash() (string, error)
}

// see more detail about hash on  https://github.com/0chain/blobber/wiki/Protocols#file-hash
type hasher struct {
	File         hash.Hash             `json:"-"`
	FixedMT      *util.FixedMerkleTree `json:"fixed_merkle_tree"`
	ValidationMT *util.ValidationTree  `json:"validation_merkle_tree"`
	BlockHasher  hash.Hash             `json:"-"`
}

// CreateHasher creat Hasher instance
func CreateHasher(dataSize int64) Hasher {
	return &hasher{
		File:        md5.New(),
		BlockHasher: md5.New(),
	}
}

func CreateFileHasher() Hasher {
	return &hasher{
		File: md5.New(),
	}
}

func (h *hasher) GetFileHash() (string, error) {
	// Fixed placeholder ETag when the content MD5 was skipped (see
	// skipActualFileHash). Distinct from md5-of-empty so it's recognisable.
	if skipActualFileHash {
		return "00000000000000000000000000000000", nil
	}
	if h == nil {
		return "", errors.Throw(constants.ErrInvalidParameter, "h")
	}

	if h.File == nil {
		return "", errors.Throw(constants.ErrInvalidParameter, "h.File")
	}
	return hex.EncodeToString(h.File.Sum(nil)), nil
}

// WriteToFile write bytes to file hasher
func (h *hasher) WriteToFile(buf []byte) error {
	// Skip the whole-file content MD5 entirely (no per-byte hashing) when
	// GOSDK_SKIP_ACTUAL_FILE_HASH=1 — the blobber never verifies ActualHash.
	if skipActualFileHash {
		return nil
	}
	if h == nil {
		return errors.Throw(constants.ErrInvalidParameter, "h")
	}

	if h.File == nil {
		return errors.Throw(constants.ErrInvalidParameter, "h.File")
	}

	_, err := h.File.Write(buf)
	return err
}

func (h *hasher) GetFixedMerkleRoot() (string, error) {
	if h == nil {
		return "", errors.Throw(constants.ErrInvalidParameter, "h")
	}

	if h.FixedMT == nil {
		return "", errors.Throw(constants.ErrInvalidParameter, "h.Challenge")
	}

	return h.FixedMT.GetMerkleRoot(), nil
}

func (h *hasher) WriteToFixedMT(buf []byte) error {
	if h == nil {
		return errors.Throw(constants.ErrInvalidParameter, "h")
	}

	if h.FixedMT == nil {
		return errors.Throw(constants.ErrInvalidParameter, "h.Challenge")
	}
	_, err := h.FixedMT.Write(buf)
	return err
}

func (h *hasher) GetValidationRoot() (string, error) {
	if h == nil {
		return "", errors.Throw(constants.ErrInvalidParameter, "h")
	}

	if h.ValidationMT == nil {
		return "", errors.Throw(constants.ErrInvalidParameter, "h.Content")
	}

	return hex.EncodeToString(h.ValidationMT.GetValidationRoot()), nil
}

func (h *hasher) WriteToValidationMT(buf []byte) error {
	if h == nil {
		return errors.Throw(constants.ErrInvalidParameter, "h")
	}

	if h.ValidationMT == nil {
		return errors.Throw(constants.ErrInvalidParameter, "h.Content")
	}
	_, err := h.ValidationMT.Write(buf)
	return err
}

func (h *hasher) WriteToBlockHasher(buf []byte) error {
	if h == nil {
		return errors.Throw(constants.ErrInvalidParameter, "h")
	}

	if h.BlockHasher == nil {
		return errors.Throw(constants.ErrInvalidParameter, "h.BlockHasher")
	}

	_, err := h.BlockHasher.Write(buf)
	return err
}

func (h *hasher) GetBlockHash() (string, error) {
	if h == nil {
		return "", errors.Throw(constants.ErrInvalidParameter, "h")
	}

	if h.BlockHasher == nil {
		return "", errors.Throw(constants.ErrInvalidParameter, "h.BlockHasher")
	}

	return hex.EncodeToString(h.BlockHasher.Sum(nil)), nil
}

func (h *hasher) Finalize() error {
	var (
		wg      sync.WaitGroup
		errChan = make(chan error, 2)
	)
	wg.Add(2)
	go func() {
		if err := h.FixedMT.Finalize(); err != nil {
			errChan <- err
		}
		wg.Done()
	}()
	go func() {
		if err := h.ValidationMT.Finalize(); err != nil {
			errChan <- err
		}
		wg.Done()
	}()
	wg.Wait()
	close(errChan)
	for err := range errChan {
		return err
	}
	return nil
}
