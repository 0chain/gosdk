package allocationchange

import (
	"path"
	"strings"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/pathutil"
	"github.com/0chain/gosdk/zboxcore/fileref"
)

type MoveFileChange struct {
	change
	ObjectTree fileref.RefEntity
	DestPath   string
}

func (ch *MoveFileChange) ProcessChange(rootRef *fileref.Ref, fileIDMeta map[string]string) (err error) {

	fields, err := common.GetPathFields(ch.DestPath)
	if err != nil {
		return
	}
	rootRef.HashToBeComputed = true
	dirRef := rootRef
	for i := 0; i < len(fields); i++ {
		found := false
		for _, child := range dirRef.Children {
			if child.GetName() == fields[i] {
				if child.GetType() != fileref.DIRECTORY {
					err = errors.New("invalid_reference_path", "Invalid reference path from the blobber")
					return
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
				Path:         path.Join("/", strings.Join(fields[:i+1], "/")),
				Name:         fields[i],
			}
			dirRef.AddChild(newRef)
			dirRef = newRef
		}
		dirRef.HashToBeComputed = true
	}

	if dirRef.GetPath() != ch.DestPath || dirRef.GetType() != fileref.DIRECTORY {
		err = errors.New("file_not_found", "Object to move not found in blobber")
		return
	}

	var affectedRef *fileref.Ref
	if ch.ObjectTree.GetType() == fileref.FILE {
		affectedRef = &(ch.ObjectTree.(*fileref.FileRef)).Ref
	} else {
		affectedRef = ch.ObjectTree.(*fileref.Ref)
	}

	oldParentPath, oldFileName := pathutil.Split(ch.ObjectTree.GetPath())
	affectedRef.Path = pathutil.Join(dirRef.GetPath(), affectedRef.Name)
	ch.processChildren(affectedRef)

	dirRef.AddChild(ch.ObjectTree)

	fields, err = common.GetPathFields(oldParentPath)
	if err != nil {
		return
	}

	delRef := rootRef
	for i := 0; i < len(fields); i++ {
		found := false
		for _, child := range delRef.Children {
			if child.GetName() == fields[i] {
				delRef = child.(*fileref.Ref)
				delRef.HashToBeComputed = true
				found = true
				break
			}
		}

		if !found {
			err = errors.New("invalid_reference_path", "Ref not found in root reference object")
			return
		}
	}

	var removed bool
	for i, child := range delRef.Children {
		if child.GetName() == oldFileName {
			delRef.RemoveChild(i)
			removed = true
			break
		}
	}

	if !removed {
		err = errors.New("incomplete_move", "could not remove ref from source path")
		return
	}

	rootRef.CalculateHash()
	return
}

func (ch *MoveFileChange) processChildren(curRef *fileref.Ref) {
	for _, childRefEntity := range curRef.Children {
		var childRef *fileref.Ref
		if childRefEntity.GetType() == fileref.FILE {
			childRef = &(childRefEntity.(*fileref.FileRef)).Ref
		} else {
			childRef = childRefEntity.(*fileref.Ref)
		}
		childRef.Path = pathutil.Join(curRef.Path, childRef.Name)
		if childRefEntity.GetType() == fileref.DIRECTORY {
			ch.processChildren(childRef)
		}
	}
}

func (n *MoveFileChange) GetAffectedPath() []string {
	return []string{n.DestPath, pathutil.Dir(n.ObjectTree.GetPath())}
}

func (n *MoveFileChange) GetSize() int64 {
	return 0
}
