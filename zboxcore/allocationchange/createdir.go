package allocationchange

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/google/uuid"
)

type DirCreateChange struct {
	Timestamp  common.Timestamp
	RemotePath string
	Uuid       uuid.UUID
}

func (d *DirCreateChange) ProcessChange(rootRef *fileref.Ref) (commitParams CommitParams, err error) {
	inodesMeta := make(map[string]string)
	fields, err := common.GetPathFields(d.RemotePath)
	if err != nil {
		return
	}

	dirRef := rootRef
	for i := 0; i < len(fields); i++ {
		found := false
		for _, child := range dirRef.Children {
			ref, ok := child.(*fileref.FileRef)
			if !ok {
				err = errors.New("invalid_ref: child node is not valid *fileref.Ref")
				return
			}

			if ref.Name == fields[i] {

				dirRef = &ref.Ref
				found = true
				break
			}
		}
		if !found {
			uid := util.GetSHA1Uuid(d.Uuid, fields[i])
			d.Uuid = uid

			newRef := &fileref.Ref{
				Type:         fileref.DIRECTORY,
				AllocationID: dirRef.AllocationID,
				Path:         filepath.Join("/", strings.Join(fields[:i+1], "/")),
				Name:         fields[i],
				FileID:       uid.String(),
			}
			inodesMeta[newRef.Path] = newRef.FileID
			newRef.HashToBeComputed = true
			dirRef.AddChild(newRef)
			dirRef = newRef
		}
	}

	commitParams.FileIDMeta = inodesMeta
	rootRef.CalculateHash()
	return
}

func (d *DirCreateChange) GetAffectedPath() []string {
	return []string{d.RemotePath}
}

func (d *DirCreateChange) GetSize() int64 {
	return 0
}
