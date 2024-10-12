package sdk

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sync"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

const (
	// EncryptionOverHead File size increases by 16 bytes after encryption. Two checksums i.e. MessageChecksum and OverallChecksum has
	// 128 bytes size each.
	// So total overhead for each encrypted data is 16 + 128*2 = 272
	EncryptionOverHead = 272
	ChecksumSize       = 256
	HeaderSize         = 128
	BlockSize          = 64 * KB
)

// error codes
const (
	NotEnoughTokens              = "not_enough_tokens"
	InvalidAuthTicket            = "invalid_authticket"
	InvalidShare                 = "invalid_share"
	InvalidRead                  = "invalid_read"
	ExceededMaxOffsetValue       = "exceeded_max_offset_value"
	NegativeOffsetResultantValue = "negative_offset_resultant_value"
	InvalidWhenceValue           = "invalid_whence_value"
)

// errors
var ErrInvalidRead = errors.New(InvalidRead, "want_size is <= 0")

const (
	// BlocksFor10MB is number of blocks required for to make 10MB data.
	// It is simply calculated as 10MB / 64KB = 160
	// If blobber cannot respond with 10MB data then client can use numBlocks field
	// in StreamDownload struct
	BlocksFor10MB = 160
)

type StreamDownloadOption struct {
	ContentMode     string
	AuthTicket      string
	BlocksPerMarker uint // Number of blocks to download per request
	VerifyDownload  bool // Verify downloaded data against ValidaitonRoot.
}

type StreamDownload struct {
	*DownloadRequest
	offset   int64
	open     bool
	fileSize int64
}

func (sd *StreamDownload) Close() error {
	sd.open = false
	return nil
}

func (sd *StreamDownload) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		if offset > sd.fileSize {
			return 0, errors.New(ExceededMaxOffsetValue, "")
		}
		sd.offset = offset
	case io.SeekCurrent:
		if sd.offset+offset >= sd.fileSize {
			return 0, errors.New(ExceededMaxOffsetValue, "")
		}
		sd.offset += offset
	case io.SeekEnd:
		newOffset := sd.fileSize - offset
		if newOffset < 0 {
			return 0, errors.New(NegativeOffsetResultantValue, "")
		}
		sd.offset = offset
	default:
		return 0, errors.New(InvalidWhenceValue,
			fmt.Sprintf("expected 0, 1 or 2, provided %d", whence))
	}
	return sd.offset, nil
}

// getStartAndEndIndex will return start and end index based on fileSize, offset and wantSize value
func (sd *StreamDownload) getStartAndEndIndex(wantsize int64) (int64, int64) {
	sizePerBlobber := (sd.fileSize +
		int64(sd.datashards) - 1) / int64(sd.datashards) // equivalent to ceil(filesize/datashards)

	totalBlocksPerBlobber := (sizePerBlobber +
		int64(sd.effectiveBlockSize) - 1) / int64(sd.effectiveBlockSize)

	effectiveChunkSize := sd.effectiveBlockSize * sd.datashards
	startInd := sd.offset / int64(effectiveChunkSize)
	endInd := (sd.offset + wantsize + int64(effectiveChunkSize) - 1) / int64(effectiveChunkSize)
	if endInd > totalBlocksPerBlobber {
		endInd = totalBlocksPerBlobber
	}
	return startInd, endInd
}

