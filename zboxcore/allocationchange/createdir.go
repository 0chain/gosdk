package allocationchange

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/fileref"
)

type DirCreateChange struct {
	RemotePath string
}

func (d *DirCreateChange) ProcessChange(
	rootRef *fileref.Ref, latestFileID int64) (
	inodesMeta map[string]int64, latestInode int64, err error) {

	inodesMeta = make(map[string]int64)
	fields, err := common.GetPathFields(d.RemotePath)
	if err != nil {
		return
	}
	dirRef := rootRef
	for i := 0; i < len(fields); i++ {
		found := false
		for _, child := range dirRef.Children {
			ref, ok := child.(*fileref.Ref)
			if !ok {
				fr, ok := child.(*fileref.FileRef)
				if !ok {
					err = errors.New("invalid_ref: child node is not valid *fileref.Ref or *fileref.FileRef ")
					return
				}
				ref = &fr.Ref
			}

			if ref.Name == fields[i] {
				dirRef = ref
				found = true
				break
			}
		}
		if !found {
			latestFileID++
			newRef := &fileref.Ref{
				Type:         fileref.DIRECTORY,
				AllocationID: dirRef.AllocationID,
				Path:         filepath.Join("/", strings.Join(fields[:i+1], "/")),
				Name:         fields[i],
				FileID:       latestFileID,
			}
			inodesMeta[newRef.Path] = latestFileID
			newRef.HashToBeComputed = true
			dirRef.AddChild(newRef)
			dirRef = newRef
		}
	}

	rootRef.CalculateHash()
	return
}

func (d *DirCreateChange) GetAffectedPath() string {
	return d.RemotePath
}

func (d *DirCreateChange) GetSize() int64 {
	return 0
}
