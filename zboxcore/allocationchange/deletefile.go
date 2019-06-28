package allocationchange

import (
	"path/filepath"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/fileref"
)

type DeleteFileChange struct {
	change
	File *fileref.FileRef
}

func (ch *DeleteFileChange) ProcessChange(rootRef *fileref.Ref) error {
	path, _ := filepath.Split(ch.File.Path)
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
		if child.GetType() == fileref.FILE && child.GetHash() == ch.File.Hash {
			idx = i
			break
		}
	}
	if idx < 0 {
		return common.NewError("file_not_found", "File to delete not found in blobber")
	}
	//dirRef.Children = append(dirRef.Children[:idx], dirRef.Children[idx+1:]...)
	dirRef.RemoveChild(idx)
	rootRef.CalculateHash()
	return nil
}

func (n *DeleteFileChange) GetAffectedPath() string {
	if n.File != nil {
		return n.File.Path
	}
	return ""
}

func (n *DeleteFileChange) GetSize() int64 {
	if n.File != nil {
		return 0 - (n.File.Size + n.File.ThumbnailSize)
	}
	return int64(0)
}
