package allocationchange

import (
	"path/filepath"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/zboxcore/fileref"
)

type DeleteFileChange struct {
	change
	//File *fileref.FileRef
	ObjectTree fileref.RefEntity
}

func (ch *DeleteFileChange) ProcessChange(rootRef *fileref.Ref) error {
	if ch.ObjectTree.GetPath() == "/" {
		rootRef.Children = nil
		rootRef.CalculateHash()
		return nil
	}

	path, _ := filepath.Split(ch.ObjectTree.GetPath())

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
			return errors.New("invalid_reference_path", "Invalid reference path from the blobber")
		}
	}
	for i, child := range dirRef.Children {
		if child.GetName() == ch.ObjectTree.GetName() && child.GetHash() == ch.ObjectTree.GetHash() {
			dirRef.RemoveChild(i)
			rootRef.CalculateHash()
			return nil
		}
	}
	return errors.New("file_not_found", "File to delete not found in blobber")
}

func (n *DeleteFileChange) GetAffectedPath() string {
	if n.ObjectTree != nil {
		return n.ObjectTree.GetPath()
	}
	return ""
}

func (n *DeleteFileChange) GetSize() int64 {
	if n.ObjectTree != nil {
		return 0 - n.ObjectTree.GetSize()
	}
	return int64(0)
}
