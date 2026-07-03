//go:build !js && !wasm
// +build !js,!wasm

package sdk

// HashVersion 2 parallel producer.
//
// Implements zs3 docs/PARALLEL_PRODUCER_TREE_HASH_DESIGN.md: the per-file
// producer serial dependency was the streaming MD5 (ActualHash over the file,
// DataHash over each shard) — bytes had to be hashed in order, so read +
// erasure-encode + hash + form-build ran on ONE goroutine per file (~150
// MiB/s/file; measured p50 132ms read+encode + 49ms processUpload per round
// while the 4 upload workers idled 1.4s p50). With the segmented tree hash
// (TreeHasher) leaves are computable per chunk in ANY order, so:
//
//   - a look-ahead goroutine reads round N+1 into its OWN pooled buffer while
//     round N is being processed (the May-29 read-ahead failed on buffer
//     reuse + arena copies; per-round DetachBuffer removes both — no copy,
//     no shared buffer),
//   - producerWorkers goroutines erasure-encode + md5 the round's chunks in
//     parallel (klauspost/reedsolomon Encode is goroutine-safe per call;
//     chunk regions of the round buffer are disjoint),
//   - the serial loop folds the leaf digests in chunk order (16 bytes/leaf —
//     negligible) and hands the round to the existing processUpload /
//     uploadChan machinery, so upload ordering, progress tracking, the final
//     drain and the commit protocol (tmp/ staging + write-marker) are
//     UNCHANGED.

import (
	"crypto/md5"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/zboxcore/logger"
)

type roundData struct {
	chunks *batchChunksData
	err    error
	readMs int64
}

// processParallel is the HashVersion-2 replacement for the serial loop in
// process(). Rounds are still CONSUMED in order; only the read of the next
// round and the per-chunk encode+hash within a round run in parallel.
func (su *ChunkedUpload) processParallel() error {
	roundChan := make(chan roundData, 1)
	go func() {
		defer close(roundChan)
		for {
			tRead := time.Now()
			chunks, err := su.readChunks(su.chunkNumber)
			rd := roundData{chunks: chunks, err: err, readMs: time.Since(tRead).Milliseconds()}
			select {
			case roundChan <- rd:
			case <-su.ctx.Done():
				if chunks != nil {
					releaseRoundBuffer(chunks.buf)
				}
				return
			}
			if err != nil || chunks.isFinal {
				return
			}
		}
	}()

	for rd := range roundChan {
		if rd.err != nil {
			if su.statusCallback != nil {
				su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.RemotePath, su.opCode, rd.err)
			}
			return rd.err
		}
		chunks := rd.chunks

		tWait := time.Now()
		tStage := time.Now()
		err := su.stageEncodeAndHash(chunks)
		stageMs := time.Since(tStage).Milliseconds()
		if err != nil {
			releaseRoundBuffer(chunks.buf)
			if su.statusCallback != nil {
				su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.RemotePath, su.opCode, err)
			}
			return err
		}

		su.shardUploadedSize += chunks.totalFragmentSize
		su.progress.ReadLength += chunks.totalReadSize

		if chunks.isFinal {
			// All leaves folded (rounds are consumed in order) — root ready.
			if su.fileMeta.ActualHash == "" {
				su.fileMeta.ActualHash, err = su.chunkReader.GetFileHash()
				if err != nil {
					releaseRoundBuffer(chunks.buf)
					if su.statusCallback != nil {
						su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.RemotePath, su.opCode, err)
					}
					return err
				}
			}
			if su.fileMeta.ActualSize == 0 {
				su.fileMeta.ActualSize = su.progress.ReadLength
				su.shardSize = getShardSize(su.fileMeta.ActualSize, su.allocationObj.DataShards, su.encryptOnUpload)
			} else if su.fileMeta.ActualSize != su.progress.ReadLength && su.thumbnailBytes == nil {
				releaseRoundBuffer(chunks.buf)
				err = thrown.New("upload_failed", "Upload failed. Uploaded size does not match with actual size: "+fmt.Sprintf("%d != %d", su.fileMeta.ActualSize, su.progress.ReadLength))
				if su.statusCallback != nil {
					su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.RemotePath, su.opCode, err)
				}
				return err
			}
		}

		tProc := time.Now()
		err = su.processUpload(
			chunks.chunkStartIndex, chunks.chunkEndIndex,
			chunks.fileShards, chunks.thumbnailShards,
			chunks.isFinal, chunks.totalReadSize,
		)
		procMs := time.Since(tProc).Milliseconds()
		// Forms are built (bodies own copies of the fragment bytes) — the
		// round buffer can go back to the pool even on error.
		releaseRoundBuffer(chunks.buf)
		if err != nil {
			if su.statusCallback != nil {
				su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.RemotePath, su.opCode, err)
			}
			return err
		}

		// Same [batch-timing] marker as the v1 loop so before/after p50s
		// come from the same grep. read+encode = the look-ahead read of THIS
		// round (overlapped with the previous round's stage+upload);
		// stage = parallel encode+hash; roundtotal = serial cost actually
		// paid by this loop for the round.
		logger.Logger.Info(fmt.Sprintf("[batch-timing] conn=%s chunkStart=%d chunkEnd=%d bytes=%d isFinal=%v read+encode=%dms stage=%dms processUpload=%dms roundtotal=%dms hashv=2",
			su.progress.ConnectionID, chunks.chunkStartIndex, chunks.chunkEndIndex, chunks.totalReadSize, chunks.isFinal, rd.readMs, stageMs, procMs, time.Since(tWait).Milliseconds()))

		if chunks.isFinal {
			break
		}
	}
	return nil
}

