package sdk

import (
	"io"
	"math"
	"strconv"
	"sync"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/zboxcore/encryption"
	"github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/klauspost/reedsolomon"
	"github.com/valyala/bytebufferpool"
)

var (
	uploadPool   bytebufferpool.Pool
	formDataPool bytebufferpool.Pool

	// ecShardsHeaderPool pools the [][]byte slice header produced by
	// erasureEncoder.Split on every chunk. Split always allocates
	// make([][]byte, totalShards) regardless of whether parity bytes need to
	// be freshly allocated (they don't when cap(data) >= needTotal, which is
	// always the case here because fileShardsDataBuffer carries parity
	// capacity). Pooling this header eliminates one GC-visible allocation per
	// chunk on the hot upload path.
	//
	// Get/Put discipline: Get is called at the top of Next() to reclaim the
	// slice used by the previous call (stored in prevFragments). The slice is
	// safe to reclaim because readChunks() fully iterates chunk.Fragments
	// before calling Next() again, and the individual []byte elements (which
	// are sub-slices of fileShardsDataBuffer, not of the header slice) are
	// appended into data.fileShards. The header slice itself is not retained
	// by any caller after Next() returns. Reset() and Release() also Put the
	// pending slice to prevent leaks on the final chunk.
	ecShardsHeaderPool sync.Pool
)

type ChunkedUploadChunkReader interface {
	// Next read, encode and encrypt next chunk
	Next() (*ChunkData, error)

	// Read read, encode and encrypt all bytes
	Read(buf []byte) ([][]byte, error)

	//Close Hash Channel
	Close()
	//GetFileHash get file hash
	GetFileHash() (string, error)
	//Reset reset offset
	Reset()
	//Release buffer
	Release()
}

// chunkedUploadChunkReader read chunk bytes from io.Reader. see detail on https://github.com/0chain/blobber/wiki/Protocols#what-is-fixedmerkletree
type chunkedUploadChunkReader struct {
	fileReader io.Reader

	//size total size of source. 0 means we don't it
	size int64
	// readSize total read size from source
	readSize int64

	// chunkSize chunk size with encryption header
	chunkSize int64

	// chunkHeaderSize encrypt header size
	chunkHeaderSize int64
	// chunkDataSize data size without encryption header in a chunk. It is same as ChunkSize if EncryptOnUpload is false
	chunkDataSize int64

	// chunkDataSizePerRead total size should be read from original io.Reader. It is DataSize * DataShards.
	chunkDataSizePerRead int64

	//totaChunkDataSizePerRead total size of data in a chunk. It is DataSize * (DataShards + ParityShards)
	totalChunkDataSizePerRead int64

	//fileShardsDataBuffer
	fileShardsDataBuffer *bytebufferpool.ByteBuffer

	//offset
	offset      int64
	chunkNumber int64

	// nextChunkIndex next index for reading
	nextChunkIndex int

	dataShards int

	// encryptOnUpload enccrypt data on upload
	encryptOnUpload bool

	uploadMask zboxutil.Uint128
	// erasureEncoder erasuer encoder
	erasureEncoder reedsolomon.Encoder
	// encscheme encryption scheme
	encscheme encryption.EncryptionScheme
	// hasher to calculate actual file hash, validation root and fixed merkle root
	hasher         Hasher
	hasherDataChan chan []byte
	hasherError    error
	hasherWG       sync.WaitGroup
	closeOnce      sync.Once

	// totalShards = dataShards + parityShards, cached to avoid recomputation.
	totalShards int

	// prevFragments holds the [][]byte header returned by the previous
	// splitChunkIntoShards call. It is returned to ecShardsHeaderPool at the
	// start of the next Next() call, once readChunks() has consumed the
	// individual fragment slices. Reset() and Release() also drain it.
	prevFragments [][]byte
}

