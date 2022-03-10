package blobber

import (
	"sync"

	"github.com/0chain/gosdk/dev/blobber/model"
)

var referencePathResults = make(map[string]*model.ReferencePathResult)
var referencePathResultsMutex sync.Mutex

func MockReferencePathResult(allocationId string, rootRef *model.Ref) func() {
	result := model.BuildReferencePathResult(rootRef)
	referencePathResultsMutex.Lock()
	defer referencePathResultsMutex.Unlock()
	referencePathResults[allocationId] = result

	return func() {
		referencePathResultsMutex.Lock()
		defer referencePathResultsMutex.Unlock()
		delete(referencePathResults, allocationId)
	}
}
