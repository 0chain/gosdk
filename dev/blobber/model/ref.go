package model

import (
	"context"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/pathutil"
)

const (
	// FileRef represents a file
	FILE = "f"

	// FileRef represents a directory
	DIRECTORY = "d"

	CHUNK_SIZE = 64 * 1024

	DIR_LIST_TAG  = "dirlist"
	FILE_LIST_TAG = "filelist"
)

type Ref struct {
	Type            string `gorm:"column:type" dirlist:"type" filelist:"type"`
	AllocationID    string `gorm:"column:allocation_id"`
	LookupHash      string `gorm:"column:lookup_hash" dirlist:"lookup_hash" filelist:"lookup_hash"`
	Name            string `gorm:"column:name" dirlist:"name" filelist:"name"`
	Path            string `gorm:"column:path" dirlist:"path" filelist:"path"`
	Hash            string `gorm:"column:hash" dirlist:"hash" filelist:"hash"`
	NumBlocks       int64  `gorm:"column:num_of_blocks" dirlist:"num_of_blocks" filelist:"num_of_blocks"`
	PathHash        string `gorm:"column:path_hash" dirlist:"path_hash" filelist:"path_hash"`
	ParentPath      string `gorm:"column:parent_path"`
	PathLevel       int    `gorm:"column:level"`
	ValidationRoot  string `gorm:"column:validation_root" filelist:"validation_root"`
	Size            int64  `gorm:"column:size" dirlist:"size" filelist:"size"`
	FixedMerkleRoot string `gorm:"column:fixed_merkle_root" filelist:"fixed_merkle_root"`
	ActualFileSize  int64  `gorm:"column:actual_file_size" filelist:"actual_file_size"`
	ActualFileHash  string `gorm:"column:actual_file_hash" filelist:"actual_file_hash"`

	Children       []*Ref `gorm:"-"`
	childrenLoaded bool

	ChunkSize int64 `gorm:"column:chunk_size" dirlist:"chunk_size" filelist:"chunk_size"`
}

func (r *Ref) CalculateHash(ctx context.Context) (string, error) {
	if r.Type == DIRECTORY {
		return r.CalculateDirHash(ctx)
	}
	return r.CalculateFileHash(ctx)
}

// GetListingData reflect and convert all fields into map[string]interface{}
func (r *Ref) GetListingData(ctx context.Context) map[string]interface{} {
	if r == nil {
		return make(map[string]interface{})
	}

	if r.Type == FILE {
		return GetListingFieldsMap(*r, FILE_LIST_TAG)
	}
	return GetListingFieldsMap(*r, DIR_LIST_TAG)
}

func GetListingFieldsMap(refEntity interface{}, tagName string) map[string]interface{} {
	result := make(map[string]interface{})
	t := reflect.TypeOf(refEntity)
	v := reflect.ValueOf(refEntity)
	// Iterate over all available fields and read the tag value
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Get the field tag value
		tag := field.Tag.Get(tagName)
		// Skip if tag is not defined or ignored
		if !field.Anonymous && (tag == "" || tag == "-") {
			continue
		}

		if field.Anonymous {
			listMap := GetListingFieldsMap(v.FieldByName(field.Name).Interface(), tagName)
			if len(listMap) > 0 {
				for k, v := range listMap {
					result[k] = v
				}

			}
		} else {
			fieldValue := v.FieldByName(field.Name).Interface()
			if fieldValue == nil {
				continue
			}
			result[tag] = fieldValue
		}

	}
	return result
}

func GetSubDirsFromPath(p string) []string {
	path := p
	parent, cur := pathutil.Split(path)
	subDirs := make([]string, 0)
	for len(cur) > 0 {
		if cur == "." {
			break
		}
		subDirs = append([]string{cur}, subDirs...)
		parent, cur = pathutil.Split(parent)
	}
	return subDirs
}

func (r *Ref) CalculateDirHash(ctx context.Context) (string, error) {
	// empty directory, return hash directly
	if len(r.Children) == 0 && !r.childrenLoaded {
		return r.Hash, nil
	}
	childHashes := make([]string, len(r.Children))
	childPathHashes := make([]string, len(r.Children))
	var refNumBlocks int64
	var size int64
	for index, childRef := range r.Children {
		_, err := childRef.CalculateHash(ctx)
		if err != nil {
			return "", err
		}
		childHashes[index] = childRef.Hash
		childPathHashes[index] = childRef.PathHash
		refNumBlocks += childRef.NumBlocks
		size += childRef.Size
	}

	r.Hash = encryption.Hash(strings.Join(childHashes, ":"))
	r.NumBlocks = refNumBlocks
	r.Size = size
	r.PathHash = encryption.Hash(strings.Join(childPathHashes, ":"))
	r.PathLevel = len(GetSubDirsFromPath(r.Path)) + 1
	r.LookupHash = GetReferenceLookup(r.AllocationID, r.Path)

	return r.Hash, nil
}

// GetReferenceLookup hash(allocationID + ":" + path) which is used to lookup the file reference in the db.
//   - allocationID is the allocation ID.
//   - path is the path of the file.
func GetReferenceLookup(allocationID string, path string) string {
	return encryption.Hash(allocationID + ":" + path)
}

func (fr *Ref) CalculateFileHash(ctx context.Context) (string, error) {
	fr.Hash = encryption.Hash(fr.GetFileHashData())
	fr.NumBlocks = int64(math.Ceil(float64(fr.Size*1.0) / float64(fr.ChunkSize)))
	fr.PathHash = GetReferenceLookup(fr.AllocationID, fr.Path)
	fr.PathLevel = len(GetSubDirsFromPath(fr.Path)) + 1
	fr.LookupHash = GetReferenceLookup(fr.AllocationID, fr.Path)

	return fr.Hash, nil
}

func (fr *Ref) GetFileHashData() string {
	hashArray := make([]string, 0, 11)
	hashArray = append(hashArray, fr.AllocationID)
	hashArray = append(hashArray, fr.Type)
	hashArray = append(hashArray, fr.Name)
	hashArray = append(hashArray, fr.Path)
	hashArray = append(hashArray, strconv.FormatInt(fr.Size, 10))
	hashArray = append(hashArray, fr.ValidationRoot)
	hashArray = append(hashArray, fr.FixedMerkleRoot)
	hashArray = append(hashArray, strconv.FormatInt(fr.ActualFileSize, 10))
	hashArray = append(hashArray, fr.ActualFileHash)
	hashArray = append(hashArray, strconv.FormatInt(fr.ChunkSize, 10))

	return strings.Join(hashArray, ":")
}
