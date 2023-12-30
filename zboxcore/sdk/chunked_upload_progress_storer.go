package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/logger"
)

// ChunkedUploadProgressStorer load and save upload progress
type ChunkedUploadProgressStorer interface {
	// Load load upload progress by id
	Load(id string) *UploadProgress
	// Save save upload progress
	Save(up UploadProgress)
	// Remove remove upload progress by id
	Remove(id string) error
}

// fsChunkedUploadProgressStorer load and save upload progress in file system
type fsChunkedUploadProgressStorer struct {
	sync.Mutex
	isRemoved bool
	up        UploadProgress
	since     time.Time
}

func createFsChunkedUploadProgress(ctx context.Context) *fsChunkedUploadProgressStorer {
	up := &fsChunkedUploadProgressStorer{
		since: time.Now(),
	}

	return up
}

// Load load upload progress from file system
func (fs *fsChunkedUploadProgressStorer) Load(progressID string) *UploadProgress {

	progress := UploadProgress{}
	buf, err := sys.Files.ReadFile(progressID)

	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err := json.Unmarshal(buf, &progress); err != nil {
		return nil
	}

	return &progress
}

// Save save upload progress in file system
func (fs *fsChunkedUploadProgressStorer) Save(up UploadProgress) {
	fs.Lock()
	defer fs.Unlock()
	fs.up = up
	now := time.Now()
	if now.Sub(fs.since).Seconds() > 1 {
		if fs.isRemoved {
			return
		}

		buf, err := json.Marshal(fs.up)
		if err != nil {
			logger.Logger.Error("[progress] save ", fs.up, err)
			return
		}
		err = sys.Files.WriteFile(fs.up.ID, buf, 0666)
		if err != nil {
			logger.Logger.Error("[progress] save ", fs.up, err)
			return
		}

		fs.since = now
	}
}

// Remove remove upload progress from file system
func (fs *fsChunkedUploadProgressStorer) Remove(progressID string) error {
	fs.Lock()
	defer fs.Unlock()
	fs.isRemoved = true
	err := sys.Files.Remove(progressID)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	return nil
}

type chunkedUploadProgressStorer struct {
	list map[string]*UploadProgress
}

// Load load upload progress by id
func (mem *chunkedUploadProgressStorer) Load(id string) *UploadProgress {
	if mem.list == nil {
		mem.list = make(map[string]*UploadProgress)
		return nil
	}
	up, ok := mem.list[id]

	if ok {
		return up
	}

	return nil
}

// Save save upload progress
func (mem *chunkedUploadProgressStorer) Save(up UploadProgress) {
	if mem.list == nil {
		mem.list = make(map[string]*UploadProgress)
	}
	mem.list[up.ID] = &up
}

// Remove remove upload progress by id
func (mem *chunkedUploadProgressStorer) Remove(id string) error {
	delete(mem.list, id)
	return nil
}
