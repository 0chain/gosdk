package fileref

import (
	"strconv"
	"strings"

	"github.com/0chain/gosdk/core/encryption"
)

// Hashnode ref node in hash tree
type Hashnode struct {
	// hash data
	AllocationID    string `json:"allocation_id,omitempty"`
	Type            string `json:"type,omitempty"`
	Name            string `json:"name,omitempty"`
	Path            string `json:"path,omitempty"`
	ValidationRoot  string `json:"validation_root,omitempty"`
	FixedMerkleRoot string `json:"fixed_merkle_root,omitempty"`
	ActualFileHash  string `json:"actual_file_hash,omitempty"`
	ChunkSize       int64  `json:"chunk_size,omitempty"`
	Size            int64  `json:"size,omitempty"`
	ActualFileSize  int64  `json:"actual_file_size,omitempty"`

	Children   []*Hashnode `json:"children,omitempty"`
	lookupHash string      `json:"-"`
}

func (n *Hashnode) AddChild(c *Hashnode) {
	if n.Children == nil {
		n.Children = make([]*Hashnode, 0, 10)
	}

	n.Children = append(n.Children, c)
}

// GetLookupHash get lookuphash
func (n *Hashnode) GetLookupHash() string {
	if n.lookupHash == "" {
		n.lookupHash = encryption.Hash(n.AllocationID + ":" + n.Path)
	}
	return n.lookupHash
}

// GetHashCode get hash code
func (n *Hashnode) GetHashCode() string {
	// dir
	if n.Type == DIRECTORY {
		if len(n.Children) == 0 {
			return ""
		}

		childHashes := make([]string, len(n.Children))

		var size int64

		for i, child := range n.Children {
			childHashes[i] = child.GetHashCode()
			size += child.Size
		}

		n.Size = size

		return encryption.Hash(strings.Join(childHashes, ":"))

	}

	hashArray := []string{
		n.AllocationID,
		n.Type,
		n.Name,
		n.Path,
		strconv.FormatInt(n.Size, 10),
		n.ValidationRoot,
		n.FixedMerkleRoot,
		strconv.FormatInt(n.ActualFileSize, 10),
		n.ActualFileHash,
		strconv.FormatInt(n.ChunkSize, 10),
	}

	return encryption.Hash(strings.Join(hashArray, ":"))

}
