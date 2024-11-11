package util

import (
	"bytes"
	"encoding/hex"
	"hash"
	"io"
	"sync"

	goError "errors"

	"github.com/0chain/errors"
	"github.com/minio/sha256-simd"
)

const (
	// MerkleChunkSize is the size of a chunk of data that is hashed
	MerkleChunkSize     = 64

	// MaxMerkleLeavesSize is the maximum size of the data that can be written to the merkle tree
	MaxMerkleLeavesSize = 64 * 1024

	// FixedMerkleLeaves is the number of leaves in the fixed merkle tree
	FixedMerkleLeaves   = 1024

	// FixedMTDepth is the depth of the fixed merkle tree
	FixedMTDepth        = 11
)

var (
	leafPool = sync.Pool{
		New: func() interface{} {
			return &leaf{
				h: sha256.New(),
			}
		},
	}
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
	l, ok := leafPool.Get().(*leaf)
	if !ok {
		return &leaf{
			h: sha256.New(),
		}
	}
	l.h.Reset()
	return l
}

// FixedMerkleTree A trusted mekerle tree for outsourcing attack protection. see section 1.8 on whitepager
// see detail on https://github.com/0chain/blobber/wiki/Protocols#what-is-fixedmerkletree
type FixedMerkleTree struct {
	// Leaves will store hash digester that calculates sha256 hash of the leaf content
	Leaves []Hashable `json:"leaves,omitempty"`

	writeLock sync.Mutex
	// isFinal is set to true once Finalize() is called.
	// After it is set to true, there will be no any writes to writeBytes field
	isFinal bool
	// writeCount will track count of bytes written to writeBytes field
	writeCount int
	// writeBytes will store bytes upto MaxMerkleLeavesSize. For the last bytes that
	// does not make upto MaxMerkleLeavesSize, it will be sliced with writeCount field.
	writeBytes []byte
	merkleRoot []byte
}

// Finalize will set isFinal to true and sends remaining bytes for leaf hash calculation
func (fmt *FixedMerkleTree) Finalize() error {
	fmt.writeLock.Lock()
	defer fmt.writeLock.Unlock()

	if fmt.isFinal {
		return goError.New("already finalized")
	}
	fmt.isFinal = true
	if fmt.writeCount > 0 {
		return fmt.writeToLeaves(fmt.writeBytes[:fmt.writeCount])
	}
	return nil
}

// NewFixedMerkleTree create a FixedMerkleTree with specify hash method
func NewFixedMerkleTree() *FixedMerkleTree {

	t := &FixedMerkleTree{
		writeBytes: make([]byte, MaxMerkleLeavesSize),
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

// writeToLeaves will divide the data with MerkleChunkSize(64 bytes) and write to
// each leaf hasher
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

		_, err := fmt.Leaves[leafInd].Write(b[i:j])
		if err != nil {
			return err
		}
		leafInd++
	}

	return nil
}

// Write will write data to the leaves once MaxMerkleLeavesSize(64 KB) is reached.
// Since each 64KB is divided into 1024 pieces with 64 bytes each, once data len reaches
// 64KB then it will be written to leaf hashes. The remaining data that is not multiple of
// 64KB will be written to leaf hashes by Finalize() function.
// This can be used to write stream of data as well.
// fmt.Finalize() is required after data write is complete.
func (fmt *FixedMerkleTree) Write(b []byte) (int, error) {

	fmt.writeLock.Lock()
	defer fmt.writeLock.Unlock()
	if fmt.isFinal {
		return 0, goError.New("cannot write. Tree is already finalized")
	}

	for i, j := 0, MaxMerkleLeavesSize-fmt.writeCount; i < len(b); i, j = j, j+MaxMerkleLeavesSize {
		if j > len(b) {
			j = len(b)
		}
		prevWriteCount := fmt.writeCount
		fmt.writeCount += int(j - i)
		copy(fmt.writeBytes[prevWriteCount:fmt.writeCount], b[i:j])

		if fmt.writeCount == MaxMerkleLeavesSize {
			// data fragment reached 64KB, so send this slice to write to leaf hashes
			err := fmt.writeToLeaves(fmt.writeBytes)
			if err != nil {
				return 0, err
			}
			fmt.writeCount = 0 // reset writeCount
		}
	}
	return len(b), nil
}

// GetMerkleRoot is only for interface compliance.
func (fmt *FixedMerkleTree) GetMerkleTree() MerkleTreeI {
	return nil
}

func (fmt *FixedMerkleTree) CalculateMerkleRoot() {
	nodes := make([][]byte, len(fmt.Leaves))
	for i := 0; i < len(nodes); i++ {
		nodes[i] = fmt.Leaves[i].GetHashBytes()
		leafPool.Put(fmt.Leaves[i])
	}

	for i := 0; i < FixedMTDepth; i++ {

		newNodes := make([][]byte, (len(nodes)+1)/2)
		nodeInd := 0
		for j := 0; j < len(nodes); j += 2 {
			newNodes[nodeInd] = MHashBytes(nodes[j], nodes[j+1])
			nodeInd++
		}
		nodes = newNodes
		if len(nodes) == 1 {
			break
		}
	}

	fmt.merkleRoot = nodes[0]
}

// FixedMerklePath is used to verify existence of leaf hash for fixed merkle tree
type FixedMerklePath struct {
	LeafHash []byte   `json:"leaf_hash"`
	RootHash []byte   `json:"root_hash"`
	Nodes    [][]byte `json:"nodes"`
	LeafInd  int
}

func (fp FixedMerklePath) VerifyMerklePath() bool {
	leafInd := fp.LeafInd
	hash := fp.LeafHash
	for i := 0; i < len(fp.Nodes); i++ {
		if leafInd&1 == 0 {
			hash = MHashBytes(hash, fp.Nodes[i])
		} else {
			hash = MHashBytes(fp.Nodes[i], hash)
		}
		leafInd = leafInd / 2
	}
	return bytes.Equal(hash, fp.RootHash)
}

// GetMerkleRoot get merkle root.
func (fmt *FixedMerkleTree) GetMerkleRoot() string {
	if fmt.merkleRoot != nil {
		return hex.EncodeToString(fmt.merkleRoot)
	}
	fmt.CalculateMerkleRoot()
	return hex.EncodeToString(fmt.merkleRoot)
}

// Reload reset and reload leaves from io.Reader
func (fmt *FixedMerkleTree) Reload(reader io.Reader) error {

	fmt.initLeaves()

	bytesBuf := bytes.NewBuffer(make([]byte, 0, MaxMerkleLeavesSize))
	for i := 0; ; i++ {
		written, err := io.CopyN(bytesBuf, reader, MaxMerkleLeavesSize)

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
