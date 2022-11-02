package allocationchange

import (
	"strings"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type CopyFileChange struct {
	change
	ObjectTree fileref.RefEntity
	DestPath   string
}

func (ch *CopyFileChange) ProcessChange(rootRef *fileref.Ref) error {

	fields, err := common.GetPathFields(ch.DestPath)
	if err != nil {
		return err
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
			newRef := &fileref.Ref{}
			newRef.Type = fileref.DIRECTORY
			newRef.AllocationID = dirRef.AllocationID
			newRef.Path = "/" + strings.Join(fields[:i+1], "/")
			newRef.Name = fields[i]
			newRef.HashToBeComputed = true
			dirRef.AddChild(newRef)
			dirRef = newRef
		}
	}

	rootRef.HashToBeComputed = true

	var affectedRef *fileref.Ref
	if ch.ObjectTree.GetType() == fileref.FILE {
		affectedRef = &(ch.ObjectTree.(*fileref.FileRef)).Ref
	} else {
		affectedRef = ch.ObjectTree.(*fileref.Ref)
	}

	affectedRef.Path = zboxutil.Join(dirRef.GetPath(), affectedRef.Name)
	ch.processChildren(affectedRef)
	dirRef.AddChild(affectedRef)

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
		childRef.Path = zboxutil.Join(curRef.Path, childRef.Name)
		if childRefEntity.GetType() == fileref.DIRECTORY {
			ch.processChildren(childRef)
		}
	}
}

func (n *CopyFileChange) GetAffectedPath() []string {
	return []string{n.DestPath}
}

func (n *CopyFileChange) GetSize() int64 {
	return n.ObjectTree.GetSize()
}
