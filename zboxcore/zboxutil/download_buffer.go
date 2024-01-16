package zboxutil

import (
	"context"
	"sync"
	"time"
)

type DownloadBuffer struct {
	buf     []byte
	length  int
	reqSize int
	mask    uint32
	mu      sync.RWMutex
}

func NewDownloadBuffer(size, numBlocks, effectiveBlockSize int) *DownloadBuffer {
	numBlocks++
	return &DownloadBuffer{
		buf:     make([]byte, size*numBlocks*effectiveBlockSize),
		length:  size,
		reqSize: effectiveBlockSize * numBlocks,
		mask:    (1 << size) - 1,
	}
}

func (r *DownloadBuffer) RequestChunk(ctx context.Context, num int) []byte {
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

func (r *DownloadBuffer) ReleaseChunk(num int) {
	num = num % r.length
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mask |= 1 << num
}

func (r *DownloadBuffer) Stats() (int, int) {
	return len(r.buf), cap(r.buf)
}
