//go:build !js && !wasm
// +build !js,!wasm

package sdk

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"testing"

	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/klauspost/reedsolomon"
	"github.com/stretchr/testify/require"
)

// refSegmentedRoot is the REFERENCE implementation of the HashVersion-2
// construction, written the way the BLOBBER computes it: split the byte
// stream at every leafSize bytes, md5 each segment, root = md5 of the
// concatenated 16-byte digests. Client-side folds must produce the same
// value for the same stream.
func refSegmentedRoot(data []byte, leafSize int) string {
	root := md5.New()
	for off := 0; off < len(data); off += leafSize {
		end := off + leafSize
		if end > len(data) {
			end = len(data)
		}
		d := md5.Sum(data[off:end])
		root.Write(d[:])
	}
	return hex.EncodeToString(root.Sum(nil))
}

func TestTreeHasherMatchesReference(t *testing.T) {
	leafSize := 64 * 1024
	for _, size := range []int{0, 1, 1000, leafSize, leafSize + 1, 3*leafSize - 17, 8 * leafSize} {
		data := make([]byte, size)
		rand.New(rand.NewSource(int64(size))).Read(data)

		th := CreateTreeHasher()
		for off := 0; off < len(data); off += leafSize {
			end := off + leafSize
			if end > len(data) {
				end = len(data)
			}
			d := md5.Sum(data[off:end])
			th.FoldLeafDigest(d[:])
		}
		got, err := th.GetBlockHash()
		require.NoError(t, err)
		require.Equal(t, refSegmentedRoot(data, leafSize), got, "size=%d", size)
	}
}

// buildReader constructs a chunk reader over data for the given hash version.
func buildTestReader(t *testing.T, data []byte, dataShards, parityShards, chunkNumber, hashVersion int) (ChunkedUploadChunkReader, Hasher) {
	encoder, err := reedsolomon.New(dataShards, parityShards, reedsolomon.WithAutoGoroutines(DefaultChunkSize))
	require.NoError(t, err)
	var h Hasher
	if hashVersion == 2 {
		h = CreateTreeHasher()
	} else {
		h = CreateFileHasher()
	}
	mask := zboxutil.NewUint128(1).Lsh(uint64(dataShards + parityShards)).Sub64(1)
	r, err := createChunkReader(bytes.NewReader(data), int64(len(data)), DefaultChunkSize,
		dataShards, parityShards, false, mask, encoder, nil, h, chunkNumber, hashVersion)
	require.NoError(t, err)
	return r, h
}

// TestV2FragmentsByteIdenticalToV1 is the byte-identical gate: the v2 reader
// (Split in Next, Encode deferred to the stage) must produce EXACTLY the same
// fragment bytes as the v1 serial reader for every shard, including the
// padded final chunk.
func TestV2FragmentsByteIdenticalToV1(t *testing.T) {
	const dataShards, parityShards, chunkNumber = 2, 1, 4
	fullChunk := DefaultChunkSize * dataShards
	encoder, err := reedsolomon.New(dataShards, parityShards, reedsolomon.WithAutoGoroutines(DefaultChunkSize))
	require.NoError(t, err)

	for _, size := range []int{1000, fullChunk, fullChunk + 1, 3*fullChunk - 4099, chunkNumber*fullChunk + 5} {
		data := make([]byte, size)
		rand.New(rand.NewSource(int64(size))).Read(data)

		collect := func(hashVersion int) [][][]byte { // [chunk][shard]bytes
			r, _ := buildTestReader(t, data, dataShards, parityShards, chunkNumber, hashVersion)
			var out [][][]byte
			done := false
			for !done { // one iteration per upload round, as readChunks does
				for i := 0; i < chunkNumber; i++ {
					c, err := r.Next()
					require.NoError(t, err)
					if c.ReadSize > 0 {
						if hashVersion == 2 {
							// simulate the parallel stage: Encode fills parity
							require.NoError(t, encoder.Encode(c.Fragments))
						}
						frags := make([][]byte, len(c.Fragments))
						for fi, f := range c.Fragments {
							frags[fi] = append([]byte(nil), f...) // own the bytes before buffer reuse
						}
						out = append(out, frags)
					}
					if c.IsFinal {
						done = true
						break
					}
				}
				if hashVersion == 2 {
					releaseRoundBuffer(r.DetachBuffer())
				}
				r.Reset()
			}
			r.Close()
			return out
		}

		v1 := collect(1)
		v2 := collect(2)
		require.Equal(t, len(v1), len(v2), "chunk count size=%d", size)
		for ci := range v1 {
			for si := range v1[ci] {
				require.True(t, bytes.Equal(v1[ci][si], v2[ci][si]),
					"fragment mismatch size=%d chunk=%d shard=%d", size, ci, si)
			}
		}
	}
}

// TestParallelStageRoots runs the REAL parallel stage (stageEncodeAndHash,
// producerWorkers goroutines) over a multi-round upload and checks:
//   - per-shard DataHash root == blobber-style segmented md5 of the
//     concatenated shard stream (leaf = ChunkSize bytes) — i.e. the value
//     eblobber's v2 CommitHasher recomputes from its temp file,
//   - file ActualHash root == segmented md5 of the raw file with
//     leaf = ChunkSize*DataShards.
func TestParallelStageRoots(t *testing.T) {
	const dataShards, parityShards, chunkNumber = 2, 1, 4
	numShards := dataShards + parityShards
	fullChunk := DefaultChunkSize * dataShards

	for _, size := range []int{1000, fullChunk, 2*chunkNumber*fullChunk + 777, chunkNumber * fullChunk} {
		data := make([]byte, size)
		rand.New(rand.NewSource(int64(size) + 42)).Read(data)

		encoder, err := reedsolomon.New(dataShards, parityShards, reedsolomon.WithAutoGoroutines(DefaultChunkSize))
		require.NoError(t, err)

		r, fileHasher := buildTestReader(t, data, dataShards, parityShards, chunkNumber, 2)
		su := &ChunkedUpload{
			fileErasureEncoder: encoder,
			fileHasher:         fileHasher,
			hashVersion:        2,
			chunkReader:        r,
			chunkNumber:        chunkNumber,
			blobbers:           make([]*ChunkedUploadBlobber, numShards),
		}
		su.progress.Blobbers = make([]*UploadBlobberStatus, numShards)
		for i := 0; i < numShards; i++ {
			su.progress.Blobbers[i] = &UploadBlobberStatus{Hasher: CreateTreeHasher()}
			su.blobbers[i] = &ChunkedUploadBlobber{progress: su.progress.Blobbers[i]}
		}

		// shard streams as the blobber would see them on disk
		shardStreams := make([][]byte, numShards)

		for {
			batch, err := su.readChunks(chunkNumber)
			require.NoError(t, err)
			require.NoError(t, su.stageEncodeAndHash(batch))
			for si, shard := range batch.fileShards {
				for _, frag := range shard {
					shardStreams[si] = append(shardStreams[si], frag...)
				}
			}
			releaseRoundBuffer(batch.buf)
			if batch.isFinal {
				break
			}
		}

		// file root
		gotFile, err := su.chunkReader.GetFileHash()
		require.NoError(t, err)
		require.Equal(t, refSegmentedRoot(data, fullChunk), gotFile, "file root size=%d", size)

		// per-shard DataHash roots vs blobber-side recompute
		for si := 0; si < numShards; si++ {
			got, err := su.progress.Blobbers[si].Hasher.GetBlockHash()
			require.NoError(t, err)
			require.Equal(t, refSegmentedRoot(shardStreams[si], DefaultChunkSize), got,
				"shard root size=%d shard=%d", size, si)
		}
	}
}
