// Stream Encoder for erasure coding.
package encoder

import (
	"bufio"
	"bytes"

	"errors"

	l "github.com/0chain/gosdk/zboxcore/logger"

	"github.com/klauspost/reedsolomon"
)

type StreamEncoder struct {
	iDataShards   int
	iParityShards int
	erasureCode   reedsolomon.Encoder
	data          [][]byte
}

// Creates New encoder instance and return index for further access
func NewEncoder(iDataShards, iParityShards int) (*StreamEncoder, error) {
	e := &StreamEncoder{}
	var err error

	e.erasureCode, err = reedsolomon.New(iDataShards, iParityShards, reedsolomon.WithAutoGoroutines(64*1024))
	if err != nil {
		return nil, err
	}
	e.iDataShards = iDataShards
	e.iParityShards = iParityShards
	return e, nil
}

// Encodes and returns the shards on success and error on fails
func (e *StreamEncoder) Encode(in []byte) ([][]byte, error) {
	var err error
	e.data, err = e.erasureCode.Split(in)
	if err != nil {
		l.Logger.Error("Split failed", err.Error())
		return [][]byte{}, err
	}

	err = e.erasureCode.Encode(e.data)
	if err != nil {
		l.Logger.Error("Encode failed", err.Error())
		return [][]byte{}, err
	}
	return e.data, nil
}

func (e *StreamEncoder) Decode(in [][]byte, shardSize int) ([]byte, error) {
	// Verify the input
	if (len(in) < e.iDataShards+e.iParityShards) || (shardSize <= 0) {
		return []byte{}, errors.New("Invalid input length")
	}

	err := e.erasureCode.Reconstruct(in)
	if err != nil {
		l.Logger.Error("Reconstruct failed -", err)
		return []byte{}, err
	}
	_, err = e.erasureCode.Verify(in)
	if err != nil {
		l.Logger.Error("Verification failed after reconstruction, data likely corrupted.", err.Error())
		return []byte{}, err
	}

	var bytesBuf bytes.Buffer
	bufWriter := bufio.NewWriter(&bytesBuf)
	bufWriter = bufio.NewWriterSize(bufWriter, (shardSize * e.iDataShards))
	err = e.erasureCode.Join(bufWriter, in, (shardSize * e.iDataShards))
	if err != nil {
		l.Logger.Error("join failed", err.Error())
		return []byte{}, err
	}
	bufWriter.Flush()
	outBuf := bytesBuf.Bytes()
	return outBuf, nil
}
