package zboxutil

import (
	"context"
	"sync"
	"time"
)

type DownloadBuffer interface {
	RequestChunk(ctx context.Context, num int) []byte
	ReleaseChunk(num int)
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
		return r.buf[ind*r.reqSize : (ind+1)*r.reqSize]
	}
}

type DownloadBufferWithMask struct {
	buf     []byte
	length  int
	reqSize int
	mask    uint32
	mu      sync.RWMutex
}

func NewDownloadBufferWithMask(size, numBlocks, effectiveBlockSize int) *DownloadBufferWithMask {
	numBlocks++
	return &DownloadBufferWithMask{
		buf:     make([]byte, size*numBlocks*effectiveBlockSize),
		length:  size,
		reqSize: effectiveBlockSize * numBlocks,
		mask:    (1 << size) - 1,
	}
}

func (r *DownloadBufferWithMask) RequestChunk(ctx context.Context, num int) []byte {
	num = num % r.length
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		r.mu.RLock()
		isSet := r.mask & (1 << num)
		r.mu.RUnlock()
		// already assigned
		if isSet == 0 {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		// assign the chunk by clearing the bit
		r.mu.Lock()
		r.mask &= ^(1 << num)
		r.mu.Unlock()
		return r.buf[num*r.reqSize : (num+1)*r.reqSize]
	}
}

func (r *DownloadBufferWithMask) ReleaseChunk(num int) {
	num = num % r.length
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mask |= 1 << num
}
