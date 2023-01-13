package allocationchange

import (
	"errors"
	"path"
	"path/filepath"
	"strings"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
)

type NewFileChange struct {
	change
	File *fileref.FileRef
}

func (ch *NewFileChange) ProcessChange(rootRef *fileref.Ref) error {
	tSubDirs, err := common.GetPathFields(path.Dir(ch.File.Path))
	if err != nil {
		return err
	}

	if ch.File.ActualFileHash == "" {
		return errors.New("empty actual file hash field")
	}

	if ch.File.ValidationRoot == "" {
		return errors.New("empty validation root field")
	}

	fileHashSign, err := client.Sign(ch.File.ActualFileHash)
	if err != nil {
		return err
	}

	validationRootSign, err := client.Sign(fileHashSign + ch.File.ValidationRoot)
	if err != nil {
		return err
	}

	ch.File.ActualFileHashSignature = fileHashSign
	ch.File.ValidationRootSignature = validationRootSign

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

func (n *NewFileChange) GetAffectedPath() []string {
	if n.File != nil {
		return []string{n.File.Path}
	}
	return nil
}

func (n *NewFileChange) GetSize() int64 {
	if n.File != nil {
		return n.File.Size
	}
	return int64(0)
}
