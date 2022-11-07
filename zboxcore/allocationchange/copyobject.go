package allocationchange

import (
	"strings"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type CopyFileChange struct {
	change
	ObjectTree fileref.RefEntity
	DestPath   string
}

func (ch *CopyFileChange) ProcessChange(rootRef *fileref.Ref, latestFileID int64) (
	commitParam CommitParams, err error) {

	inodesMeta := make(map[string]int64)
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
			latestFileID++
			newRef := &fileref.Ref{}
			newRef.Path = "/" + strings.Join(fields[:i+1], "/")
			newRef.FileID = latestFileID
			inodesMeta[newRef.Path] = latestFileID
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
	commitParam.WmFileID = dirRef.FileID
	var affectedRef *fileref.Ref
	if ch.ObjectTree.GetType() == fileref.FILE {
		affectedRef = &(ch.ObjectTree.(*fileref.FileRef)).Ref
	} else {
		affectedRef = ch.ObjectTree.(*fileref.Ref)
	}

	affectedRef.Path = zboxutil.Join(dirRef.GetPath(), affectedRef.Name)
	latestFileID++
	affectedRef.FileID = latestFileID
	affectedRef.HashToBeComputed = true
	inodesMeta[affectedRef.Path] = latestFileID
	commitParam.LatestFileID = ch.processChildren(affectedRef, inodesMeta, latestFileID)
	commitParam.InodesMeta = inodesMeta
	commitParam.Operation = marker.Copy
	dirRef.AddChild(ch.ObjectTree)

	rootRef.CalculateHash()
	return
}

func (ch *CopyFileChange) processChildren(
	curRef *fileref.Ref, inodesMeta map[string]int64, latestFileID int64) int64 {

	for _, childRefEntity := range curRef.Children {
		var childRef *fileref.Ref
		if childRefEntity.GetType() == fileref.FILE {
			childRef = &(childRefEntity.(*fileref.FileRef)).Ref
		} else {
			childRef = childRefEntity.(*fileref.Ref)
		}

		childRef.HashToBeComputed = true
		childRef.Path = zboxutil.Join(curRef.Path, childRef.Name)
		latestFileID++
		childRef.FileID = latestFileID
		inodesMeta[childRef.Path] = latestFileID

		if childRefEntity.GetType() == fileref.DIRECTORY {
			latestFileID = ch.processChildren(childRef, inodesMeta, latestFileID)
		}
	}

	return latestFileID
}

func (n *CopyFileChange) GetAffectedPath() string {
	return n.DestPath
}

func (n *CopyFileChange) GetSize() int64 {
	return n.ObjectTree.GetSize()
}
