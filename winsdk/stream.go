package main

import (
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"

	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/sdk"
	lru "github.com/hashicorp/golang-lru/v2"
)

var (
	numBlocks                 int   = 100
	sizePerRequest            int64 = sdk.CHUNK_SIZE * int64(numBlocks) // 6400K 100 blocks
	server                    *httptest.Server
	streamAllocationID        string
	cachedDownloadedBlocks, _ = lru.New[string, []byte](1000)
)

// StartStreamServer - start local media stream server
// ## Inputs
//   - allocationID
//
// ## Outputs
//
//	{
//	"error":"",
//	"result":"http://127.0.0.1:4313/",
//	}
//
//export StartStreamServer
func StartStreamServer(allocationID string) (string, error) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		streamingMedia(w, r)
	})

	server = httptest.NewServer(handler)
	streamAllocationID = allocationID

	return server.URL, nil
}

func streamingMedia(w http.ResponseWriter, req *http.Request) {

	remotePath := req.URL.Path
	f, err := getFileMeta(streamAllocationID, remotePath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", f.MimeType)

	rangeHeader := req.Header.Get("Range")
	// we can simply hint Chrome to send serial range requests for media file by
	//
	// if rangeHeader == "" {
	// 	w.Header().Set("Accept-Ranges", "bytes")
	// 	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	// 	w.WriteHeader(200)
	// 	fmt.Printf("hint browser to send range requests, total size: %d\n", size)
	// 	return
	// }
	//
	// but this not worked for Safari and Firefox
	if rangeHeader == "" {
		ra := httpRange{
			start:  0,
			length: sizePerRequest,
			total:  f.ActualFileSize,
		}
		w.Header().Set("Accept-Ranges", "bytes")

		w.Header().Set("Content-Range", ra.Header())

		w.WriteHeader(http.StatusPartialContent)

		if req.Method != "HEAD" {
			buf, err := downloadBlocks(remotePath, ra)
			if err != nil {
				http.Error(w, err.Error(), 500)
			}

			written, err := w.Write(buf)

			if err != nil {
				http.Error(w, err.Error(), 500)
			}

			w.Header().Set("Content-Length", strconv.Itoa(written))

		}
		return
	}

	ranges, err := parseRange(rangeHeader, f.ActualFileSize)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// multi-part requests are not supported
	if len(ranges) > 1 {
		http.Error(w, "unsupported multi-part", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	ra := ranges[0]

	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Range", ra.Header())

	w.WriteHeader(http.StatusPartialContent)

	if req.Method != "HEAD" {

		buf, err := downloadBlocks(remotePath, ra)

		if err != nil {
			http.Error(w, err.Error(), 500)
		}

		written, err := w.Write(buf)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		w.Header().Set("Content-Length", strconv.Itoa(written))
	}
}

func downloadBlocks(remotePath string, ra httpRange) ([]byte, error) {

	alloc, err := getAllocation(streamAllocationID)

	if err != nil {
		return nil, err
	}

	startBlock := int64(math.Ceil(float64(ra.start)/float64(sdk.CHUNK_SIZE)/float64(alloc.DataShards))) + 1

	endBlock := int64(math.Ceil(float64(ra.start+ra.length)/float64(sdk.CHUNK_SIZE)/float64(alloc.DataShards))) + 1

	lookupHash := getLookupHash(streamAllocationID, remotePath)
	key := lookupHash + fmt.Sprintf(":%v-%v-%v", startBlock, endBlock, numBlocks)

	buf, ok := cachedDownloadedBlocks.Get(key)
	if ok {
		return buf, nil
	}

	statusBar := NewStatusBar(statusDownload, lookupHash+fmt.Sprintf(":%v-%v-%v", startBlock, endBlock, numBlocks))

	f := &sys.MemFile{}
	err = alloc.DownloadByBlocksToFileHandler(f, remotePath, startBlock, endBlock, numBlocks, true, statusBar, true)

	if err != nil {
		return nil, err
	}

	buf = f.Buffer.Bytes()

	cachedDownloadedBlocks.Add(key, buf)

	return buf, nil
}
