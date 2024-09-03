package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type fileMetaResponse struct {
	fileref    *fileref.FileRef
	blobberIdx int
	err        error
}

type fileMetaByNameResponse struct {
	filerefs   []*fileref.FileRef
	blobberIdx int
	err        error
}

func (req *ListRequest) getFileMetaInfoFromBlobber(blobber *blockchain.StorageNode, blobberIdx int, rspCh chan<- *fileMetaResponse) {
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)

	var fileRef *fileref.FileRef
	var err error
	fileMetaRetFn := func() {
		rspCh <- &fileMetaResponse{fileref: fileRef, blobberIdx: blobberIdx, err: err}
	}
	defer fileMetaRetFn()
	if len(req.remotefilepath) > 0 {
		req.remotefilepathhash = fileref.GetReferenceLookup(req.allocationID, req.remotefilepath)
	}
	if singleClientMode {
		fileMetaHash := fileref.GetCacheKey(req.remotefilepathhash, blobber.ID)
		cachedRef, ok := fileref.GetFileRef(fileMetaHash)
		if ok {
			fileRef = &cachedRef
			return
		}
		defer func() {
			if fileRef != nil && err == nil {
				fileref.StoreFileRef(fileMetaHash, *fileRef)
			}
		}()
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
	httpreq, err := zboxutil.NewFileMetaRequest(blobber.Baseurl, req.allocationID, req.allocationTx, req.sig, body)
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
		resp_body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "Error: Resp")
		}
		// l.Logger.Info("File Meta result:", string(resp_body))
		l.Logger.Debug("File meta response status: ", resp.Status)
		if resp.StatusCode == http.StatusOK {
			err = json.Unmarshal(resp_body, &fileRef)
			if err != nil {
				return errors.Wrap(err, "file meta data response parse error")
			}
			return nil
		} else if resp.StatusCode == http.StatusBadRequest {
			return constants.ErrNotFound
		}
		return fmt.Errorf("unexpected response. status code: %d, response: %s",
			resp.StatusCode, string(resp_body))
	})
}

func (req *ListRequest) getFileMetaByNameInfoFromBlobber(blobber *blockchain.StorageNode, blobberIdx int, rspCh chan<- *fileMetaByNameResponse) {
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)

	var fileRef []*fileref.FileRef
	var err error
	fileMetaRetFn := func() {
		rspCh <- &fileMetaByNameResponse{filerefs: fileRef, blobberIdx: blobberIdx, err: err}
	}
	defer fileMetaRetFn()

	if req.filename != "" {
		err = formWriter.WriteField("name", req.filename)
		if err != nil {
			l.Logger.Error("File meta info request error: ", err.Error())
			return
		}
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
	httpreq, err := zboxutil.NewFileMetaRequest(blobber.Baseurl, req.allocationID, req.allocationTx, req.sig, body)
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
		resp_body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "Error: Resp")
		}
		// l.Logger.Info("File Meta result:", string(resp_body))
		l.Logger.Debug("File meta response status: ", resp.Status)
		if resp.StatusCode == http.StatusOK {
			err = json.Unmarshal(resp_body, &fileRef)
			if err != nil {
				return errors.Wrap(err, "file meta data response parse error")
			}
			return nil
		} else if resp.StatusCode == http.StatusBadRequest {
			return constants.ErrNotFound
		}
		return fmt.Errorf("unexpected response. status code: %d, response: %s",
			resp.StatusCode, string(resp_body))
	})
}

func (req *ListRequest) getFileMetaFromBlobbers() []*fileMetaResponse {
	numList := len(req.blobbers)
	rspCh := make(chan *fileMetaResponse, numList)
	for i := 0; i < numList; i++ {
		go req.getFileMetaInfoFromBlobber(req.blobbers[i], i, rspCh)
	}
	fileInfos := make([]*fileMetaResponse, len(req.blobbers))
	for i := 0; i < numList; i++ {
		ch := <-rspCh
		fileInfos[ch.blobberIdx] = ch
	}
	return fileInfos
}

func (req *ListRequest) getFileMetaByNameFromBlobbers() []*fileMetaByNameResponse {
	numList := len(req.blobbers)
	rspCh := make(chan *fileMetaByNameResponse, numList)
	for i := 0; i < numList; i++ {
		go req.getFileMetaByNameInfoFromBlobber(req.blobbers[i], i, rspCh)
	}
	fileInfos := make([]*fileMetaByNameResponse, len(req.blobbers))
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

func (req *ListRequest) getMultipleFileConsensusFromBlobbers() (zboxutil.Uint128, zboxutil.Uint128, []*fileref.FileRef, []*fileMetaByNameResponse) {
	lR := req.getFileMetaByNameFromBlobbers()
	var filerRefs []*fileref.FileRef
	uniquePathHashes := map[string]bool{}
	for i := 0; i < len(lR); i++ {
		ti := lR[i]
		if ti.err != nil || len(ti.filerefs) == 0 {
			continue
		}
		for _, fileRef := range ti.filerefs {
			uniquePathHashes[fileRef.PathHash] = true
		}
	}
	// take the pathhash as unique and for each path hash append the fileref which have consensus.

	for pathHash := range uniquePathHashes {
		req.consensus = 0
		retMap := make(map[string]int)
	outerLoop:
		for i := 0; i < len(lR); i++ {
			ti := lR[i]
			if ti.err != nil || len(ti.filerefs) == 0 {
				continue
			}
			for _, fRef := range ti.filerefs {
				if fRef == nil {
					continue
				}
				if pathHash == fRef.PathHash {
					fileMetaHash := fRef.FileMetaHash
					retMap[fileMetaHash]++
					if retMap[fileMetaHash] > req.consensus {
						req.consensus = retMap[fileMetaHash]
					}
					if req.isConsensusOk() {
						filerRefs = append(filerRefs, fRef)
						break outerLoop
					}
					break
				}
			}
		}
	}
	return zboxutil.NewUint128(0), zboxutil.NewUint128(0), filerRefs, lR
}

// return upload mask and delete mask
