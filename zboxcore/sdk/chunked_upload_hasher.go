package sdk

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
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
	// WriteToContent write bytes to content hasher
	WriteToContent(hash []byte, chunkIndex int) error
}

// see more detail about hash on  https://github.com/0chain/blobber/wiki/Protocols#file-hash
type hasher struct {
	File      hash.Hash             `json:"-"`
	Challenge *util.FixedMerkleTree `json:"challenge"`
	Content   hash.Hash             `json:"-"`
}

// CreateHasher creat Hasher instance
func CreateHasher(chunkSize int) Hasher {
	h := &hasher{
		File:      sha256.New(),
		Challenge: &util.FixedMerkleTree{ChunkSize: chunkSize},
		Content:   sha256.New(),
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

	return hex.EncodeToString(h.Content.Sum(nil)), nil
}

// WriteToContent write bytes to content hasher
func (h *hasher) WriteToContent(buf []byte, chunkIndex int) error {
	if h == nil {
		return errors.Throw(constants.ErrInvalidParameter, "h")
	}

	if h.Content == nil {
		return errors.Throw(constants.ErrInvalidParameter, "h.Content")
	}

	_, err := h.Content.Write(buf)

	return err
}
