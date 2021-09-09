package sdk

import (
	"github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/util"
)

type Hasher interface {
	GetFileHash() (string, error)
	// WriteToFile write bytes to file hasher
	WriteToFile(buf []byte, chunkIndex int64) error
}

type hasher struct {
	File *util.CompactMerkleTree
}

func createHasher() Hasher {
	h := &hasher{
		File: util.NewCompactMerkleTree(nil),
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

	return h.File.GetMerkleRoot(), nil
}

// WriteToFile write bytes to file hasher
func (h *hasher) WriteToFile(buf []byte, chunkIndex int64) error {
	if h == nil {
		return errors.Throw(constants.ErrInvalidParameter, "h")
	}

	if h.File == nil {
		return errors.Throw(constants.ErrInvalidParameter, "h.File")
	}

	return h.File.AddDataBlocks(buf, int(chunkIndex))
}
