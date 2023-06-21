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

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	l "github.com/0chain/gosdk/zboxcore/logger"
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
	err = formWriter.WriteField("path_hash", req.remotefilepathhash)
	if err != nil {
		l.Logger.Error("File meta info request error: ", err.Error())
		return
	}

	if req.authToken != nil {
		authTokenBytes, err := json.Marshal(req.authToken)
		if err != nil {
			l.Logger.Error(blobber.Baseurl, " creating auth token bytes", err)
			return
		}
		err = formWriter.WriteField("auth_token", string(authTokenBytes))
		if err != nil {
			l.Logger.Error(blobber.Baseurl, "error writing field", err)
			return
		}
	}

	formWriter.Close()
	httpreq, err := zboxutil.NewFileMetaRequest(blobber.Baseurl, req.allocationID, req.allocationTx, body)
	if err != nil {
		l.Logger.Error("File meta info request error: ", err.Error())
		return
	}

	httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
	ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 30))
	err = zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			l.Logger.Error("GetFileMeta : ", err)
			return err
		}
		defer resp.Body.Close()
		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "Error: Resp")
		}
		l.Logger.Info("File Meta result:", string(resp_body))
		l.Logger.Debug("File meta response status: ", resp.Status)
		s.WriteString(string(resp_body))
		if resp.StatusCode == http.StatusOK {
			err = json.Unmarshal(resp_body, &fileRef)
			if err != nil {
				return errors.Wrap(err, "file meta data response parse error")
			}
			return nil
		}
		return fmt.Errorf("unexpected response. status code: %d, response: %s",
			resp.StatusCode, s.String())
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

func (req *ListRequest) getFileConsensusFromBlobbers() (zboxutil.Uint128, zboxutil.Uint128, *fileref.FileRef, []*fileMetaResponse) {
	lR := req.getFileMetaFromBlobbers()
	var selected *fileMetaResponse
	foundMask := zboxutil.NewUint128(0)
	deleteMask := zboxutil.NewUint128(0)
	req.consensus = 0
	retMap := make(map[string]int)
	for i := 0; i < len(lR); i++ {
		ti := lR[i]
		if ti.err != nil || ti.fileref == nil {
			continue
		}
		fileMetaHash := ti.fileref.FileMetaHash
		retMap[fileMetaHash]++
		if retMap[fileMetaHash] > req.consensus {
			req.consensus = retMap[fileMetaHash]
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
		l.Logger.Error("File consensus not found for ", req.remotefilepath)
		for i := 0; i < len(lR); i++ {
			ti := lR[i]
			if ti.err != nil || ti.fileref == nil {
				continue
			}
			shift := zboxutil.NewUint128(1).Lsh(uint64(ti.blobberIdx))
			deleteMask = deleteMask.Or(shift)
		}
		return foundMask, deleteMask, nil, nil
	}

	for i := 0; i < len(lR); i++ {
		if lR[i].fileref != nil && selected.fileref.FileMetaHash == lR[i].fileref.FileMetaHash {
			shift := zboxutil.NewUint128(1).Lsh(uint64(lR[i].blobberIdx))
			foundMask = foundMask.Or(shift)
		} else if lR[i].fileref != nil {
			shift := zboxutil.NewUint128(1).Lsh(uint64(lR[i].blobberIdx))
			deleteMask = deleteMask.Or(shift)
		}
	}
	return foundMask, deleteMask, selected.fileref, lR
}

// return upload mask and delete mask
