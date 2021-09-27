package allocationchange

import (
	"path/filepath"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type RenameFileChange struct {
	change
	ObjectTree fileref.RefEntity
	NewName    string
}

func (ch *RenameFileChange) ProcessChange(rootRef *fileref.Ref) error {
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
	idx := -1
	for i, child := range dirRef.Children {
		if child.GetPath() == ch.ObjectTree.GetPath() && child.GetHash() == ch.ObjectTree.GetHash() {
			idx = i
			break
		}
	}
	if idx < 0 {
		return errors.New("file_not_found", "Object to rename not found in blobber")
	}
	dirRef.Children[idx] = ch.ObjectTree
	// Logger.Info("Old name: " + dirRef.Children[idx].GetName())
	var affectedRef *fileref.Ref
	if ch.ObjectTree.GetType() == fileref.FILE {
		affectedRef = &(ch.ObjectTree.(*fileref.FileRef)).Ref
	} else {
		affectedRef = ch.ObjectTree.(*fileref.Ref)
	}

	path, _ = filepath.Split(affectedRef.Path)
	path = zboxutil.RemoteClean(path)
	affectedRef.Name = ch.NewName
	affectedRef.Path = zboxutil.Join(path, ch.NewName)

	// Logger.Info("Changed name: " + dirRef.Children[idx].GetName())

	ch.processChildren(affectedRef)
	// Logger.Info("Process hash for renaming")
	rootRef.CalculateHash()
	return nil
}

func (ch *RenameFileChange) processChildren(curRef *fileref.Ref) {
	for _, childRefEntity := range curRef.Children {
		var childRef *fileref.Ref
		if childRefEntity.GetType() == fileref.FILE {
			childRef = &(childRefEntity.(*fileref.FileRef)).Ref
		} else {
			childRef = childRefEntity.(*fileref.Ref)
		}
		childRef.Path = filepath.Join(curRef.Path, childRef.Name)
		if childRefEntity.GetType() == fileref.DIRECTORY {
			ch.processChildren(childRef)
		}
	}
}

func (n *RenameFileChange) GetAffectedPath() string {
	if n.ObjectTree != nil {
		return n.ObjectTree.GetPath()
	}
	return ""
}

func (n *RenameFileChange) GetSize() int64 {
	return int64(0)
}
