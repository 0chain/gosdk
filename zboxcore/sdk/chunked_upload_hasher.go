package sdk

import (
	"crypto/md5"
	"encoding/hex"
	"hash"
	"sync"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/util"
)

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

// ---------------------------------------------------------------------------
// HashVersion 2 — segmented (tree) hasher.
//
// The v1 hashes (ActualHash = streaming MD5 of the file, DataHash = streaming
// MD5 of the shard) require bytes IN ORDER, which forces the upload producer
// to be a single goroutine per file (see the May-29 notes in chunked_upload.go
// and zs3 docs/PARALLEL_PRODUCER_TREE_HASH_DESIGN.md). Version 2 replaces the
// streaming MD5 with a 2-level construction:
//
//	leaf_i = md5(segment_i bytes)                    — parallel, any order
//	root   = md5(leaf_0 || leaf_1 || ... || leaf_n)  — 16-byte digests folded
//	                                                    serially in index order
//
// Segment boundaries:
//   - per-blobber DataHash: one erasure fragment per leaf (ChunkSize bytes,
//     last fragment may be shorter) — exactly the byte ranges the blobber
//     writes at chunkIndex*ChunkSize, so the blobber re-derives the same
//     root from its temp file with a chunk-size splitter (see eblobber
//     filestore.CommitHasher hash_version=2).
//   - file ActualHash: one full chunk read per leaf (ChunkSize*DataShards
//     bytes). Client-attested only (signature), never recomputed by blobbers.
//
// Leaf digests are computed in parallel by the producer workers; the serial
// round loop folds them in chunk order, so the root needs no leaf storage.
type TreeHasher struct {
	mu   sync.Mutex
	root hash.Hash
}

// CreateTreeHasher creates a HashVersion-2 hasher.
func CreateTreeHasher() *TreeHasher {
	return &TreeHasher{root: md5.New()}
}

// FoldLeafDigest folds one 16-byte md5 leaf digest into the root. Calls MUST
// be made in leaf (chunk index) order — the serial round loop guarantees it.
func (h *TreeHasher) FoldLeafDigest(digest []byte) {
	h.mu.Lock()
	h.root.Write(digest) //nolint:errcheck // md5 Write never fails
	h.mu.Unlock()
}

func (h *TreeHasher) GetFileHash() (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return hex.EncodeToString(h.root.Sum(nil)), nil
}

func (h *TreeHasher) GetBlockHash() (string, error) {
	return h.GetFileHash()
}

// WriteToFile / WriteToBlockHasher are v1 streaming entry points; on the v2
// path leaves are folded via FoldLeafDigest by the round loop. Erroring here
// surfaces any accidental v1-style use of a TreeHasher.
func (h *TreeHasher) WriteToFile([]byte) error {
	return errors.New("tree_hasher", "v2 hasher takes FoldLeafDigest, not WriteToFile")
}

func (h *TreeHasher) WriteToBlockHasher([]byte) error {
	return errors.New("tree_hasher", "v2 hasher takes FoldLeafDigest, not WriteToBlockHasher")
}

func (h *TreeHasher) GetFixedMerkleRoot() (string, error) {
	return "", errors.New("tree_hasher", "fixed merkle root not supported on v2 upload path")
}

func (h *TreeHasher) WriteToFixedMT([]byte) error {
	return errors.New("tree_hasher", "fixed merkle tree not supported on v2 upload path")
}

func (h *TreeHasher) GetValidationRoot() (string, error) {
	return "", errors.New("tree_hasher", "validation root not supported on v2 upload path")
}

func (h *TreeHasher) WriteToValidationMT([]byte) error {
	return errors.New("tree_hasher", "validation tree not supported on v2 upload path")
}

func (h *TreeHasher) Finalize() error { return nil }

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
