package sdk

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"hash"
	"os"
	"sync"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/util"
	"github.com/cespare/xxhash/v2"
)

// GOSDK_SKIP_ACTUAL_FILE_HASH=1 replaces the whole-file ActualHash md5 with
// a random per-upload value. The File hasher md5s every payload byte on top
// of the per-shard BlockHasher; on a 64-core gateway doing bulk uploads md5
// was 49% of ALL cycles (perf 2026-06-12). Blobbers cannot verify ActualHash
// (each only holds its shard) — only DataHash (BlockHasher) is recomputed
// and enforced blobber-side (filestore/storage.go hash_mismatch), so that
// one always stays real. The random value is constant for the upload, so
// every blobber stores the same ActualHash and cross-blobber consistency
// comparisons still match. Trade-off: the stored content-MD5 is no longer
// the real file md5 (zs3 surfaces it as the object ETag — pair with
// ZS3_SKIP_MD5_ETAG semantics). Default off.
var skipActualFileHash = os.Getenv("GOSDK_SKIP_ACTUAL_FILE_HASH") == "1"

type nopMD5 struct{ sum [md5.Size]byte }

func (n *nopMD5) Write(p []byte) (int, error) { return len(p), nil }
func (n *nopMD5) Sum(b []byte) []byte         { return append(b, n.sum[:]...) }
func (n *nopMD5) Reset()                      {}
func (n *nopMD5) Size() int                   { return md5.Size }
func (n *nopMD5) BlockSize() int              { return md5.BlockSize }

func newActualFileHasher() hash.Hash {
	if skipActualFileHash {
		h := &nopMD5{}
		_, _ = rand.Read(h.sum[:])
		return h
	}
	return md5.New()
}

// GOSDK_SKIP_DATA_HASH=1 replaces the per-shard DataHash md5 (BlockHasher)
// with an opaque random value. UNLIKE ActualHash, blobbers DO recompute and
// enforce DataHash at commit — this flag is only valid against blobbers
// running with BLOBBER_SKIP_DATA_HASH=1, otherwise every upload fails with
// hash_mismatch. The shard md5 was the last full-payload hash on the upload
// path (12-19% of gateway cycles). Default off.
var skipDataHash = os.Getenv("GOSDK_SKIP_DATA_HASH") == "1"

// GOSDK_DATA_HASH selects the per-shard DataHash algorithm: "md5" (default,
// legacy) or "xxh3"/"xxhash" (xxHash64, SIMD, ~20x md5). The DataHash is a
// NON-adversarial content/corruption check — the gateway computes it, the
// blobber recomputes + enforces it at commit. The adversarial integrity (the
// challenge proofs = validation_root + fixed_merkle_root, and the BLS
// WriteMarker signature) is computed separately and is NOT touched here, so a
// fast non-cryptographic hash is safe for DataHash. The blobber must run the
// matching BLOBBER_DATA_HASH; the per-file hash-type is carried so commit-time
// re-verify and challenge re-hash pick the right algorithm (see dataHashName).
var dataHashAlgo = os.Getenv("GOSDK_DATA_HASH")

// dataHashName is the on-the-wire/stored algo tag for the active DataHash, used
// to stamp each file so the blobber re-verifies with the matching algorithm.
func dataHashName() string {
	switch dataHashAlgo {
	case "xxh3", "xxhash", "xxh":
		return "xxh3"
	}
	return "md5"
}

func newBlockHasher() hash.Hash {
	if skipDataHash {
		h := &nopMD5{}
		_, _ = rand.Read(h.sum[:])
		return h
	}
	switch dataHashAlgo {
	case "xxh3", "xxhash", "xxh":
		return xxhash.New()
	}
	return md5.New()
}

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
		File:        newActualFileHasher(),
		BlockHasher: newBlockHasher(),
	}
}

func CreateFileHasher() Hasher {
	return &hasher{
		File: newActualFileHasher(),
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
