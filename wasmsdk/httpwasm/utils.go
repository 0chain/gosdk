package httpwasm

import (
	"strconv"

	"github.com/0chain/gosdk/core/encryption"
)

func GetMockId(allocation int) string {
	return encryption.Hash("random id" + strconv.Itoa(allocation))
}

func GetMockAllocationId(allocation int) string {
	return encryption.Hash("mock allocation id" + strconv.Itoa(allocation))
}

func GetMockBlobberId(index int) string {
	return encryption.Hash("mockBlobber_" + strconv.Itoa(index))
}
