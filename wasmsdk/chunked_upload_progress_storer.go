package main

import "github.com/0chain/gosdk/zboxcore/sdk"

// chunkedUploadProgressStorer load and save upload progress
type chunkedUploadProgressStorer struct {
	list map[string]*sdk.UploadProgress
}

// Load load upload progress by id
func (mem *chunkedUploadProgressStorer) Load(id string) *sdk.UploadProgress {
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
	if mem.list == nil {
		mem.list = make(map[string]*sdk.UploadProgress)
	}
	mem.list[up.ID] = &up
}

// Remove remove upload progress by id
func (mem *chunkedUploadProgressStorer) Remove(id string) error {
	delete(mem.list, id)
	return nil
}