// createChunkReader create ChunkReader instance
func createChunkReader(fileReader io.Reader, size, chunkSize int64, dataShards, parityShards int, encryptOnUpload bool, uploadMask zboxutil.Uint128, erasureEncoder reedsolomon.Encoder, encscheme encryption.EncryptionScheme, hasher Hasher, chunkNumber int) (ChunkedUploadChunkReader, error) {

	if chunkSize <= 0 {
		return nil, errors.Throw(constants.ErrInvalidParameter, "chunkSize: "+strconv.FormatInt(chunkSize, 10))
	}

	if dataShards <= 0 {
		return nil, errors.Throw(constants.ErrInvalidParameter, "dataShards: "+strconv.Itoa(dataShards))
	}

	if erasureEncoder == nil {
		return nil, errors.Throw(constants.ErrInvalidParameter, "erasureEncoder")
	}

	if hasher == nil {
		return nil, errors.Throw(constants.ErrInvalidParameter, "hasher")
	}

	r := &chunkedUploadChunkReader{
		fileReader:      fileReader,
		size:            size,
		chunkSize:       chunkSize,
		nextChunkIndex:  0,
		dataShards:      dataShards,
		totalShards:     dataShards + parityShards,
		encryptOnUpload: encryptOnUpload,
		uploadMask:      uploadMask,
		erasureEncoder:  erasureEncoder,
		encscheme:       encscheme,
		hasher:          hasher,
		hasherDataChan:  make(chan []byte, 3*chunkNumber),
		hasherWG:        sync.WaitGroup{},
		chunkNumber:     int64(chunkNumber),
	}

	if r.encryptOnUpload {
		//additional 16 bytes to save encrypted data
		r.chunkHeaderSize = EncryptedDataPaddingSize + EncryptionHeaderSize
		r.chunkDataSize = chunkSize - r.chunkHeaderSize
	} else {
		r.chunkDataSize = chunkSize
	}

	r.chunkDataSizePerRead = r.chunkDataSize * int64(dataShards)
	r.totalChunkDataSizePerRead = r.chunkDataSize * int64(dataShards+parityShards)
	if CurrentMode == UploadModeHigh {
		r.hasherWG.Add(1)
		go r.hashData()
	}
	return r, nil
}

// ChunkData data of a chunk
type ChunkData struct {
	// Index current index of chunks
	Index int
	// IsFinal last chunk or not
	IsFinal bool

	// ReadSize total size read from original reader (un-encoded, un-encrypted)
	ReadSize int64
	// FragmentSize fragment size for a blobber (un-encrypted)
	FragmentSize int64
	// Fragments data shared for bloobers
	Fragments [][]byte
}

// func (r *chunkReader) GetChunkDataSize() int64 {
// 	if r == nil {
// 		return 0
// 	}
// 	return r.chunkDataSize
// }

