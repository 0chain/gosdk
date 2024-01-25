package allocationchange

import (
	"path"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/pathutil"
	"github.com/0chain/gosdk/zboxcore/fileref"
)

type RenameFileChange struct {
	change
	ObjectTree fileref.RefEntity
	NewName    string
}

func (ch *RenameFileChange) ProcessChange(rootRef *fileref.Ref, _ map[string]string) (err error) {
	parentPath := path.Dir(ch.ObjectTree.GetPath())
	fields, err := common.GetPathFields(parentPath)
	if err != nil {
		return
	}
	dirRef := rootRef
	for i := 0; i < len(fields); i++ {
		found := false
		for _, child := range dirRef.Children {
			if child.GetName() == fields[i] {
				dirRef = child.(*fileref.Ref)
				found = true
				break
			}
		}
		if !found {
			err = errors.New("invalid_reference_path", "Invalid reference path from the blobber")
			return
		}
	}

	found := false
	var affectedRef *fileref.Ref
	for i, child := range dirRef.Children {
		if child.GetPath() == ch.ObjectTree.GetPath() {
			dirRef.RemoveChild(i)

			if ch.ObjectTree.GetType() == fileref.FILE {
				affectedRef = &(ch.ObjectTree.(*fileref.FileRef)).Ref
			} else {
				affectedRef = ch.ObjectTree.(*fileref.Ref)
			}

			affectedRef.Path = pathutil.Join(parentPath, ch.NewName)
			affectedRef.Name = ch.NewName
			affectedRef.HashToBeComputed = true

			dirRef.AddChild(ch.ObjectTree)
			found = true
			break
		}
	}

	if !found {
		err = errors.New("file_not_found", "Object to rename not found in blobber")
		return
	}
	ch.processChildren(affectedRef)
	return
}

func (ch *RenameFileChange) processChildren(curRef *fileref.Ref) {
	for _, childRefEntity := range curRef.Children {
		var childRef *fileref.Ref
		if childRefEntity.GetType() == fileref.FILE {
			childRef = &(childRefEntity.(*fileref.FileRef)).Ref
		} else {
			childRef = childRefEntity.(*fileref.Ref)
		}
		childRef.HashToBeComputed = true
		childRef.Path = pathutil.Join(curRef.Path, childRef.Name)
		if childRefEntity.GetType() == fileref.DIRECTORY {
			ch.processChildren(childRef)
		}
	}
}

func (n *RenameFileChange) GetAffectedPath() []string {
	if n.ObjectTree != nil {
		return []string{n.ObjectTree.GetPath()}
	}
	return nil
}

func (n *RenameFileChange) GetSize() int64 {
	return int64(0)
}
