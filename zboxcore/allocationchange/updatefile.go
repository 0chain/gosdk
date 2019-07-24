package allocationchange

import (
	"path/filepath"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/fileref"
)

type UpdateFileChange struct {
	change
	OldFile *fileref.FileRef
	NewFile *fileref.FileRef
}

func (ch *UpdateFileChange) ProcessChange(rootRef *fileref.Ref) error {
	path, _ := filepath.Split(ch.NewFile.Path)
	tSubDirs := getSubDirs(path)
	dirRef := rootRef
	treelevel := 0
	for treelevel < len(tSubDirs) {
		found := false
		for _, child := range dirRef.Children {
			if child.GetType() == fileref.DIRECTORY && treelevel < len(tSubDirs) {
				if (child.(*fileref.Ref)).Name == tSubDirs[treelevel] {
					dirRef = child.(*fileref.Ref)
					found = true
					break
				}
			}
		}
		if found {
			treelevel++
		} else {
			return common.NewError("invalid_reference_path", "Invalid reference path from the blobber")
		}
	}
	idx := -1
	for i, child := range dirRef.Children {
		if child.GetType() == fileref.FILE && child.GetPath() == ch.NewFile.Path {
			ch.OldFile = child.(*fileref.FileRef)
			idx = i
			break
		}
	}
	if idx < 0 || ch.OldFile == nil {
		return common.NewError("file_not_found", "File to update not found in blobber")
	}
	dirRef.Children[idx] = ch.NewFile
	rootRef.CalculateHash()
	return nil
}

func (n *UpdateFileChange) GetAffectedPath() string {
	if n.NewFile != nil {
		return n.NewFile.Path
	}
	return ""
}

func (n *UpdateFileChange) GetSize() int64 {
	if n.NewFile != nil && n.OldFile != nil {
		return n.NewFile.Size - n.OldFile.Size
	}
	return int64(0)
}
