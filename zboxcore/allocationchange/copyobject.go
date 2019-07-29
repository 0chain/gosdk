package allocationchange

import (
	"path/filepath"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/fileref"
)

type CopyFileChange struct {
	change
	ObjectTree fileref.RefEntity
	DestPath   string
}

func (ch *CopyFileChange) ProcessChange(rootRef *fileref.Ref) error {
	path, _ := filepath.Split(ch.DestPath)
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
	var foundRef fileref.RefEntity
	if dirRef.GetPath() == ch.DestPath && dirRef.GetType() == fileref.DIRECTORY {
		foundRef = dirRef
	} else {
		for i, child := range dirRef.Children {
			if child.GetPath() == ch.DestPath && child.GetType() == fileref.DIRECTORY {
				foundRef = dirRef.Children[i]
				break
			}
		}
	}

	if foundRef == nil {
		return common.NewError("file_not_found", "Object to copy not found in blobber")
	}

	var affectedRef *fileref.Ref
	if ch.ObjectTree.GetType() == fileref.FILE {
		affectedRef = &(ch.ObjectTree.(*fileref.FileRef)).Ref
	} else {
		affectedRef = ch.ObjectTree.(*fileref.Ref)
	}

	affectedRef.Path = filepath.Join(foundRef.GetPath(), affectedRef.Name)
	ch.processChildren(affectedRef)

	destRef := foundRef.(*fileref.Ref)
	destRef.AddChild(ch.ObjectTree)

	rootRef.CalculateHash()
	return nil
}

func (ch *CopyFileChange) processChildren(curRef *fileref.Ref) {
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

func (n *CopyFileChange) GetAffectedPath() string {
	return n.DestPath
}

func (n *CopyFileChange) GetSize() int64 {
	return n.ObjectTree.GetSize()
}
