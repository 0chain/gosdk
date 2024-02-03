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
	"github.com/hitenjain14/fasthttp"
)

type DownloadProgressStorer interface {
	// Load load download progress by id
	Load(id string) *DownloadProgress
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

func CreateFsDownloadProgress() *FsDownloadProgressStorer {
	down := &FsDownloadProgressStorer{
		queue: make(queue, 0),
	}
	heap.Init(&down.queue)
	return down
}

func (ds *FsDownloadProgressStorer) Start(ctx context.Context) {
	ds.next += ds.dp.numBlocks
	tc := fasthttp.AcquireTimer(2 * time.Second)
	defer fasthttp.ReleaseTimer(tc)
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
}

func (fs *FsDownloadProgressStorer) Load(progressID string) *DownloadProgress {
	dp := &DownloadProgress{}
	buf, err := sys.Files.ReadFile(progressID)
	if err != nil {
		return nil
	}
	if err = json.Unmarshal(buf, dp); err != nil {
		return nil
	}
	fs.dp = dp
	return fs.dp
}

func (fs *FsDownloadProgressStorer) saveToDisk() {
	fs.Lock()
	defer fs.Unlock()
	if fs.isRemoved {
		return
	}
	buf, err := json.Marshal(fs.dp)
	if err != nil {
		return
	}
	err = sys.Files.WriteFile(fs.dp.ID, buf, 0666)
	if err != nil {
		return
	}
}

func (fs *FsDownloadProgressStorer) Save(dp *DownloadProgress) {
	fs.dp = dp
	fs.saveToDisk()
}

func (fs *FsDownloadProgressStorer) Update(writtenBlock int) {
	fs.Lock()
	defer fs.Unlock()
	if fs.isRemoved {
		return
	}
	heap.Push(&fs.queue, writtenBlock)
}

func (fs *FsDownloadProgressStorer) Remove() error {
	fs.Lock()
	defer fs.Unlock()
	if fs.isRemoved || fs.dp == nil {
		return nil
	}
	fs.isRemoved = true
	err := sys.Files.Remove(fs.dp.ID)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
