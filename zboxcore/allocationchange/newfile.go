package allocationchange

import (
	"path/filepath"
	"strings"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/fileref"
)

type NewFileChange struct {
	change
	File *fileref.FileRef
}

func (ch *NewFileChange) ProcessChange(rootRef *fileref.Ref) error {
	tSubDirs, err := common.GetPathFields(filepath.Dir(ch.File.Path))
	if err != nil {
		return err
	}

	dirRef := rootRef
	for i := 0; i < len(tSubDirs); i++ {
		found := false
		for _, child := range dirRef.Children {
			if child.GetType() == fileref.DIRECTORY && child.(*fileref.Ref).Name == tSubDirs[i] {
				dirRef = child.(*fileref.Ref)
				found = true
				break
			}
		}
		if !found {
			newRef := &fileref.Ref{
				Type:         fileref.DIRECTORY,
				AllocationID: dirRef.AllocationID,
				Path:         filepath.Join("/", strings.Join(tSubDirs[:i+1], "/")),
				Name:         tSubDirs[i],
			}
			dirRef.AddChild(newRef)
			dirRef = newRef
		}
	}

	dirRef.AddChild(ch.File)
	rootRef.CalculateHash()
	return nil
}

func (n *NewFileChange) GetAffectedPath() string {
	if n.File != nil {
		return n.File.Path
	}
	return ""
}

func (n *NewFileChange) GetSize() int64 {
	if n.File != nil {
		return n.File.Size
	}
	return int64(0)
}
