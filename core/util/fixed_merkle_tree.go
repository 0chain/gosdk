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
	MerkleChunkSize     = 64
	MaxMerkleLeavesSize = 64 * 1024
	FixedMerkleLeaves   = 1024
	FixedMTDepth        = 11
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
		writeLock:  &sync.Mutex{},
	}
	t.initLeaves()

	return t

}

func (fmt *FixedMerkleTree) initLeaves() {
	fmt.Leaves = make([]Hashable, FixedMerkleLeaves)
	for i := 0; i < FixedMerkleLeaves; i++ {
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

	leafInd := 0
	for i := 0; i < len(b); i += MerkleChunkSize {
		j := i + MerkleChunkSize
		if j > len(b) {
			j = len(b)
		}

		fmt.Leaves[leafInd].Write(b[i:j])
		leafInd++
	}

	return nil
}

func (fmt *FixedMerkleTree) Write(b []byte) (int, error) {

	fmt.writeLock.Lock()
	defer fmt.writeLock.Unlock()
	if fmt.isFinal {
		return 0, goError.New("cannot write. Tree is already finalized")
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
				return 0, err
			}
			fmt.writeCount = 0
		}
	}
	return int(byteLen), nil
}

// GetMerkleRoot get merkle tree
func (fmt *FixedMerkleTree) GetMerkleTree() MerkleTreeI {
	merkleLeaves := make([]Hashable, FixedMerkleLeaves)
	copy(merkleLeaves, fmt.Leaves)
	mt := &MerkleTree{}
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
			_, err = fmt.Write(bytesBuf.Bytes())
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
