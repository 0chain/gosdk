package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"sync"

	goError "errors"

	"github.com/0chain/errors"
)

const (
	merkleChunkSize     = 64
	MaxMerkleLeavesSize = 64 * 1024
)

type leaf struct {
	h hash.Hash
}

func (l *leaf) GetHashBytes() []byte {
	return l.h.Sum(nil)
}

func (l *leaf) GetHash() string {
	return hex.EncodeToString(l.h.Sum(nil))
}

func (l *leaf) Write(b []byte) (int, error) {
	return l.h.Write(b)
}

func getNewLeaf() *leaf {
	return &leaf{
		h: sha256.New(),
	}
}

// FixedMerkleTree A trusted mekerl tree for outsourcing attack protection. see section 1.8 on whitepager
// see detail on https://github.com/0chain/blobber/wiki/Protocols#what-is-fixedmerkletree
type FixedMerkleTree struct {
	// ChunkSize size of chunk
	ChunkSize int `json:"chunk_size,omitempty"`
	// Leaves a leaf is a CompactMerkleTree for 1/1024 shard data
	Leaves []Hashable `json:"leaves,omitempty"`

	writeLock  *sync.Mutex
	isFinal    bool
	writeCount int
	writeBytes []byte
}

func (fmt *FixedMerkleTree) Finalize() error {
	fmt.writeLock.Lock()
	if fmt.isFinal {
		return goError.New("already finalized")
	}
	fmt.isFinal = true
	fmt.writeLock.Unlock()

	return fmt.writeToLeaves(fmt.writeBytes[:fmt.writeCount])
}

// NewFixedMerkleTree create a FixedMerkleTree with specify hash method
func NewFixedMerkleTree(chunkSize int) *FixedMerkleTree {

	t := &FixedMerkleTree{
		ChunkSize:  chunkSize,
		writeBytes: make([]byte, MaxMerkleLeavesSize),
	}
	t.initLeaves()

	return t

}

func (fmt *FixedMerkleTree) initLeaves() {
	fmt.Leaves = make([]Hashable, 1024)
	for i := 0; i < 1024; i++ {
		fmt.Leaves[i] = getNewLeaf()
	}
}

func (fmt *FixedMerkleTree) writeToLeaves(b []byte) error {
	if len(b) > MaxMerkleLeavesSize {
		return goError.New("data size greater than maximum required size")
	}

	if len(b) < MaxMerkleLeavesSize && !fmt.isFinal {
		return goError.New("invalid merkle leaf write")
	}

	dataLen := len(b)
	shouldContinue := true
	leafInd := 0
	for i, j := 0, merkleChunkSize; shouldContinue; i, j = j, j+merkleChunkSize {
		if j > dataLen {
			j = dataLen
			shouldContinue = false
		}
		fmt.Leaves[leafInd].Write(b[i:j])
	}

	return nil
}

func (fmt *FixedMerkleTree) Write(b []byte, chunkIndex int) error {

	fmt.writeLock.Lock()
	defer fmt.writeLock.Unlock()
	if fmt.isFinal {
		return goError.New("cannot write. Tree is already finalized")
	}

	byteLen := int64(len(b))
	shouldContinue := true

	for i, j := int64(0), MaxMerkleLeavesSize-int64(fmt.writeCount); shouldContinue; i, j = j, j+MaxMerkleLeavesSize {
		if j > byteLen {
			j = byteLen
			shouldContinue = false
		}
		prevWriteCount := fmt.writeCount
		fmt.writeCount += int(j - i)
		copy(fmt.writeBytes[prevWriteCount:fmt.writeCount], b[i:j])
		if fmt.writeCount == MaxMerkleLeavesSize {
			err := fmt.writeToLeaves(fmt.writeBytes)
			if err != nil {
				return err
			}
			fmt.writeCount = 0
		}
	}
	return nil
}

// GetMerkleRoot get merkle tree
func (fmt *FixedMerkleTree) GetMerkleTree() MerkleTreeI {
	merkleLeaves := make([]Hashable, 1024)

	for idx, leaf := range fmt.Leaves {

		merkleLeaves[idx] = leaf
	}
	var mt MerkleTreeI = &MerkleTree{}

	mt.ComputeTree(merkleLeaves)
	return mt
}

// GetMerkleRoot get merkle root
func (fmt *FixedMerkleTree) GetMerkleRoot() string {
	return fmt.GetMerkleTree().GetRoot()
}

// Reload reset and reload leaves from io.Reader
func (fmt *FixedMerkleTree) Reload(reader io.Reader) error {

	fmt.initLeaves()

	bytesBuf := bytes.NewBuffer(make([]byte, 0, fmt.ChunkSize))
	for i := 0; ; i++ {
		written, err := io.CopyN(bytesBuf, reader, int64(fmt.ChunkSize))

		if written > 0 {
			err = fmt.Write(bytesBuf.Bytes(), i)
			bytesBuf.Reset()

			if err != nil {
				return err
			}

		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return err
		}

	}

	return nil
}
