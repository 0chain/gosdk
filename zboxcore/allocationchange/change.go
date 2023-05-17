package allocationchange

import (
	"github.com/0chain/gosdk/zboxcore/fileref"
)

type change struct {
	Size      int64  `json:"size"`
	NumBlocks int64  `json:"num_of_blocks"`
	Operation string `json:"operation"`
}

type AllocationChange interface {
	ProcessChange(rootRef *fileref.Ref, fileIDMeta map[string]string) error
	GetAffectedPath() []string
	GetSize() int64
}
