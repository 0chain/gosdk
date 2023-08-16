package sdk

import (
	"context"
	"io"

	"github.com/0chain/errors"
)

// Example implementation of a reader
type DataChan struct {
	data []byte
	err  error
}

func NewStreamReader(dataChan chan *DataChan) *StreamReader {
	return &StreamReader{dataChan}
}

type StreamReader struct {
	dataChan chan *DataChan
}

// Client should always send bytes equal to less than chunkDataSizePerRead
func (r *StreamReader) Read(p []byte) (int, error) {
	recieveData, ok := <-r.dataChan
	if !ok && recieveData == nil {
		return 0, io.EOF
	}
	if recieveData.err != nil {
		return 0, recieveData.err
	}
	if len(recieveData.data) > len(p) {
		return 0, io.ErrShortBuffer
	}
	copy(p, recieveData.data)
	return len(recieveData.data), nil
}

func StartWriteWorker(ctx context.Context, source io.Reader, dataChan chan *DataChan, chunkDataSizePerRead int64) {
	defer close(dataChan)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		data := make([]byte, chunkDataSizePerRead)
		dataToSend := &DataChan{}
		n, err := source.Read(data)
		if n < int(chunkDataSizePerRead) {
			data = data[:n]
		}
		if err != nil && !errors.Is(err, io.EOF) {
			dataToSend.err = err
			dataChan <- dataToSend
			return
		}
		dataToSend.data = data
		dataChan <- dataToSend
		if n < int(chunkDataSizePerRead) {
			return
		}
	}

}