func (sd *StreamDownload) Read(b []byte) (int, error) {
	if !sd.open {
		return 0, errors.New("file_closed", "")
	}

	if sd.offset >= sd.fileSize {
		return 0, io.EOF
	}

	wantSize := int64(math.Min(float64(len(b)), float64(sd.fileSize-sd.offset)))
	if wantSize <= 0 {
		return 0, ErrInvalidRead
	}

	startInd, endInd := sd.getStartAndEndIndex(wantSize)
	var numBlocks int64
	if sd.numBlocks > 0 {
		numBlocks = sd.numBlocks
	} else {
		numBlocks = endInd - startInd
		if numBlocks > BlocksFor10MB {
			numBlocks = BlocksFor10MB
		}
	}

	wantBlocksPerShard := (wantSize + int64(sd.effectiveBlockSize) - 1) / int64(sd.effectiveBlockSize)
	sd.blocksPerShard = wantBlocksPerShard

	// effectiveChunkSize := sd.effectiveBlockSize * sd.datashards
	n := 0
	for startInd < endInd {
		if startInd+numBlocks > endInd {
			// this numBlocks should not exceed number greater than required data
			// otherwise `no shard data` error will occur in erasure reconstruction.
			numBlocks = endInd - startInd
		}

		data, err := sd.getBlocksData(startInd, numBlocks, true)
		if err != nil {
			return 0, err
		}

		// offset := sd.offset % int64(effectiveChunkSize)
		// size of buffer `b` can be any number but we don't want to copy more than want size
		// offset is important parameter because without it data will be corrupted.
		// If previously set offset was 65536 + 1(block number 0) and we get data block with block number 1
		// then we should not copy whole data to the buffer rather after offset.
		n += copy(b[n:wantSize], data[0][0])

		startInd += numBlocks
	}
	sd.offset += int64(n)
	return n, nil
}

// GetDStorageFileReader will initialize erasure decoder, decrypter if file is encrypted and other
// necessary fields and returns a reader that comply with io.ReadSeekCloser interface.
func GetDStorageFileReader(alloc *Allocation, ref *ORef, sdo *StreamDownloadOption) (io.ReadSeekCloser, error) {

	sd := &StreamDownload{
		DownloadRequest: &DownloadRequest{
			allocationID:     alloc.ID,
			allocationTx:     alloc.Tx,
			allocOwnerID:     alloc.Owner,
			allocOwnerPubKey: alloc.OwnerPublicKey,
			datashards:       alloc.DataShards,
			parityshards:     alloc.ParityShards,
			remotefilepath:   ref.Path,
			numBlocks:        int64(sdo.BlocksPerMarker),
			shouldVerify:     sdo.VerifyDownload,
			Consensus: Consensus{
				RWMutex:         &sync.RWMutex{},
				fullconsensus:   alloc.fullconsensus,
				consensusThresh: alloc.consensusThreshold,
			},
			blobbers:           alloc.Blobbers,
			downloadMask:       zboxutil.NewUint128(1).Lsh(uint64(len(alloc.Blobbers))).Sub64(1),
			effectiveBlockSize: BlockSize,
			chunkSize:          BlockSize,
			maskMu:             &sync.Mutex{},
			connectionID:       zboxutil.NewConnectionId(),
		},
		open: true,
	}

	if sdo.ContentMode == DOWNLOAD_CONTENT_THUMB {
		sd.fileSize = ref.ActualThumbnailSize
	} else {
		sd.fileSize = ref.ActualFileSize
	}

	if sdo.AuthTicket != "" {
		sEnc, err := base64.StdEncoding.DecodeString(sdo.AuthTicket)
		if err != nil {
			return nil, errors.New("auth_ticket_decode_error", "Error decoding the auth ticket."+err.Error())
		}
		at := &marker.AuthTicket{}
		err = json.Unmarshal(sEnc, at)
		if err != nil {
			return nil, errors.New("auth_ticket_decode_error", "Error unmarshaling the auth ticket."+err.Error())
		}

		sd.authTicket = at
	}

	sd.ctx, sd.ctxCncl = context.WithCancel(alloc.ctx)

	err := sd.initEC()
	if err != nil {
		return nil, err
	}

	if ref.EncryptedKey != "" {
		sd.effectiveBlockSize = BlockSize - EncryptionOverHead
		sd.encryptedKey = ref.EncryptedKey
		err = sd.initEncryption()
		if err != nil {
			return nil, err
		}
	}

	return sd, err
}
