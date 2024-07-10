package allocationchange

import (
	"path"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/fileref"
)

type DeleteFileChange struct {
	change
	FileMetaRef fileref.RefEntity
}

func (ch *DeleteFileChange) ProcessChange(rootRef *fileref.Ref, _ map[string]string) (err error) {

	if ch.FileMetaRef.GetPath() == "/" {
		rootRef.Children = nil
		return
	}

	parentPath := path.Dir(ch.FileMetaRef.GetPath())

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
		if child.GetName() == ch.FileMetaRef.GetName() {
			dirRef.RemoveChild(i)
			return
		}
	}

	err = errors.New("file_not_found", "File to delete not found in blobber")
	return
}

func (n *DeleteFileChange) GetAffectedPath() []string {
	if n.FileMetaRef != nil {
		return []string{n.FileMetaRef.GetPath()}
	}
	return nil
}

func (n *DeleteFileChange) GetSize() int64 {
	if n.FileMetaRef != nil {
		return -n.FileMetaRef.GetSize()
	}
	return int64(0)
}