// Next read next chunks for blobbers
func (r *chunkedUploadChunkReader) Next() (*ChunkData, error) {

	if r == nil {
		return nil, errors.Throw(constants.ErrInvalidParameter, "r")
	}

	// Return the previous call's fragment-header slice to the pool now that
	// readChunks() has finished iterating chunk.Fragments. The individual
	// []byte elements inside it are sub-slices of fileShardsDataBuffer and
	// remain valid; only the [][]byte header wrapper is recycled.
	r.putPrevFragments()

	if r.fileShardsDataBuffer == nil {
		totalDataSize := r.totalChunkDataSizePerRead * r.chunkNumber
		readSize := r.chunkDataSizePerRead * r.chunkNumber
		if r.size > 0 && readSize > r.size {
			chunkNum := (r.size + r.chunkDataSizePerRead - 1) / r.chunkDataSizePerRead
			totalDataSize = r.totalChunkDataSizePerRead * chunkNum
		}
		buf := uploadPool.Get()
		if cap(buf.B) < int(totalDataSize) {
			logger.Logger.Debug("creating buffer with size: ", " totalDataSize: ", totalDataSize)
			buf.B = make([]byte, 0, totalDataSize)
		} else {
			logger.Logger.Debug("reusing buffer with size: ", cap(buf.B), " totalDataSize: ", totalDataSize, " len: ", len(buf.B))
		}
		r.fileShardsDataBuffer = buf
	}

	chunk := &ChunkData{
		Index:   r.nextChunkIndex,
		IsFinal: false,

		ReadSize:     0,
		FragmentSize: 0,
	}
	chunkBytes := r.fileShardsDataBuffer.B[r.offset : r.offset+r.chunkDataSizePerRead : r.offset+r.totalChunkDataSizePerRead]
	var (
		readLen int
		err     error
	)
	for readLen < len(chunkBytes) && err == nil {
		var nn int
		nn, err = r.fileReader.Read(chunkBytes[readLen:])
		readLen += nn
	}
	if err != nil {

		if !errors.Is(err, io.EOF) {
			return nil, err
		}

		//all bytes are read
		chunk.IsFinal = true
	}

	if readLen == 0 {
		chunk.IsFinal = true
		return chunk, nil
	}

	chunk.FragmentSize = int64(math.Ceil(float64(readLen)/float64(r.dataShards))) + r.chunkHeaderSize
	if readLen < int(r.chunkDataSizePerRead) {
		chunkBytes = chunkBytes[:readLen]
		chunk.IsFinal = true
	}

	chunk.ReadSize = int64(readLen)
	r.readSize += chunk.ReadSize
	if r.size > 0 {
		if r.readSize >= r.size {
			chunk.IsFinal = true
		}
	}

	if r.hasherError != nil {
		return chunk, r.hasherError
	}

	if CurrentMode == UploadModeHigh {
		r.hasherDataChan <- chunkBytes
	} else {
		_ = r.hasher.WriteToFile(chunkBytes)
	}

	// splitChunkIntoShards uses a pooled [][]byte header (via ecShardsHeaderPool)
	// instead of letting erasureEncoder.Split allocate a fresh one each call.
	// fileShardsDataBuffer.B has capacity totalChunkDataSizePerRead per chunk
	// (= chunkDataSize × totalShards), so chunkBytes' cap always covers parity
	// slots — no AllocAligned needed inside Split.
	fragments := r.splitChunkIntoShards(chunkBytes)

	err = r.erasureEncoder.Encode(fragments)
	if err != nil {
		r.releaseFragments(fragments)
		return nil, err
	}
	var pos uint64
	if r.encryptOnUpload {
		for i := r.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
			pos = uint64(i.TrailingZeros())
			encMsg, err := r.encscheme.Encrypt(fragments[pos])
			if err != nil {
				r.releaseFragments(fragments)
				return nil, err
			}
			fragments[pos] = make([]byte, len(encMsg.EncryptedData)+EncryptionHeaderSize)
			n := copy(fragments[pos], encMsg.MessageChecksum+encMsg.OverallChecksum)
			copy(fragments[pos][n:], encMsg.EncryptedData)
		}
	}

	// Keep this call's fragment header in prevFragments. putPrevFragments()
	// will return it to ecShardsHeaderPool at the start of the NEXT Next()
	// call, once readChunks() has finished iterating chunk.Fragments and all
	// individual []byte elements have been appended into data.fileShards.
	r.prevFragments = fragments
	chunk.Fragments = fragments
	r.nextChunkIndex++
	r.offset += r.totalChunkDataSizePerRead
	return chunk, nil
}

