package zbox

import (
	"sort"
	"testing"

	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/influxdata/influxdb/pkg/testing/assert"
)

func TestListSorting(t *testing.T) {
	list := make([]sdk.ListResult, 0)
	list = append(list, sdk.ListResult{Name: "test4.ts"})
	list = append(list, sdk.ListResult{Name: "test2.ts"})
	list = append(list, sdk.ListResult{Name: "test3.ts"})
	list = append(list, sdk.ListResult{Name: "test1.ts"})
	list = append(list, sdk.ListResult{Name: "test5.ts"})
	list = append(list, sdk.ListResult{Name: "test6.ts"})
	list = append(list, sdk.ListResult{Name: "test10.ts"})
	list = append(list, sdk.ListResult{Name: "test15.ts"})
	list = append(list, sdk.ListResult{Name: "test14.ts"})
	list = append(list, sdk.ListResult{Name: "test42.ts"})

	sort.Slice(list, func(i, j int) bool {
		return GetNumber(list[i].Name) < GetNumber(list[j].Name)
	})

	initId := sort.Search(len(list), func(i int) bool {
		return list[i].Name > "test5"
	})
	assert.Equal(t, initId, 4)

	found := sort.Search(len(list), func(i int) bool {
		return list[i].Name == "test122"
	})

	assert.Equal(t, found, len(list))

}
