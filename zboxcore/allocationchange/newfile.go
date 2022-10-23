package allocationchange

import (
	"path/filepath"
	"strings"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/marker"
)

type NewFileChange struct {
	change
	File *fileref.FileRef
}

func (ch *NewFileChange) ProcessChange(
	rootRef *fileref.Ref, latestFileID int64) (
	commitParams CommitParams, err error) {

	inodesMeta := make(map[string]int64)
	tSubDirs, err := common.GetPathFields(filepath.Dir(ch.File.Path))
	if err != nil {
		return
	}

	dirRef := rootRef
	for i := 0; i < len(tSubDirs); i++ {
		found := false
		for _, child := range dirRef.Children {
			if child.GetType() == fileref.DIRECTORY && child.(*fileref.Ref).Name == tSubDirs[i] {
				dirRef = child.(*fileref.Ref)
				found = true
				break
			}
		}
		if !found {
			latestFileID++
			newRef := &fileref.Ref{
				Type:         fileref.DIRECTORY,
				AllocationID: dirRef.AllocationID,
				Path:         filepath.Join("/", strings.Join(tSubDirs[:i+1], "/")),
				Name:         tSubDirs[i],
				FileID:       latestFileID,
			}
			inodesMeta[newRef.Path] = latestFileID
			dirRef.AddChild(newRef)
			dirRef = newRef
		}
	}
	latestFileID++
	inodesMeta[ch.File.GetPath()] = latestFileID

	ch.File.FileID = latestFileID
	commitParams.WmFileID = latestFileID
	commitParams.LatestFileID = latestFileID
	commitParams.Operation = marker.Upload
	commitParams.InodesMeta = inodesMeta
	dirRef.AddChild(ch.File)
	rootRef.CalculateHash()
	return
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
