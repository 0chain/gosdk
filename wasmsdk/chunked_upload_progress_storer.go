//go:build js && wasm
// +build js,wasm

package main

import (
	"sync"

	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

// chunkedUploadProgressStorer load and save upload progress
type chunkedUploadProgressStorer struct {
	list map[string]*sdk.UploadProgress
	lock sync.Mutex
}

// Load load upload progress by id
func (mem *chunkedUploadProgressStorer) Load(id string) *sdk.UploadProgress {
	mem.lock.Lock()
	defer mem.lock.Unlock()
	if mem.list == nil {
		mem.list = make(map[string]*sdk.UploadProgress)
		return nil
	}
	up, ok := mem.list[id]

	if ok {
		return up
	}

	return nil
}

// Save save upload progress
func (mem *chunkedUploadProgressStorer) Save(up sdk.UploadProgress) {
	mem.lock.Lock()
	defer mem.lock.Unlock()
	if mem.list == nil {
		mem.list = make(map[string]*sdk.UploadProgress)
	}
	mem.list[up.ID] = &up
}

//nolint:golint,unused
func (mem *chunkedUploadProgressStorer) Update(id string, chunkIndex int, _ zboxutil.Uint128) {
	mem.lock.Lock()
	defer mem.lock.Unlock()
	if mem.list == nil {
		return
	}
	up, ok := mem.list[id]
	if ok {
		up.ChunkIndex = chunkIndex
	}
}

// Remove remove upload progress by id
func (mem *chunkedUploadProgressStorer) Remove(id string) error {
	mem.lock.Lock()
	defer mem.lock.Unlock()
	delete(mem.list, id)
	return nil
}
