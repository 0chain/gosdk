package fileref

import (
	"math"
	"strconv"
	"strings"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/encryption"
)

const CHUNK_SIZE = 64 * 1024

const (
	FILE      = "f"
	DIRECTORY = "d"
)

type CommitMetaTxn struct {
	RefID     int64  `json:"ref_id"`
	TxnID     string `json:"txn_id"`
	CreatedAt string `json:"created_at"`
}

type Collaborator struct {
	RefID     int64  `json:"ref_id"`
	ClientID  string `json:"client_id"`
	CreatedAt string `json:"created_at"`
}

type FileRef struct {
	Ref                 `mapstructure:",squash"`
	CustomMeta          string          `json:"custom_meta" mapstructure:"custom_meta"`
	ContentHash         string          `json:"content_hash" mapstructure:"content_hash"`
	MerkleRoot          string          `json:"merkle_root" mapstructure:"merkle_root"`
	ThumbnailSize       int64           `json:"thumbnail_size" mapstructure:"thumbnail_size"`
	ThumbnailHash       string          `json:"thumbnail_hash" mapstructure:"thumbnail_hash"`
	ActualFileSize      int64           `json:"actual_file_size" mapstructure:"actual_file_size"`
	ActualFileHash      string          `json:"actual_file_hash" mapstructure:"actual_file_hash"`
	ActualThumbnailSize int64           `json:"actual_thumbnail_size" mapstructure:"actual_thumbnail_size"`
	ActualThumbnailHash string          `json:"actual_thumbnail_hash" mapstructure:"actual_thumbnail_hash"`
	MimeType            string          `json:"mimetype" mapstructure:"mimetype"`
	EncryptedKey        string          `json:"encrypted_key" mapstructure:"encrypted_key"`
	CommitMetaTxns      []CommitMetaTxn `json:"commit_meta_txns" mapstructure:"commit_meta_txns"`
	Collaborators       []Collaborator  `json:"collaborators" mapstructure:"collaborators"`
}

type RefEntity interface {
	GetNumBlocks() int64
	GetSize() int64
	GetHash() string
	CalculateHash() string
	GetType() string
	GetPathHash() string
	GetLookupHash() string
	GetPath() string
	GetName() string
	GetCreatedAt() common.Timestamp
	GetUpdatedAt() common.Timestamp
}

type Ref struct {
	Type             string `json:"type" mapstructure:"type"`
	AllocationID     string `json:"allocation_id" mapstructure:"allocation_id"`
	Name             string `json:"name" mapstructure:"name"`
	Path             string `json:"path" mapstructure:"path"`
	Size             int64  `json:"size" mapstructure:"size"`
	ActualSize       int64  `json:"actual_file_size" mapstructure:"actual_file_size"`
	Hash             string `json:"hash" mapstructure:"hash"`
	ChunkSize        int64  `json:"chunk_size" mapstructure:"chunk_size"`
	NumBlocks        int64  `json:"num_of_blocks" mapstructure:"num_of_blocks"`
	PathHash         string `json:"path_hash" mapstructure:"path_hash"`
	LookupHash       string `json:"lookup_hash" mapstructure:"lookup_hash"`
	HashToBeComputed bool
	ChildrenLoaded   bool
	Children         []RefEntity      `json:"-" mapstructure:"-"`
	CreatedAt        common.Timestamp `json:"created_at" mapstructure:"created_at"`
	UpdatedAt        common.Timestamp `json:"updated_at" mapstructure:"updated_at"`
}

func GetReferenceLookup(allocationID string, path string) string {
	return encryption.Hash(allocationID + ":" + path)
}

func (r *Ref) CalculateHash() string {
	if len(r.Children) == 0 && !r.ChildrenLoaded && !r.HashToBeComputed {
		return r.Hash
	}

	childHashes := make([]string, len(r.Children))
	childPaths := make([]string, len(r.Children))
	var refNumBlocks int64
	var size int64

	for index, childRef := range r.Children {
		childRef.CalculateHash()
		childHashes[index] = childRef.GetHash()
		childPaths[index] = childRef.GetPath()
		refNumBlocks += childRef.GetNumBlocks()
		size += childRef.GetSize()
	}

	r.Hash = encryption.Hash(strings.Join(childHashes, ":"))

	r.PathHash = encryption.Hash(strings.Join(childPaths, ":"))
	r.NumBlocks = refNumBlocks
	r.Size = size

	return r.Hash
}

func (r *Ref) GetHash() string {
	return r.Hash
}

func (r *Ref) GetType() string {
	return r.Type
}

func (r *Ref) GetNumBlocks() int64 {
	return r.NumBlocks
}

func (r *Ref) GetSize() int64 {
	return r.Size
}

func (r *Ref) GetPathHash() string {
	return r.PathHash
}

func (r *Ref) GetLookupHash() string {
	return r.LookupHash
}

func (r *Ref) GetPath() string {
	return r.Path
}

func (r *Ref) GetName() string {
	return r.Name
}

func (r *Ref) GetCreatedAt() common.Timestamp {
	return r.CreatedAt
}

func (r *Ref) GetUpdatedAt() common.Timestamp {
	return r.UpdatedAt
}

func (r *Ref) AddChild(child RefEntity) {
	if r.Children == nil {
		r.Children = make([]RefEntity, 0)
	}
	var index int
	var ltFound bool // less than found
	// Add child in sorted fashion
	for i, ref := range r.Children {
		if strings.Compare(child.GetPath(), ref.GetPath()) == -1 {
			index = i
			ltFound = true
			break
		}
	}
	if ltFound {
		r.Children = append(r.Children[:index+1], r.Children[index:]...)
		r.Children[index] = child
	} else {
		r.Children = append(r.Children, child)
	}
	r.ChildrenLoaded = true
}

func (r *Ref) RemoveChild(idx int) {
	if idx < 0 {
		return
	}
	r.Children = append(r.Children[:idx], r.Children[idx+1:]...)
}

func (fr *FileRef) GetHashData() string {
	hashArray := make([]string, 0, 10)
	hashArray = append(hashArray,
		fr.AllocationID,
		fr.Type,
		fr.Name,
		fr.Path,
		strconv.FormatInt(fr.Size, 10),
		fr.ContentHash,
		fr.MerkleRoot,
		strconv.FormatInt(fr.ActualFileSize, 10),
		fr.ActualFileHash,
		strconv.FormatInt(fr.ChunkSize, 10),
	)
	return strings.Join(hashArray, ":")
}

func (fr *FileRef) GetHash() string {
	return fr.Hash
}

func (fr *FileRef) CalculateHash() string {
	fr.Hash = encryption.Hash(fr.GetHashData())
	fr.NumBlocks = int64(math.Ceil(float64(fr.Size*1.0) / CHUNK_SIZE))
	return fr.Hash
}

func (fr *FileRef) GetType() string {
	return fr.Type
}

func (fr *FileRef) GetNumBlocks() int64 {
	return fr.NumBlocks
}

func (fr *FileRef) GetSize() int64 {
	return fr.Size
}

func (fr *FileRef) GetPathHash() string {
	return fr.PathHash
}

func (fr *FileRef) GetLookupHash() string {
	return fr.LookupHash
}

func (fr *FileRef) GetPath() string {
	return fr.Path
}
func (fr *FileRef) GetName() string {
	return fr.Name
}

func (fr *FileRef) GetCreatedAt() common.Timestamp {
	return fr.CreatedAt
}

func (fr *FileRef) GetUpdatedAt() common.Timestamp {
	return fr.UpdatedAt
}
