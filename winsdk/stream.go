package main

/*
#include <stdlib.h>
*/
import (
	"C"
)

import (
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
func StartStreamServer(allocationID *C.char) *C.char {
	allocID := C.GoString(allocationID)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		streamingMedia(w, r)
	})

	if server != nil {
		server.Close()
	}

	server = httptest.NewServer(handler)
	streamAllocationID = allocID
	log.Info("win: ", server.URL)
	return WithJSON(server.URL, nil)
}

func streamingMedia(w http.ResponseWriter, req *http.Request) {

	remotePath := req.URL.Path
	log.Info("win: start streaming media: ", streamAllocationID, remotePath)

	f, err := getFileMeta(streamAllocationID, remotePath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", f.MimeType)

	rangeHeader := req.Header.Get("Range")

	log.Info("win: range: ", rangeHeader, " mimetype: ", f.MimeType, " numBlocks:", f.NumBlocks, " ActualNumBlocks:", f.ActualNumBlocks, " ActualFileSize:", f.ActualFileSize)
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
			buf, err := downloadBlocks(remotePath, f, ra)
			if err != nil {
				http.Error(w, err.Error(), 500)
			}

			written, err := w.Write(buf)

			if err != nil {
				http.Error(w, err.Error(), 500)
			}

			w.Header().Set("Content-Length", strconv.Itoa(written))

			for k, v := range w.Header() {
				log.Info("win: response ", k, " = ", v[0])
			}

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

		buf, err := downloadBlocks(remotePath, f, ra)

		if err != nil {
			http.Error(w, err.Error(), 500)
		}

		written, err := w.Write(buf)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		w.Header().Set("Content-Length", strconv.Itoa(written))

		for k, v := range w.Header() {
			log.Info("win: response ", k, " = ", v[0])
		}
	}
}

func downloadBlocks(remotePath string, f *sdk.ConsolidatedFileMeta, ra httpRange) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: ", r)
		}
	}()
	alloc, err := getAllocation(streamAllocationID)

	if err != nil {
		return nil, err
	}
	var startBlock int64

	if ra.start == 0 {
		startBlock = 1
	} else {
		startBlock = int64(math.Floor(float64(ra.start)/float64(sdk.CHUNK_SIZE)/float64(alloc.DataShards))) + 1
	}

	blocks := int(math.Ceil(float64(ra.length) / float64(sdk.CHUNK_SIZE) / float64(alloc.DataShards)))

	endBlock := startBlock + int64(blocks)

	if endBlock > f.NumBlocks {
		endBlock = f.NumBlocks
	}

	if startBlock == endBlock {
		endBlock = 0
	}

	offset := 0
	blockStart := (startBlock - 1) * sdk.CHUNK_SIZE * int64(alloc.DataShards)
	if ra.start > blockStart {
		offset = int(ra.start) - int(blockStart)
	}

	lookupHash := getLookupHash(streamAllocationID, remotePath)
	key := lookupHash + fmt.Sprintf(":%v-%v", startBlock, endBlock)
	log.Info("win: start download blocks ", startBlock, " - ", endBlock, "/", f.NumBlocks, "(", f.ActualNumBlocks, ") for ", remotePath)
	buf, ok := cachedDownloadedBlocks.Get(key)
	if ok {
		return buf, nil
	}

	statusBar := NewStatusBar(statusDownload, key)

	status := statusBar.getStatus(key)
	status.wg.Add(1)

	//mf := &sys.MemFile{}
	//err = alloc.DownloadByBlocksToFileHandler(mf, remotePath, startBlock, endBlock, numBlocks, true, statusBar, true)
	mf := filepath.Join(os.TempDir(), strings.ReplaceAll(remotePath, "/", "_")+fmt.Sprintf("_%v_%v", startBlock, endBlock))
	defer os.Remove(mf)

	log.Info("win: download blocks to ", mf)
	err = alloc.DownloadFileByBlock(mf, remotePath, startBlock, endBlock, numBlocks, false, statusBar, true)
	//err = alloc.DownloadFile(mf, remotePath, true, statusBar, true)
	if err != nil {
		return nil, err
	}

	log.Info("win: waiting for download to done")
	status.wg.Wait()
	//buf = mf.Buffer.Bytes()

	buf, err = os.ReadFile(mf)
	if err != nil {
		return nil, err
	}

	log.Info("win: downloaded blocks ", len(buf), " start:", ra.start, " blockStart:", blockStart, " offset:", offset, " len:", ra.length)
	if len(buf) > 0 {
		if len(buf) > int(ra.length) {
			b := buf[offset : offset+int(ra.length)]
			cachedDownloadedBlocks.Add(key, b)
			return b, nil
		}
		cachedDownloadedBlocks.Add(key, buf)
		return buf[offset:], nil
	}

	return nil, nil
}
