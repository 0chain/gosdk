package allocationchange

import (
	"strings"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/pathutil"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/google/uuid"
)

type CopyFileChange struct {
	change
	ObjectTree fileref.RefEntity
	DestPath   string
	Uuid       uuid.UUID
}

func (ch *CopyFileChange) ProcessChange(rootRef *fileref.Ref, fileIDMeta map[string]string) (err error) {

	var fields []string
	fields, err = common.GetPathFields(ch.DestPath)
	if err != nil {
		return
	}
	rootRef.HashToBeComputed = true
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
			uid := util.GetSHA1Uuid(ch.Uuid, fields[i])
			ch.Uuid = uid
			newRef.FileID = uid.String()
			newRef.Path = "/" + strings.Join(fields[:i+1], "/")
			fileIDMeta[newRef.Path] = newRef.FileID
			newRef.Type = fileref.DIRECTORY
			newRef.AllocationID = dirRef.AllocationID
			newRef.Name = fields[i]

			dirRef.AddChild(newRef)
			dirRef = newRef
		}
		dirRef.HashToBeComputed = true
	}

	if dirRef.GetPath() != ch.DestPath || dirRef.GetType() != fileref.DIRECTORY {
		err = errors.New("file_not_found", "Object to copy not found in blobber")
		return
	}
	var affectedRef *fileref.Ref
	if ch.ObjectTree.GetType() == fileref.FILE {
		affectedRef = &(ch.ObjectTree.(*fileref.FileRef)).Ref
	} else {
		affectedRef = ch.ObjectTree.(*fileref.Ref)
	}

	affectedRef.Path = pathutil.Join(dirRef.GetPath(), affectedRef.Name)
	uid := util.GetSHA1Uuid(ch.Uuid, affectedRef.Name)
	ch.Uuid = uid
	affectedRef.FileID = uid.String()

	affectedRef.HashToBeComputed = true
	fileIDMeta[affectedRef.Path] = affectedRef.FileID

	ch.processChildren(affectedRef, fileIDMeta)
	dirRef.AddChild(ch.ObjectTree)

	rootRef.CalculateHash()
	return
}

func (ch *CopyFileChange) processChildren(curRef *fileref.Ref, fileIDMeta map[string]string) {
	for _, childRefEntity := range curRef.Children {
		var childRef *fileref.Ref
		if childRefEntity.GetType() == fileref.FILE {
			childRef = &(childRefEntity.(*fileref.FileRef)).Ref
		} else {
			childRef = childRefEntity.(*fileref.Ref)
		}

		childRef.HashToBeComputed = true
		childRef.Path = pathutil.Join(curRef.Path, childRef.Name)
		uid := util.GetSHA1Uuid(ch.Uuid, childRef.Name)
		ch.Uuid = uid
		childRef.FileID = uid.String()
		fileIDMeta[childRef.Path] = childRef.FileID

		if childRefEntity.GetType() == fileref.DIRECTORY {
			ch.processChildren(childRef, fileIDMeta)
		}
	}
}

func (n *CopyFileChange) GetAffectedPath() []string {
	return []string{n.DestPath}
}

func (n *CopyFileChange) GetSize() int64 {
	return n.ObjectTree.GetSize()
}
