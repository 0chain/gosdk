package allocationchange

import (
	"path/filepath"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/zboxcore/fileref"
)

type AttributesChange struct {
	change

	ConnectionID string             `json:"connection_id"`
	AllocationID string             `json:"allocation_id"`
	Path         string             `json:"path"`
	Attributes   fileref.Attributes `json:"attributes"` // new attributes
}

func (ac *AttributesChange) ProcessChange(root *fileref.Ref) (err error) {

	var (
		path, _   = filepath.Split(ac.Path)
		tSubDirs  = getSubDirs(path)
		dirRef    = root
		treelevel int
	)

	for treelevel < len(tSubDirs) {
		var found bool
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
		} else {
			return errors.New("attributes_change_process",
				"Invalid reference path from the blobber")
		}
	}

	var (
		idx  = -1
		file *fileref.FileRef
	)
	for i, child := range dirRef.Children {
		if child.GetType() == fileref.FILE && child.GetPath() == ac.Path {
			file = child.(*fileref.FileRef)
			idx = i
			break
		}
	}

	if idx < 0 || file == nil {
		return errors.New("attributes_change_process",
			"File, to update attributes for, not found in blobber")
	}

	file.Attributes = ac.Attributes // change attributes to new
	dirRef.Children[idx] = file     // set changed back
	root.CalculateHash()            // recalculate hash

	return
}

func (ac *AttributesChange) GetAffectedPath() string {
	return ac.Path
}

func (n *AttributesChange) GetSize() int64 {
	return int64(0)
}
