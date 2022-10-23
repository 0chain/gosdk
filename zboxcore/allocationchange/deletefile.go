package allocationchange

import (
	"path"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/marker"
)

type DeleteFileChange struct {
	change
	ObjectTree fileref.RefEntity
}

func (ch *DeleteFileChange) ProcessChange(
	rootRef *fileref.Ref, _ int64) (
	commitParams CommitParams, err error) {

	if ch.ObjectTree.GetPath() == "/" {
		rootRef.Children = nil
		rootRef.CalculateHash()
		commitParams.WmFileID = rootRef.FileID
		commitParams.Operation = marker.Delete
		return
	}

	parentPath := path.Dir(ch.ObjectTree.GetPath())

	fields, err := common.GetPathFields(parentPath)
	if err != nil {
		return
	}

	dirRef := rootRef
	for _, name := range fields {
		found := false
		for _, child := range dirRef.Children {
			if child.GetName() == name {
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

	for i, child := range dirRef.Children {
		if child.GetName() == ch.ObjectTree.GetName() {
			dirRef.RemoveChild(i)
			rootRef.CalculateHash()
			commitParams.WmFileID = child.GetFileID()
			commitParams.Operation = marker.Delete
			return
		}
	}

	err = errors.New("file_not_found", "File to delete not found in blobber")
	return
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
