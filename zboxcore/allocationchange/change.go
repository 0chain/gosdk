package allocationchange

import (
	"strings"

	"github.com/0chain/gosdk/zboxcore/fileref"
)

type change struct {
	Size      int64  `json:"size"`
	NumBlocks int64  `json:"num_of_blocks"`
	Operation string `json:"operation"`
}

type AllocationChange interface {
	ProcessChange(rootRef *fileref.Ref) error
	GetAffectedPath() string
	GetSize() int64
}

func getSubDirs(p string) []string {
	subDirs := strings.Split(p, "/")
	tSubDirs := make([]string, 0)
	for _, s := range subDirs {
		if s != "" {
			tSubDirs = append(tSubDirs, s)
		}
	}
	return tSubDirs
}
