package sdk

import (
	"container/heap"
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/sys"
)

type DownloadProgressStorer interface {
	// Load load download progress by id
	Load(id string, numBlocks int) *DownloadProgress
	// Update download progress
	Update(writtenBlock int)
	// Remove remove download progress by id
	Remove() error
	// Start start download progress
	Start(ctx context.Context)
	// Save download progress
	Save(dp *DownloadProgress)
}

type FsDownloadProgressStorer struct {
	sync.Mutex
	isRemoved bool
	dp        *DownloadProgress
	next      int
	queue     queue
}

// CreateFsDownloadProgress create a download progress storer instance to track download progress and queue
func CreateFsDownloadProgress() *FsDownloadProgressStorer {
	down := &FsDownloadProgressStorer{
		queue: make(queue, 0),
	}
	heap.Init(&down.queue)
	return down
}

func (ds *FsDownloadProgressStorer) Start(ctx context.Context) {
	tc := time.NewTicker(2 * time.Second)
	ds.next += ds.dp.numBlocks
	go func() {
		defer tc.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tc.C:
				ds.Lock()
				if ds.isRemoved {
					ds.Unlock()
					return
				}
				if len(ds.queue) > 0 && ds.queue[0] == ds.next {
					for len(ds.queue) > 0 && ds.queue[0] == ds.next {
						ds.dp.LastWrittenBlock = ds.next
						heap.Pop(&ds.queue)
						ds.next += ds.dp.numBlocks
					}
					ds.Unlock()
					ds.saveToDisk()
				} else {
					ds.Unlock()
				}
			}
		}
	}()
}

func (ds *FsDownloadProgressStorer) Load(progressID string, numBlocks int) *DownloadProgress {
	dp := &DownloadProgress{}
	buf, err := sys.Files.LoadProgress(progressID)
	if err != nil {
		return nil
	}
	if err = json.Unmarshal(buf, dp); err != nil {
		return nil
	}
	ds.dp = dp
	dp.numBlocks = numBlocks
	ds.next = dp.LastWrittenBlock
	return ds.dp
}

func (ds *FsDownloadProgressStorer) saveToDisk() {
	ds.Lock()
	defer ds.Unlock()
	if ds.isRemoved {
		return
	}
	buf, err := json.Marshal(ds.dp)
	if err != nil {
		return
	}
	err = sys.Files.SaveProgress(ds.dp.ID, buf, 0666)
	if err != nil {
		return
	}
}

func (ds *FsDownloadProgressStorer) Save(dp *DownloadProgress) {
	ds.dp = dp
	ds.saveToDisk()
}

func (ds *FsDownloadProgressStorer) Update(writtenBlock int) {
	ds.Lock()
	defer ds.Unlock()
	if ds.isRemoved {
		return
	}
	heap.Push(&ds.queue, writtenBlock)
}

func (ds *FsDownloadProgressStorer) Remove() error {
	ds.Lock()
	defer ds.Unlock()
	if ds.isRemoved || ds.dp == nil {
		return nil
	}
	ds.isRemoved = true
	err := sys.Files.RemoveProgress(ds.dp.ID)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
