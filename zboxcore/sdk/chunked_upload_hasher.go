package sdk

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/util"
)

type Hasher interface {

	// GetFileHash get file hash
	GetFileHash() (string, error)
	// WriteToFile write bytes to file hasher
	WriteToFile(buf []byte, chunkIndex int) error

	// GetChallengeHash get challenge hash
	GetChallengeHash() (string, error)
	// WriteToChallenge write bytes to challenge hasher
	WriteToChallenge(buf []byte, chunkIndex int) error

	// GetContentHash get content hash
	GetContentHash() (string, error)
	// WriteHashToContent write hash leaf to content hasher
	WriteHashToContent(hash string, chunkIndex int) error
}

// see more detail about hash on  https://github.com/0chain/blobber/wiki/Protocols#file-hash
type hasher struct {
	File      hash.Hash               `json:"-"`
	Challenge *util.FixedMerkleTree   `json:"challenge"`
	Content   *util.CompactMerkleTree `json:"content"`
}

// CreateHasher creat Hasher instance
func CreateHasher(chunkSize int) Hasher {
	h := &hasher{
		File:      sha256.New(),
		Challenge: &util.FixedMerkleTree{ChunkSize: chunkSize},
		Content: util.NewCompactMerkleTree(func(left, right string) string {
			return encryption.Hash(left + right)
		}),
	}

	return h
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
func (h *hasher) WriteToFile(buf []byte, chunkIndex int) error {
	if h == nil {
		return errors.Throw(constants.ErrInvalidParameter, "h")
	}

	if h.File == nil {
		return errors.Throw(constants.ErrInvalidParameter, "h.File")
	}

	_, err := h.File.Write(buf)

	return err
}

// GetChallengeHash get challenge hash
func (h *hasher) GetChallengeHash() (string, error) {
	if h == nil {
		return "", errors.Throw(constants.ErrInvalidParameter, "h")
	}

	if h.Challenge == nil {
		return "", errors.Throw(constants.ErrInvalidParameter, "h.Challenge")
	}

	return h.Challenge.GetMerkleRoot(), nil
}

// WriteToChallenge write bytes to challenge hasher
func (h *hasher) WriteToChallenge(buf []byte, chunkIndex int) error {
	if h == nil {
		return errors.Throw(constants.ErrInvalidParameter, "h")
	}

	if h.Challenge == nil {
		return errors.Throw(constants.ErrInvalidParameter, "h.Challenge")
	}

	return h.Challenge.Write(buf, chunkIndex)
}

// GetContentHash get content hash
func (h *hasher) GetContentHash() (string, error) {
	if h == nil {
		return "", errors.Throw(constants.ErrInvalidParameter, "h")
	}

	if h.Content == nil {
		return "", errors.Throw(constants.ErrInvalidParameter, "h.Content")
	}

	return h.Content.GetMerkleRoot(), nil
}

// WriteHashToContent write hash leaf to content hasher
func (h *hasher) WriteHashToContent(hash string, chunkIndex int) error {
	if h == nil {
		return errors.Throw(constants.ErrInvalidParameter, "h")
	}

	if h.Content == nil {
		return errors.Throw(constants.ErrInvalidParameter, "h.Content")
	}

	return h.Content.AddLeaf(hash, chunkIndex)
}
