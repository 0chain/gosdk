package fileref

import (
	"fmt"
	"hash/fnv"
	"math"
	"strconv"
	"strings"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/encryption"
	lru "github.com/hashicorp/golang-lru/v2"
)

const CHUNK_SIZE = 64 * 1024

const (
	FILE      = "f"
	DIRECTORY = "d"
	REGULAR   = "regular"
)

var fileCache, _ = lru.New[string, FileRef](100)

type Collaborator struct {
	RefID     int64  `json:"ref_id"`
	ClientID  string `json:"client_id"`
	CreatedAt string `json:"created_at"`
}

type FileRef struct {
	Ref        `mapstructure:",squash"`
	CustomMeta string `json:"custom_meta" mapstructure:"custom_meta"`
	// ValidationRootSignature is signature signed by client for hash_of(ActualFileHashSignature + ValidationRoot)
	ThumbnailSize  int64  `json:"thumbnail_size" mapstructure:"thumbnail_size"`
	ThumbnailHash  string `json:"thumbnail_hash" mapstructure:"thumbnail_hash"`
	ActualFileSize int64  `json:"actual_file_size" mapstructure:"actual_file_size"`
	ActualFileHash string `json:"actual_file_hash" mapstructure:"actual_file_hash"`
	// ActualFileHashSignature is signature signed by client for ActualFileHash
	ActualFileHashSignature string         `json:"actual_file_hash_signature" mapstructure:"actual_file_hash_signature"`
	ActualThumbnailSize     int64          `json:"actual_thumbnail_size" mapstructure:"actual_thumbnail_size"`
	ActualThumbnailHash     string         `json:"actual_thumbnail_hash" mapstructure:"actual_thumbnail_hash"`
	MimeType                string         `json:"mimetype" mapstructure:"mimetype"`
	EncryptedKey            string         `json:"encrypted_key" mapstructure:"encrypted_key"`
	EncryptedKeyPoint       string         `json:"encrypted_key_point" mapstructure:"encrypted_key_point"`
	Collaborators           []Collaborator `json:"collaborators" mapstructure:"collaborators"`
}

func (fRef *FileRef) MetaID() string {

	hash := fnv.New64a()
	hash.Write([]byte(fRef.Path))

	return strconv.FormatUint(hash.Sum64(), 36)
}

type RefEntity interface {
	GetNumBlocks() int64
	GetSize() int64
	GetFileMetaHash() string
	GetHash() string
	CalculateHash() string
	GetType() string
	GetPathHash() string
	GetLookupHash() string
	GetPath() string
	GetName() string
	GetFileID() string
	GetCreatedAt() common.Timestamp
	GetUpdatedAt() common.Timestamp
	GetAllocationVersion() int64
}

type Ref struct {
	Type                string `json:"type" mapstructure:"type"`
	AllocationID        string `json:"allocation_id" mapstructure:"allocation_id"`
	Name                string `json:"name" mapstructure:"name"`
	Path                string `json:"path" mapstructure:"path"`
	Size                int64  `json:"size" mapstructure:"size"`
	ActualSize          int64  `json:"actual_file_size" mapstructure:"actual_file_size"`
	Hash                string `json:"hash" mapstructure:"hash"`
	ChunkSize           int64  `json:"chunk_size" mapstructure:"chunk_size"`
	NumBlocks           int64  `json:"num_of_blocks" mapstructure:"num_of_blocks"`
	PathHash            string `json:"path_hash" mapstructure:"path_hash"`
	LookupHash          string `json:"lookup_hash" mapstructure:"lookup_hash"`
	FileID              string `json:"file_id" mapstructure:"file_id"`
	FileMetaHash        string `json:"file_meta_hash" mapstructure:"file_meta_hash"`
	ThumbnailHash       string `json:"thumbnail_hash" mapstructure:"thumbnail_hash"`
	ThumbnailSize       int64  `json:"thumbnail_size" mapstructure:"thumbnail_size"`
	ActualThumbnailHash string `json:"actual_thumbnail_hash" mapstructure:"actual_thumbnail_hash"`
	ActualThumbnailSize int64  `json:"actual_thumbnail_size" mapstructure:"actual_thumbnail_size"`
	IsEmpty             bool   `json:"is_empty" mapstructure:"is_empty"`
	AllocationVersion   int64  `json:"allocation_version" mapstructure:"allocation_version"`
	HashToBeComputed    bool
	ChildrenLoaded      bool
	Children            []RefEntity      `json:"-" mapstructure:"-"`
	CreatedAt           common.Timestamp `json:"created_at" mapstructure:"created_at"`
	UpdatedAt           common.Timestamp `json:"updated_at" mapstructure:"updated_at"`
}

