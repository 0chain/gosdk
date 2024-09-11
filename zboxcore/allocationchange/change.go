package allocationchange

import (
	"github.com/0chain/common/core/util/wmpt"
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

type AllocationChangeV2 interface {
	ProcessChangeV2(trie *wmpt.WeightedMerkleTrie, changeIndex uint64) error
	GetLookupHash(changeIndex uint64) string
	GetHash(changeIndex uint64, id string) string
}

type EmptyFileChange struct {
	change
}

func (ch *EmptyFileChange) ProcessChange(rootRef *fileref.Ref, _ map[string]string) (err error) {
	return
}

func (ch *EmptyFileChange) GetAffectedPath() []string {
	return []string{}
}

func (ch *EmptyFileChange) GetSize() int64 {
	return 0
}
