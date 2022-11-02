package allocationchange

import (
	"path/filepath"
	"strings"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type MoveFileChange struct {
	change
	ObjectTree fileref.RefEntity
	DestPath   string
}

func (ch *MoveFileChange) ProcessChange(rootRef *fileref.Ref) error {
	fields, err := common.GetPathFields(ch.DestPath)
	if err != nil {
		return err
	}
	rootRef.HashToBeComputed = true
	dirRef := rootRef
	for i := 0; i < len(fields); i++ {
		found := false
		for _, child := range dirRef.Children {
			if child.GetName() == fields[i] {
				if child.GetType() != fileref.DIRECTORY {
					return errors.New("invalid_reference_path", "Invalid reference path from the blobber")
				}
				dirRef = child.(*fileref.Ref)
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

	if dirRef.GetPath() != ch.DestPath || dirRef.GetType() != fileref.DIRECTORY {
		return errors.New("file_not_found", "Object to copy not found in blobber")
	}

	var affectedRef *fileref.Ref
	if ch.ObjectTree.GetType() == fileref.FILE {
		affectedRef = &(ch.ObjectTree.(*fileref.FileRef)).Ref
	} else {
		affectedRef = ch.ObjectTree.(*fileref.Ref)
	}

	affectedRef.Path = zboxutil.Join(dirRef.GetPath(), affectedRef.Name)
	ch.processChildren(affectedRef)

	dirRef.AddChild(affectedRef)

	delParentPath, delFileName := filepath.Split(ch.ObjectTree.GetPath())
	fields, err = common.GetPathFields(delParentPath)
	if err != nil {
		return err
	}

	delRef := rootRef
	for i := 0; i < len(fields); i++ {
		found := false
		for _, child := range delRef.Children {
			if child.GetName() == fields[i] {
				delRef = child.(*fileref.Ref)
				found = true
				break
			}
		}

		if !found {
			return errors.New("invalid_reference_path", "Ref not found in root reference object")
		}
	}

	var removed bool
	for i, child := range delRef.Children {
		if child.GetName() == delFileName {
			delRef.RemoveChild(i)
			removed = true
			break
		}
	}

	if !removed {
		return errors.New("incomplete_move", "could not remove ref from source path")
	}

	rootRef.CalculateHash()
	return nil
}

func (ch *MoveFileChange) processChildren(curRef *fileref.Ref) {
	for _, childRefEntity := range curRef.Children {
		var childRef *fileref.Ref
		if childRefEntity.GetType() == fileref.FILE {
			childRef = &(childRefEntity.(*fileref.FileRef)).Ref
		} else {
			childRef = childRefEntity.(*fileref.Ref)
		}
		childRef.Path = zboxutil.Join(curRef.Path, childRef.Name)
		if childRefEntity.GetType() == fileref.DIRECTORY {
			ch.processChildren(childRef)
		}
	}
}

func (n *MoveFileChange) GetAffectedPath() []string {
	return []string{n.DestPath, filepath.Dir(n.ObjectTree.GetPath())}
}

func (n *MoveFileChange) GetSize() int64 {
	return n.ObjectTree.GetSize()
}
