package sdk

import (
	"container/heap"
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

// ChunkedUploadProgressStorer load and save upload progress
type ChunkedUploadProgressStorer interface {
	// Load load upload progress by id
	Load(id string) *UploadProgress
	// Save save upload progress
	Save(up UploadProgress)
	// Remove remove upload progress by id
	Remove(id string) error
	// Update update upload progress
	Update(id string, chunkIndex int, upMask zboxutil.Uint128)
}

// fsChunkedUploadProgressStorer load and save upload progress in file system
type fsChunkedUploadProgressStorer struct {
	sync.Mutex
	isRemoved  bool
	up         UploadProgress
	queue      queue
	next       int
	uploadMask zboxutil.Uint128
}

type queue []int

func (pq queue) Len() int { return len(pq) }

func (pq queue) Less(i, j int) bool {
	return pq[i] < pq[j]
}

func (pq queue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *queue) Push(x interface{}) {
	*pq = append(*pq, x.(int))
}

func (pq *queue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

func createFsChunkedUploadProgress(ctx context.Context) *fsChunkedUploadProgressStorer {
	up := &fsChunkedUploadProgressStorer{
		queue: make(queue, 0),
	}
	heap.Init(&up.queue)
	go saveProgress(ctx, up)
	return up
}

func saveProgress(ctx context.Context, fs *fsChunkedUploadProgressStorer) {
	tc := time.NewTicker(2 * time.Second)
	defer tc.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tc.C:
			fs.Lock()
			if fs.isRemoved {
				fs.Unlock()
				return
			}
			if len(fs.queue) > 0 && fs.next == fs.queue[0] {
				for len(fs.queue) > 0 && fs.next == fs.queue[0] {
					fs.up.ChunkIndex = fs.queue[0]
					heap.Pop(&fs.queue)
					fs.next += fs.up.ChunkNumber
				}
				fs.Unlock()
				fs.Save(fs.up)
			} else {
				fs.Unlock()
			}
		}
	}
}

// Load load upload progress from file system
func (fs *fsChunkedUploadProgressStorer) Load(progressID string) *UploadProgress {

	progress := UploadProgress{}
	buf, err := sys.Files.LoadProgress(progressID)

	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err := json.Unmarshal(buf, &progress); err != nil {
		return nil
	}

	// if progress is not updated within 25 min, return nil
	if !progress.LastUpdated.Within(25 * 60) {
		sys.Files.Remove(progressID) //nolint:errcheck
		return nil
	}

	return &progress
}

// Save save upload progress in file system
func (fs *fsChunkedUploadProgressStorer) Save(up UploadProgress) {
	fs.Lock()
	defer fs.Unlock()
	fs.up = up
	fs.up.LastUpdated = common.Now()

	if fs.isRemoved {
		return
	}
	if fs.next == 0 {
		if up.ChunkNumber == -1 {
			fs.next = up.ChunkNumber - 1
		} else {
			fs.next = up.ChunkIndex + up.ChunkNumber
		}
	}
	fs.up.UploadMask = fs.uploadMask
	buf, err := json.Marshal(fs.up)
	if err != nil {
		logger.Logger.Error("[progress] save ", fs.up, err)
		return
	}
	err = sys.Files.SaveProgress(fs.up.ID, buf, 0666)
	if err != nil {
		logger.Logger.Error("[progress] save ", fs.up, err)
		return
	}
}

func (fs *fsChunkedUploadProgressStorer) Update(id string, chunkIndex int, upMask zboxutil.Uint128) {
	fs.Lock()
	defer fs.Unlock()
	if !fs.isRemoved {
		fs.uploadMask = upMask
		heap.Push(&fs.queue, chunkIndex)
	}
}

// Remove remove upload progress from file system
func (fs *fsChunkedUploadProgressStorer) Remove(progressID string) error {
	fs.Lock()
	defer fs.Unlock()
	fs.isRemoved = true
	err := sys.Files.RemoveProgress(progressID)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	return nil
}