func GetReferenceLookup(allocationID string, path string) string {
	return encryption.Hash(allocationID + ":" + path)
}

func GetCacheKey(lookuphash, blobberID string) string {
	return encryption.FastHash(lookuphash + ":" + blobberID)
}

func StoreFileRef(key string, fr FileRef) {
	fileCache.Add(key, fr)
}

func GetFileRef(key string) (FileRef, bool) {
	if fr, ok := fileCache.Get(key); ok {
		return fr, true
	}
	return FileRef{}, false
}

func DeleteFileRef(key string) {
	fileCache.Remove(key)
}

func (r *Ref) CalculateHash() string {
	if len(r.Children) == 0 && !r.ChildrenLoaded && !r.HashToBeComputed {
		return r.Hash
	}

	childHashes := make([]string, len(r.Children))
	childFileMetaHashes := make([]string, len(r.Children))
	childPaths := make([]string, len(r.Children))
	var refNumBlocks int64
	var size int64

	for index, childRef := range r.Children {
		childRef.CalculateHash()
		childFileMetaHashes[index] = childRef.GetFileMetaHash()
		childHashes[index] = childRef.GetHash()
		childPaths[index] = childRef.GetPath()
		refNumBlocks += childRef.GetNumBlocks()
		size += childRef.GetSize()
	}

	r.FileMetaHash = encryption.Hash(r.GetPath() + strings.Join(childFileMetaHashes, ":"))
	r.Hash = encryption.Hash(r.GetHashData() + strings.Join(childHashes, ":"))

	r.PathHash = encryption.Hash(strings.Join(childPaths, ":"))
	r.NumBlocks = refNumBlocks
	r.Size = size

	return r.Hash
}

func (r *Ref) GetFileMetaHash() string {
	return r.FileMetaHash
}

func (r *Ref) GetHash() string {
	return r.Hash
}

func (r *Ref) GetHashData() string {
	return fmt.Sprintf("%s:%s:%s", r.AllocationID, r.Path, r.FileID)
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

func (r *Ref) GetFileID() string {
	return r.FileID
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

func (r *Ref) GetAllocationVersion() int64 {
	return r.AllocationVersion
}

func (fr *FileRef) GetFileMetaHash() string {
	return fr.FileMetaHash
}
func (fr *FileRef) GetFileMetaHashData() string {
	return fmt.Sprintf(
		"%s:%d:%d:%s",
		fr.Path, fr.Size,
		fr.ActualFileSize, fr.ActualFileHash)
}

func (fr *FileRef) GetHashData() string {
	return fmt.Sprintf(
		"%s:%s:%s:%s:%d:%d:%s:%d:%s",
		fr.AllocationID,
		fr.Type, // don't need to add it as well
		fr.Name, // don't see any utility as fr.Path below has name in it
		fr.Path,
		fr.Size,
		fr.ActualFileSize,
		fr.ActualFileHash,
		fr.ChunkSize,
		fr.FileID,
	)
}

func (fr *FileRef) GetHash() string {
	return fr.Hash
}

func (fr *FileRef) CalculateHash() string {
	fr.Hash = encryption.Hash(fr.GetHashData())
	fr.FileMetaHash = encryption.Hash(fr.GetFileMetaHashData())
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

func (fr *FileRef) GetFileID() string {
	return fr.FileID
}

func (fr *FileRef) GetCreatedAt() common.Timestamp {
	return fr.CreatedAt
}

func (fr *FileRef) GetUpdatedAt() common.Timestamp {
	return fr.UpdatedAt
}