// splitChunkIntoShards builds the shard sub-slice header for chunkBytes using
// a pooled [][]byte from ecShardsHeaderPool, avoiding the allocation that
// erasureEncoder.Split would otherwise make on every chunk. The data bytes
// themselves are sub-slices of fileShardsDataBuffer — no copy is performed.
//
// Precondition: cap(data) >= r.totalShards * perShard (guaranteed by how
// fileShardsDataBuffer is allocated in Next()).
func (r *chunkedUploadChunkReader) splitChunkIntoShards(data []byte) [][]byte {
	n := r.totalShards
	// Get or allocate a header slice of the right length.
	var dst [][]byte
	if v := ecShardsHeaderPool.Get(); v != nil {
		if s, ok := v.([][]byte); ok && len(s) == n {
			dst = s
		}
	}
	if dst == nil {
		dst = make([][]byte, n)
	}

	dataShards := r.dataShards
	// perShard mirrors the reedsolomon.Split calculation.
	perShard := (len(data) + dataShards - 1) / dataShards
	needTotal := n * perShard

	// Extend data into the parity area within the existing capacity and zero it.
	// Precondition guarantees cap >= needTotal; the else branch is a safety net.
	dataLen := len(data)
	if cap(data) >= needTotal {
		data = data[:needTotal]
		for i := dataLen; i < needTotal; i++ {
			data[i] = 0
		}
	} else {
		// Should not happen given fileShardsDataBuffer sizing, but fall back to
		// the standard Split (which handles AllocAligned for the parity area).
		// Return the pooled header unused — it will be re-pooled on the next call.
		shards, _ := r.erasureEncoder.Split(data)
		if len(shards) == n {
			copy(dst, shards)
		} else {
			dst = shards
		}
		return dst
	}
	// Build sub-slice headers (no data copy).
	for i := 0; i < n; i++ {
		dst[i] = data[:perShard:perShard]
		data = data[perShard:]
	}
	return dst
}

// releaseFragments is a helper that zeroes element pointers and returns
// fragments to the pool. Used on error paths inside Next() before returning.
func (r *chunkedUploadChunkReader) releaseFragments(frags [][]byte) {
	for i := range frags {
		frags[i] = nil
	}
	ecShardsHeaderPool.Put(frags)
}

// putPrevFragments returns the previous call's fragment-header slice to the
// pool. Safe to call only when readChunks() has finished iterating
// chunk.Fragments (all elements have been appended into data.fileShards).
// Called at the start of the next Next(), and on Reset/Release.
func (r *chunkedUploadChunkReader) putPrevFragments() {
	if r.prevFragments != nil {
		r.releaseFragments(r.prevFragments)
		r.prevFragments = nil
	}
}

// Read read, encode and encrypt all bytes
func (r *chunkedUploadChunkReader) Read(buf []byte) ([][]byte, error) {

	if buf == nil {
		return nil, nil
	}

	if r == nil {
		return nil, errors.Throw(constants.ErrInvalidParameter, "r")
	}

	fragments, err := r.erasureEncoder.Split(buf)
	if err != nil {
		return nil, err
	}

	err = r.erasureEncoder.Encode(fragments)
	if err != nil {
		return nil, err
	}

	var pos uint64
	if r.encryptOnUpload {
		for i := r.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
			pos = uint64(i.TrailingZeros())
			encMsg, err := r.encscheme.Encrypt(fragments[pos])
			if err != nil {
				return nil, err
			}
			fragments[pos] = make([]byte, len(encMsg.EncryptedData)+EncryptionHeaderSize)
			n := copy(fragments[pos], encMsg.MessageChecksum+encMsg.OverallChecksum)
			copy(fragments[pos][n:], encMsg.EncryptedData)
		}
	}

	return fragments, nil
}

func (r *chunkedUploadChunkReader) Reset() {
	// Return the last call's fragment header to the pool before re-use.
	// readChunks() calls Reset() after finishing its chunk loop, so all
	// chunk.Fragments slices have already been fully iterated.
	r.putPrevFragments()
	r.offset = 0
}

func (r *chunkedUploadChunkReader) Close() {
	r.closeOnce.Do(func() {
		close(r.hasherDataChan)
		r.hasherWG.Wait()
	})

}

func (r *chunkedUploadChunkReader) GetFileHash() (string, error) {
	r.Close()
	if r.hasherError != nil {
		return "", r.hasherError
	}
	return r.hasher.GetFileHash()
}

func (r *chunkedUploadChunkReader) Release() {
	// Drain any pending fragment header before releasing the data buffer.
	r.putPrevFragments()
	if r.fileShardsDataBuffer != nil {
		uploadPool.Put(r.fileShardsDataBuffer)
	}
}

func (r *chunkedUploadChunkReader) hashData() {
	defer r.hasherWG.Done()
	for data := range r.hasherDataChan {
		err := r.hasher.WriteToFile(data)
		if err != nil {
			r.hasherError = err
			return
		}
	}
}
