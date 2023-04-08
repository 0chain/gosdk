package sdk

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sync"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type FileStatus int

const (
	Closed = iota
	Open
)

const Retry = 3

const (
	// EncryptionOverHead File size increases by 16 bytes after encryption. Two checksums i.e. MessageChecksum and OverallChecksum has
	// 128 bytes size each.
	// So total overhead for each encrypted data is 16 + 128*2 = 272
	EncryptionOverHead = 272
	ChecksumSize       = 256
	HeaderSize         = 128
	BlockSize          = 64 * KB

	// TooManyRequestWaitTime wait for this time to re-request when too_many_requests errors ocucurs
	TooManyRequestWaitTime = time.Millisecond * 100
)

// error codes
const (
	LessThan67Percent            = "less_than_67_percent"
	ExceedingFailedBlobber       = "exceeding_failed_blobber"
	ReadCounterUpdate            = "rc_update"
	TooManyRequests              = "too_many_requests"
	ContextCancelled             = "context_cancelled"
	MarshallError                = "marshall_error"
	SigningError                 = "error_while_signing"
	ReedSolomonEndocerError      = "reedsolomon_endocer_error"
	ErasureReconstructError      = "erasure_reconstruct_error"
	ResponseError                = "response_error"
	NoRequiredShards             = "no_required_shards"
	NotEnoughTokens              = "not_enough_tokens"
	Panic                        = "code_panicked"
	InvalidHeader                = "invalid_header"
	DecryptionError              = "decryption_error"
	UnknownDownloadType          = "unknown_download_type"
	InvalidBlocksPerMarker       = "invalid_blocks_per_marker"
	ReDecryptUnmarshallFail      = "redecrypt_unmarshall_fail"
	ReDecryptionFail             = "redecryption_fail"
	InvalidRead                  = "invalid_read"
	InvalidDownloadType          = "invalid_download_type"
	StaleReadMarker              = "stale_read_marker"
	InvalidReadMarker            = "invalid_read_marker"
	ExceededMaxOffsetValue       = "exceeded_max_offset_value"
	NegativeOffsetResultantValue = "negative_offset_resultant_value"
	InvalidWhenceValue           = "invalid_whence_value"
)

//errors
var (
	ErrLessThan67PercentBlobber = errors.New(LessThan67Percent, "less than 67% blobbers able to respond")
	ErrReadCounterUpdate        = errors.New(ReadCounterUpdate, "")
	ErrTooManyRequests          = errors.New(TooManyRequests, "")
	ErrContextCancelled         = errors.New(ContextCancelled, "")
	ErrMarshallError            = errors.New(MarshallError, "")
	ErrSigningError             = errors.New(SigningError, "")
	ErrReedSolomonEncoderError  = errors.New(ReedSolomonEndocerError, "")
	ErrErasureReconstructError  = errors.New(ErasureReconstructError, "")
	ErrFromResponse             = errors.New(ResponseError, "")
	ErrNoRequiredShards         = errors.New(NoRequiredShards, "")
	ErrNotEnoughTokens          = errors.New(NotEnoughTokens, "")
	ErrPanic                    = errors.New(Panic, "")
	ErrInvalidHeader            = errors.New(InvalidHeader, "")
	ErrDecryption               = errors.New(DecryptionError, "")
	ErrUnknownDownloadType      = errors.New(UnknownDownloadType, "")
	ErrInvalidBlocksPerMarker   = errors.New(InvalidBlocksPerMarker, "")
	ErrReDecryptUnmarshallFail  = errors.New(ReDecryptUnmarshallFail, "")
	ErrReDecryptionFail         = errors.New(ReDecryptionFail, "")
	ErrInvalidRead              = errors.New(InvalidRead, "want_size is <= 0")
	ErrStaleReadMarker          = errors.New(StaleReadMarker, "")
	ErrInvalidReadMarker        = errors.New(InvalidReadMarker, "")
)

// errors func
var (
	ErrExceedingFailedBlobber = func(failed, parity int) error {
		msg := fmt.Sprintf("number of failed %v blobbers exceeds %v parity shards", failed, parity)
		return errors.New(ExceedingFailedBlobber, msg)
	}

	ErrInvalidDownloadType = func(downloadType string) error {
		msg := fmt.Sprintf("%v download type is not supported", downloadType)
		return errors.New(InvalidDownloadType, msg)
	}
)

const (
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

func (sd *StreamDownload) getStartAndEndIndex(wantsize int64) (int64, int64) {
	sizePerBlobber := (sd.fileSize + int64(sd.datashards) - 1) / int64(sd.datashards)
	totalBlocksPerBlobber := (sizePerBlobber + int64(sd.effectiveBlockSize) - 1) / int64(sd.effectiveBlockSize)

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

	effectiveChunkSize := sd.effectiveBlockSize * sd.datashards
	n := 0
	for startInd < endInd {
		if startInd+numBlocks > endInd {
			numBlocks = endInd - startInd
		}

		data, err := sd.getBlocksData(startInd, numBlocks)
		if err != nil {
			return 0, err
		}

		offset := sd.offset % int64(effectiveChunkSize)
		n += copy(b[n:wantSize], data[offset:])

		startInd += numBlocks
	}

	sd.offset += int64(n)

	return n, nil
}

// GetDStorageFileReader Get a reader that provides io.Reader interface
func GetDStorageFileReader(alloc *Allocation, ref *ORef, sdo *StreamDownloadOption) (io.ReadSeekCloser, error) {

	sd := &StreamDownload{
		DownloadRequest: &DownloadRequest{
			allocationID:      alloc.ID,
			allocationTx:      alloc.Tx,
			allocOwnerID:      alloc.Owner,
			allocOwnerPubKey:  alloc.OwnerPublicKey,
			datashards:        alloc.DataShards,
			parityshards:      alloc.ParityShards,
			remotefilepath:    ref.Path,
			numBlocks:         int64(sdo.BlocksPerMarker),
			validationRootMap: make(map[string]*blobberFile),
			shouldVerify:      sdo.VerifyDownload,
			Consensus: Consensus{
				fullconsensus:   alloc.fullconsensus,
				consensusThresh: alloc.consensusThreshold,
			},
			blobbers:           alloc.Blobbers,
			downloadMask:       zboxutil.NewUint128(1).Lsh(uint64(len(alloc.Blobbers))).Sub64(1),
			effectiveBlockSize: BlockSize,
			chunkSize:          BlockSize,
			maskMu:             &sync.Mutex{},
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