// stageEncodeAndHash erasure-encodes and leaf-hashes all chunks of a round in
// parallel, then folds the digests into the tree hashers in chunk order.
func (su *ChunkedUpload) stageEncodeAndHash(data *batchChunksData) error {
	n := len(data.chunks)
	if n == 0 {
		return nil
	}
	numShards := len(su.blobbers)

	fileLeaves := make([][md5.Size]byte, n)
	blockLeaves := make([][][md5.Size]byte, numShards)
	for b := 0; b < numShards; b++ {
		blockLeaves[b] = make([][md5.Size]byte, n)
	}

	hashChunk := func(i int) error {
		c := data.chunks[i]
		if err := su.fileErasureEncoder.Encode(c.Fragments); err != nil {
			return err
		}
		fileLeaves[i] = md5.Sum(c.RawData)
		for b, frag := range c.Fragments {
			blockLeaves[b][i] = md5.Sum(frag)
		}
		return nil
	}

	workers := producerWorkers
	if workers > n {
		workers = n
	}
	if workers <= 1 {
		for i := 0; i < n; i++ {
			if err := hashChunk(i); err != nil {
				return err
			}
		}
	} else {
		var (
			next  atomic.Int64
			wg    sync.WaitGroup
			errCh = make(chan error, workers)
		)
		next.Store(-1)
		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					i := int(next.Add(1))
					if i >= n {
						return
					}
					if err := hashChunk(i); err != nil {
						select {
						case errCh <- err:
						default:
						}
						return
					}
				}
			}()
		}
		wg.Wait()
		close(errCh)
		for err := range errCh {
			return err
		}
	}

	// Serial fold in chunk order — 16 bytes per leaf.
	fh, ok := su.fileHasher.(*TreeHasher)
	if !ok {
		return thrown.New("tree_hasher", "hashVersion 2 requires a TreeHasher file hasher")
	}
	for i := 0; i < n; i++ {
		fh.FoldLeafDigest(fileLeaves[i][:])
	}
	for b := 0; b < numShards; b++ {
		bh, ok := su.progress.Blobbers[b].Hasher.(*TreeHasher)
		if !ok {
			return thrown.New("tree_hasher", "hashVersion 2 requires TreeHasher blobber hashers")
		}
		for i := 0; i < n; i++ {
			bh.FoldLeafDigest(blockLeaves[b][i][:])
		}
	}
	return nil
}
