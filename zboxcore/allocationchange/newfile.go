package allocationchange

import (
	"path/filepath"
	"strings"

	"github.com/0chain/gosdk/zboxcore/fileref"
)

type NewFileChange struct {
	change
	File *fileref.FileRef
}

func (ch *NewFileChange) ProcessChange(rootRef *fileref.Ref) error {
	path, _ := filepath.Split(ch.File.Path)
	tSubDirs := getSubDirs(path)
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
		if len(tSubDirs) > treelevel {
			newRef := &fileref.Ref{}
			newRef.Type = fileref.DIRECTORY
			newRef.AllocationID = dirRef.AllocationID
			newRef.Path = "/" + strings.Join(tSubDirs[:treelevel+1], "/")
			newRef.Name = tSubDirs[treelevel]
			//dirRef.Children = append(dirRef.Children, newRef)
			dirRef.AddChild(newRef)
			dirRef = newRef
			treelevel++
			continue
		} else {
			break
		}
	}
	//dirRef.Children = append(dirRef.Children, ch.File)
	dirRef.AddChild(ch.File)
	rootRef.CalculateHash()
	return nil
}

func (n *NewFileChange) GetAffectedPath() string {
	if n.File != nil {
		return n.File.Path
	}
	return ""
}

func (n *NewFileChange) GetSize() int64 {
	if n.File != nil {
		return n.File.Size
	}
	return int64(0)
}
