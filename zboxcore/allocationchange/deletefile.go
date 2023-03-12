package allocationchange

import (
	"path"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/fileref"
)

type DeleteFileChange struct {
	change
	ObjectTree fileref.RefEntity
}

func (ch *DeleteFileChange) ProcessChange(rootRef *fileref.Ref) (commitParams CommitParams, err error) {

	if ch.ObjectTree.GetPath() == "/" {
		rootRef.Children = nil
		rootRef.CalculateHash()
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
			return
		}
	}

	err = errors.New("file_not_found", "File to delete not found in blobber")
	return
}

func (n *DeleteFileChange) GetAffectedPath() []string {
	if n.ObjectTree != nil {
		return []string{n.ObjectTree.GetPath()}
	}
	return nil
}

func (n *DeleteFileChange) GetSize() int64 {
	if n.ObjectTree != nil {
		return -n.ObjectTree.GetSize()
	}
	return int64(0)
}
