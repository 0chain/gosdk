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

var uploadPool bytebufferpool.Pool

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
	offset int64

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
		encryptOnUpload: encryptOnUpload,
		uploadMask:      uploadMask,
		erasureEncoder:  erasureEncoder,
		encscheme:       encscheme,
		hasher:          hasher,
		hasherDataChan:  make(chan []byte, 3*chunkNumber),
		hasherWG:        sync.WaitGroup{},
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
	totalDataSize := r.totalChunkDataSizePerRead * int64(chunkNumber)
	readSize := r.chunkDataSizePerRead * int64(chunkNumber)
	if size > 0 && readSize > size {
		chunkNum := (size + r.chunkDataSizePerRead - 1) / r.chunkDataSizePerRead
		totalDataSize = r.totalChunkDataSizePerRead * chunkNum
	}
	buf := uploadPool.Get()
	if cap(buf.B) < int(totalDataSize) {
		buf.B = make([]byte, 0, totalDataSize)
		logger.Logger.Debug("creating buffer with size: ", " totalDataSize: ", totalDataSize)
	} else {
		logger.Logger.Debug("reusing buffer with size: ", cap(buf.B), " totalDataSize: ", totalDataSize, " len: ", len(buf.B))
	}
	r.fileShardsDataBuffer = buf
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

	fragments, err := r.erasureEncoder.Split(chunkBytes)
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

	chunk.Fragments = fragments
	r.nextChunkIndex++
	r.offset += r.totalChunkDataSizePerRead
	return chunk, nil
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
	r.offset = 0
}

func (r *chunkedUploadChunkReader) Close() {
	r.closeOnce.Do(func() {
		close(r.hasherDataChan)
		r.hasherWG.Wait()
		uploadPool.Put(r.fileShardsDataBuffer)
		r.fileShardsDataBuffer = nil
	})

}

func (r *chunkedUploadChunkReader) GetFileHash() (string, error) {
	r.Close()
	if r.hasherError != nil {
		return "", r.hasherError
	}
	return r.hasher.GetFileHash()
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
