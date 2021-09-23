package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"time"

	"github.com/0chain/gosdk/zboxcore/logger"
)

// ChunkedUploadProgressStorer load and save upload progress
type ChunkedUploadProgressStorer interface {
	// Load load upload progress by id
	Load(id string) *UploadProgress
	// Save save upload progress
	Save(up *UploadProgress)
	// Remove remove upload progress by id
	Remove(id string) error
}

// fsChunkedUploadProgressStorer load and save upload progress in file system
type fsChunkedUploadProgressStorer struct {
	isRemoved bool
	up        *UploadProgress
}

func createFsChunkedUploadProgress(ctx context.Context) *fsChunkedUploadProgressStorer {
	up := &fsChunkedUploadProgressStorer{}

	go up.start()

	return up
}

func (fs *fsChunkedUploadProgressStorer) start() {

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {

		<-ticker.C

		if fs.up == nil {
			continue
		}

		if fs.isRemoved {
			break
		}

		buf, err := json.Marshal(fs.up)
		if err != nil {
			logger.Logger.Error("[progress] save ", fs.up, err)
			continue
		}

		err = ioutil.WriteFile(fs.up.ID, buf, 0644)
		if err != nil {
			logger.Logger.Error("[progress] save ", fs.up, err)
			continue
		}

	}
}

// Load load upload progress from file system
func (fs *fsChunkedUploadProgressStorer) Load(progressID string) *UploadProgress {

	progress := UploadProgress{}

	buf, err := ioutil.ReadFile(progressID)

	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err := json.Unmarshal(buf, &progress); err != nil {
		return nil
	}

	return &progress
}

// Save save upload progress in file system
func (fs *fsChunkedUploadProgressStorer) Save(up *UploadProgress) {
	fs.up = up
}

// Remove remove upload progress from file system
func (fs *fsChunkedUploadProgressStorer) Remove(progressID string) error {
	fs.isRemoved = true
	err := os.Remove(progressID)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	return nil
}
