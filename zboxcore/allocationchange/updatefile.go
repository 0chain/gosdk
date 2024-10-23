package allocationchange

import (
	"fmt"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/pathutil"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
)

type UpdateFileChange struct {
	change
	OldFile *fileref.FileRef
	NewFile *fileref.FileRef
}

func (ch *UpdateFileChange) ProcessChange(rootRef *fileref.Ref, _ map[string]string) (err error) {

	if ch.NewFile.ActualFileHash == "" {
		err = fmt.Errorf("empty actual file hash field")
		return
	}

	fileHashSign, err := client.Sign(ch.NewFile.ActualFileHash)
	if err != nil {
		return
	}

	ch.NewFile.ActualFileHashSignature = fileHashSign

	fields, err := common.GetPathFields(pathutil.Dir(ch.NewFile.Path))

	if err != nil {
		return
	}

	rootRef.HashToBeComputed = true
	dirRef := rootRef
	for i := 0; i < len(fields); i++ {
		found := false
		for _, child := range dirRef.Children {
			if child.GetName() == fields[i] {
				var ok bool
				dirRef, ok = child.(*fileref.Ref)
				if !ok {
					err = errors.New("invalid_reference_path", "Invalid reference path from the blobber")
					return
				}
				dirRef.HashToBeComputed = true
				found = true
				break
			}
		}

		if !found {
			err = errors.New("invalid_reference_path", "Invalid reference path from the blobber")
			return
		}
	}
	idx := -1
	for i, child := range dirRef.Children {
		if child.GetType() == fileref.FILE && child.GetPath() == ch.NewFile.Path {
			ch.OldFile = child.(*fileref.FileRef)
			idx = i
			break
		}
	}
	if idx < 0 || ch.OldFile == nil {
		err = errors.New("file_not_found", "File to update not found in blobber")
		return
	}

	ch.NewFile.HashToBeComputed = true
	ch.NewFile.FileID = ch.OldFile.FileID
	dirRef.Children[idx] = ch.NewFile
	return
}

func (n *UpdateFileChange) GetAffectedPath() []string {
	if n.NewFile != nil {
		return []string{n.NewFile.Path}
	}
	return nil
}

func (n *UpdateFileChange) GetSize() int64 {
	if n.NewFile != nil && n.OldFile != nil {
		return n.NewFile.Size - n.OldFile.Size
	}
	return int64(0)
}
