package allocationchange

import (
	"strings"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type CopyFileChange struct {
	change
	ObjectTree fileref.RefEntity
	DestPath   string
}

func (ch *CopyFileChange) ProcessChange(rootRef *fileref.Ref) error {
	// path, _ := filepath.Split(ch.DestPath)
	tSubDirs := getSubDirs(ch.DestPath)
	dirRef := rootRef
	treelevel := 0
	for true {
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
			continue
		}
		if len(tSubDirs) <= treelevel {
			break
		}
		newRef := &fileref.Ref{}
		newRef.Type = fileref.DIRECTORY
		newRef.AllocationID = dirRef.AllocationID
		newRef.Path = "/" + strings.Join(tSubDirs[:treelevel+1], "/")
		newRef.Name = tSubDirs[treelevel]
		dirRef.AddChild(newRef)
		dirRef = newRef
		treelevel++
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

	dirRef.AddChild(ch.ObjectTree)

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

func (n *CopyFileChange) GetAffectedPath() string {
	return n.DestPath
}

func (n *CopyFileChange) GetSize() int64 {
	return n.ObjectTree.GetSize()
}
