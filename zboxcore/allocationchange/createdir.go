package allocationchange

import (
	"path/filepath"
	"strings"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/fileref"
)

type DirCreateChange struct {
	RemotePath string
}

func (d *DirCreateChange) ProcessChange(rootRef *fileref.Ref) error {
	fields, err := common.GetPathFields(d.RemotePath)
	if err != nil {
		return err
	}
	dirRef := rootRef
	for i := 0; i < len(fields); i++ {
		found := false
		for _, child := range dirRef.Children {
			ref, ok := child.(*fileref.Ref)
			if !ok {
				fr, ok := child.(*fileref.FileRef)
				if ok {
					ref = &fr.Ref
				}
			}

			if ref.Name == fields[i] {
				dirRef = ref
				found = true
				break
			}
		}
		if !found {
			newRef := &fileref.Ref{
				Type:         fileref.DIRECTORY,
				AllocationID: dirRef.AllocationID,
				Path:         filepath.Join("/", strings.Join(fields[:i+1], "/")),
				Name:         fields[i],
			}
			newRef.HashToBeComputed = true
			dirRef.AddChild(newRef)
			dirRef = newRef
		}
	}

	rootRef.CalculateHash()
	return nil
}

func (d *DirCreateChange) GetAffectedPath() string {
	return d.RemotePath
}

func (d *DirCreateChange) GetSize() int64 {
	return 0
}
