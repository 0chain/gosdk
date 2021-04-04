package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type fileMetaResponse struct {
	fileref     *fileref.FileRef
	responseStr string
	blobberIdx  int
	err         error
}

func (req *ListRequest) getFileMetaInfoFromBlobber(blobber *blockchain.StorageNode, blobberIdx int, rspCh chan<- *fileMetaResponse) {
	defer req.wg.Done()
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)

	var fileRef *fileref.FileRef
	var s strings.Builder
	var err error
	fileMetaRetFn := func() {
		rspCh <- &fileMetaResponse{fileref: fileRef, responseStr: s.String(), blobberIdx: blobberIdx, err: err}
	}
	defer fileMetaRetFn()
	if len(req.remotefilepath) > 0 {
		req.remotefilepathhash = fileref.GetReferenceLookup(req.allocationID, req.remotefilepath)
	}
	formWriter.WriteField("path_hash", req.remotefilepathhash)

	if req.authToken != nil {
		authTokenBytes, err := json.Marshal(req.authToken)
		if err != nil {
			Logger.Error(blobber.Baseurl, " creating auth token bytes", err)
			return
		}
		formWriter.WriteField("auth_token", string(authTokenBytes))
	}

	formWriter.Close()
	httpreq, err := zboxutil.NewFileMetaRequest(blobber.Baseurl, req.allocationTx, body)
	if err != nil {
		Logger.Error("File meta info request error: ", err.Error())
		return
	}

	httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
	ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 30))
	err = zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("GetFileMeta : ", err)
			return err
		}
		defer resp.Body.Close()
		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Error: Resp : %s", err.Error())
		}
		s.WriteString(string(resp_body))
		if resp.StatusCode == http.StatusOK {
			err = json.Unmarshal(resp_body, &fileRef)
			if err != nil {
				return fmt.Errorf("file meta data response parse error: %s", err.Error())
			}
			return nil
		}
		return err
	})
}

func (req *ListRequest) getFileMetaFromBlobbers() []*fileMetaResponse {
	numList := len(req.blobbers)
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numList)
	rspCh := make(chan *fileMetaResponse, numList)
	for i := 0; i < numList; i++ {
		go req.getFileMetaInfoFromBlobber(req.blobbers[i], i, rspCh)
	}
	req.wg.Wait()
	fileInfos := make([]*fileMetaResponse, len(req.blobbers))
	for i := 0; i < numList; i++ {
		ch := <-rspCh
		fileInfos[ch.blobberIdx] = ch
	}
	return fileInfos
}

func (req *ListRequest) getFileConsensusFromBlobbers() (uint32, *fileref.FileRef, []*fileMetaResponse) {
	lR := req.getFileMetaFromBlobbers()
	var selected *fileMetaResponse
	foundMask := uint32(0)
	req.consensus = 0
	retMap := make(map[string]float32)
	for i := 0; i < len(lR); i++ {
		ti := lR[i]
		if ti.err != nil || ti.fileref == nil {
			continue
		}
		actualHash := ti.fileref.ActualFileHash
		retMap[actualHash]++
		if retMap[actualHash] > req.consensus {
			req.consensus = retMap[actualHash]
			selected = ti
		}
		if req.isConsensusOk() {
			selected = ti
			break
		} else {
			selected = nil
		}
	}
	if selected == nil {
		Logger.Error("File consensus not found for ", req.remotefilepath)
		return foundMask, nil, nil
	}

	for i := 0; i < len(lR); i++ {
		if lR[i].fileref != nil && selected.fileref.ActualFileHash == lR[i].fileref.ActualFileHash {
			foundMask |= (1 << uint32(lR[i].blobberIdx))
		}
	}
	return foundMask, selected.fileref, lR
}
