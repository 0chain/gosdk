package util

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"io"

	goErrors "errors"

	"github.com/0chain/errors"

	"github.com/0chain/gosdk/core/encryption"
)

var (
	// ErrLeafExists leaf has been computed, it can be skipped now
	ErrLeafExists = goErrors.New("merkle: leaf exists, it can be skipped")
	// ErrLeafNoSequenced leaf MUST be pushed one by one
	ErrLeafNoSequenced = goErrors.New("merkle: leaf must be pushed with sequence")
)

// CompactMerkleTree it is a stateful algorithm. It takes data in (leaf nodes), hashes it, and computes as many parent hashes as it can.
// see detail on https://github.com/0chain/blobber/wiki/Protocols#what-is-compactmerkletree
type CompactMerkleTree struct {
	Tree  []string                        `json:"tree"`  //node tree with computed as many parent hashes as it can
	Hash  func(left, right string) string `json:"-"`     //it should be set once CompactMerkleTree is created
	Count int                             `json:"count"` //how many leaves has been pushed
}

// NewCompactMerkleTree create a CompactMerkleTree with specify hash method
func NewCompactMerkleTree(hash func(left, right string) string) *CompactMerkleTree {

	if hash == nil {
		hash = func(left, right string) string {
			return encryption.Hash(left + right)
		}
	}
	return &CompactMerkleTree{
		Tree: make([]string, 0, 10),
		Hash: hash,
	}

}

// AddLeaf add leaf hash and update the the Merkle tree.
func (cmt *CompactMerkleTree) AddDataBlocks(buf []byte, index int) error {

	h := sha1.New()
	h.Write(buf)

	return cmt.AddLeaf(hex.EncodeToString(h.Sum(nil)), index)
}

// AddLeaf add leaf hash and update the the Merkle tree.
func (cmt *CompactMerkleTree) AddLeaf(leaf string, index int) error {

	if index < cmt.Count {
		return ErrLeafExists
	}

	if index > cmt.Count {
		return ErrLeafNoSequenced
	}

	if cmt.Hash == nil {
		cmt.Hash = func(left, right string) string {
			return encryption.Hash(left + right)
		}
	}

	rightHash := leaf

	for i, node := range cmt.Tree {
		if node == "" { // If we find an empty spot in the nodes, we put the hash there and quit.
			cmt.Tree[i] = rightHash
			cmt.Count++
			return nil
		}
		// Otherwise, hash the old hash with the new hash.
		leftHash := cmt.Tree[i]
		rightHash = cmt.Hash(leftHash, rightHash)
		// We no longer need to keep the old hash at this level in memory.
		cmt.Tree[i] = ""
	}

	if cmt.Tree == nil {
		cmt.Tree = make([]string, 0, 10)
	}

	//no valid left hash found, so make it as a new leaf hash
	cmt.Tree = append(cmt.Tree, rightHash)
	cmt.Count++
	return nil

}

// GetMerkleRoot calculate the Merkle root when all leave has been added,
// For the last, lowest-level hash, we hash it with itself.
// From there, the nodes are hashed to the top level
// to calculate the Merkle root.
func (cmt *CompactMerkleTree) GetMerkleRoot() string {
	if cmt.Hash == nil {
		cmt.Hash = func(left, right string) string {
			return encryption.Hash(left + right)
		}
	}

	rightHash := ""

	// Fill in missing nodes.
	for i := range cmt.Tree {

		leftHash := cmt.Tree[i]
		if i == len(cmt.Tree) && rightHash == "" {
			// Perfectly balanced Merkle tree.
			return leftHash
		}
		if leftHash == "" && rightHash == "" {
			// Both leaves are null (subsumed by a higher node hash)
			continue
		} else if rightHash == "" {
			// If there is no right hash (at this level or lower in the tree),
			// Hash the left hash with itself.
			rightHash = cmt.Hash(leftHash, leftHash)
		} else if leftHash == "" {
			// Similarly, if there is no left half,
			// hash the right with itself.
			rightHash = cmt.Hash(rightHash, rightHash)
		} else {
			// Otherwise, the hash at this level will be the right hash
			// for higher levels in the tree.
			rightHash = cmt.Hash(leftHash, rightHash)
		}
	}

	return rightHash
}

// Reload reset and reload leaves from io.Reader
func (cmt *CompactMerkleTree) Reload(chunkSize int64, reader io.Reader) error {
	cmt.Tree = make([]string, 0, 10)

	bytesBuf := bytes.NewBuffer(make([]byte, 0, chunkSize))
	for i := 0; ; i++ {
		written, err := io.CopyN(bytesBuf, reader, chunkSize)

		if written > 0 {
			cmt.AddDataBlocks(bytesBuf.Bytes(), i)
			bytesBuf.Reset()
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
