package zboxutil

import (
	"context"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/0chain/gosdk/core/sys"
	"github.com/valyala/bytebufferpool"
)

// dlBufPollInterval is how long RequestChunk sleeps between retries when its
// ring slot is momentarily busy. The historical 200ms is a severe throughput
// bug: every block whose slot is in use stalls ~200ms even with CPU/NIC/disk
// idle, capping NFS/S3 download throughput to ~(concurrency * blockBytes /
// 200ms). Tunable via ZUS_DL_BUF_POLL_MS (default 200 = legacy; set 1 to fix).
var dlBufPollInterval = func() time.Duration {
	if v := os.Getenv("ZUS_DL_BUF_POLL_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n) * time.Millisecond
		}
	}
	return 200 * time.Millisecond
}()

type DownloadBuffer interface {
	RequestChunk(ctx context.Context, num int) []byte
	ReleaseChunk(num int)
	ClearBuffer()
}

type DownloadBufferWithChan struct {
	buf     []byte
	length  int
	reqSize int
	ch      chan int
	mu      sync.Mutex
	mp      map[int]int
}

func NewDownloadBufferWithChan(size, numBlocks, effectiveBlockSize int) *DownloadBufferWithChan {
	numBlocks++
	db := &DownloadBufferWithChan{
		buf:     make([]byte, size*numBlocks*effectiveBlockSize),
		length:  size,
		reqSize: effectiveBlockSize * numBlocks,
		ch:      make(chan int, size),
		mp:      make(map[int]int),
	}
	for i := 0; i < size; i++ {
		db.ch <- i
	}
	return db
}

func (r *DownloadBufferWithChan) ReleaseChunk(num int) {
	r.mu.Lock()
	ind, ok := r.mp[num]
	if !ok {
		r.mu.Unlock()
		return
	}
	delete(r.mp, num)
	r.mu.Unlock()
	r.ch <- ind
}

func (r *DownloadBufferWithChan) RequestChunk(ctx context.Context, num int) []byte {
	select {
	case <-ctx.Done():
		return nil
	case ind := <-r.ch:
		r.mu.Lock()
		r.mp[num] = ind
		r.mu.Unlock()
		return r.buf[ind*r.reqSize : (ind+1)*r.reqSize : (ind+1)*r.reqSize]
	}
}

func (r *DownloadBufferWithChan) ClearBuffer() {
	r.buf = nil
	close(r.ch)
	for k := range r.mp {
		delete(r.mp, k)
	}
	r.mp = nil
}

type DownloadBufferWithMask struct {
	downloadBuf []*bytebufferpool.ByteBuffer
	length      int
	reqSize     int
	numBlocks   int
	mask        uint32
	mu          sync.Mutex
}

func NewDownloadBufferWithMask(size, numBlocks, effectiveBlockSize int) *DownloadBufferWithMask {
	numBlocks++
	return &DownloadBufferWithMask{
		length:      size,
		reqSize:     effectiveBlockSize * numBlocks,
		mask:        (1 << size) - 1,
		downloadBuf: make([]*bytebufferpool.ByteBuffer, size),
	}
}

func (r *DownloadBufferWithMask) SetNumBlocks(numBlocks int) {
	r.numBlocks = numBlocks
}

func (r *DownloadBufferWithMask) RequestChunk(ctx context.Context, num int) []byte {
	num = num / r.numBlocks
	num = num % r.length
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		r.mu.Lock()
		isSet := r.mask & (1 << num)
		// already assigned
		if isSet == 0 {
			r.mu.Unlock()
			sys.Sleep(dlBufPollInterval)
			continue
		}
		// assign the chunk by clearing the bit
		r.mask &= ^(1 << num)
		if r.downloadBuf[num] == nil {
			buff := BufferPool.Get()
			if cap(buff.B) < r.reqSize {
				buff.B = make([]byte, r.reqSize)
			}
			r.downloadBuf[num] = buff
		}
		r.mu.Unlock()
		return r.downloadBuf[num].B[:r.reqSize:r.reqSize]
	}
}

func (r *DownloadBufferWithMask) ReleaseChunk(num int) {
	num = num / r.numBlocks
	num = num % r.length
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mask |= 1 << num
}

func (r *DownloadBufferWithMask) ClearBuffer() {
	for _, buff := range r.downloadBuf {
		if buff != nil {
			BufferPool.Put(buff)
		}
	}
	r.downloadBuf = nil
}
