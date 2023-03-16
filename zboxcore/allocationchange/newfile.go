package allocationchange

import (
	"errors"
	"fmt"
	"path"
	"strings"

	zError "github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/google/uuid"
)

type NewFileChange struct {
	change
	File *fileref.FileRef
	Uuid uuid.UUID
}

func (ch *NewFileChange) ProcessChange(rootRef *fileref.Ref) (
	commitParams CommitParams, err error) {

	fileIDMeta := make(map[string]string)
	fields, err := common.GetPathFields(common.GetPathDir(ch.File.Path))
	if err != nil {
		return
	}

	if ch.File.ActualFileHash == "" {
		err = errors.New("empty actual file hash field")
		return
	}

	if ch.File.ValidationRoot == "" {
		err = errors.New("empty validation root field")
		return
	}

	fileHashSign, err := client.Sign(ch.File.ActualFileHash)
	if err != nil {
		return
	}

	validationRootSign, err := client.Sign(fileHashSign + ch.File.ValidationRoot)
	if err != nil {
		return
	}

	ch.File.ActualFileHashSignature = fileHashSign
	ch.File.ValidationRootSignature = validationRootSign

	rootRef.HashToBeComputed = true
	dirRef := rootRef
	for i := 0; i < len(fields); i++ {
		found := false
		for _, child := range dirRef.Children {
			if child.GetName() == fields[i] {
				if child.GetType() == fileref.DIRECTORY {
					dirRef = child.(*fileref.Ref)
					found = true
					break
				}
				err = zError.New("invalid_file_path",
					fmt.Sprintf("type of %s is required to be directory", child.GetPath()))
				return
			}
		}

		if !found {
			uid := util.GetSHA1Uuid(ch.Uuid, fields[i])
			ch.Uuid = uid
			newRef := &fileref.Ref{
				Type:         fileref.DIRECTORY,
				AllocationID: dirRef.AllocationID,
				Path:         path.Join("/", strings.Join(fields[:i+1], "/")),
				Name:         fields[i],
				FileID:       uid.String(),
			}
			fileIDMeta[newRef.Path] = newRef.FileID
			dirRef.AddChild(newRef)
			dirRef = newRef
		}
		dirRef.HashToBeComputed = true
	}
	uid := util.GetSHA1Uuid(ch.Uuid, ch.File.Name)
	ch.Uuid = uid

	ch.File.FileID = uid.String()
	ch.File.HashToBeComputed = true
	fileIDMeta[ch.File.GetPath()] = ch.File.FileID

	dirRef.AddChild(ch.File)
	rootRef.CalculateHash()
	commitParams.FileIDMeta = fileIDMeta
	return
}

func (n *NewFileChange) GetAffectedPath() []string {
	if n.File != nil {
		return []string{n.File.Path}
	}
	return nil
}

func (n *NewFileChange) GetSize() int64 {
	if n.File != nil {
		return n.File.Size
	}
	return int64(0)
}
