package sdk

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	"errors"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"

	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type CopyRequest struct {
	allocationObj  *Allocation
	allocationID   string
	allocationTx   string
	blobbers       []*blockchain.StorageNode
	remotefilepath string
	destPath       string
	ctx            context.Context

	connectionID string
	Consensus
}

type CopyResult struct {
	BlobberIndex int
	FileRef      fileref.RefEntity
	Copied       bool
}

func (req *CopyRequest) getObjectTreeFromBlobber(blobber *blockchain.StorageNode) (fileref.RefEntity, error) {
	return getObjectTreeFromBlobber(req.ctx, req.allocationID, req.allocationTx, req.remotefilepath, blobber)
}

func (req *CopyRequest) copyBlobberObject(blobber *blockchain.StorageNode) (fileref.RefEntity, error) {
	refEntity, err := req.getObjectTreeFromBlobber(blobber)
	if err != nil {
		return nil, err
	}

	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)

	_ = formWriter.WriteField("connection_id", req.connectionID)
	formWriter.WriteField("path", req.remotefilepath)
	formWriter.WriteField("dest", req.destPath)

	formWriter.Close()
	httpreq, err := zboxutil.NewCopyRequest(blobber.Baseurl, req.allocationTx, body)
	if err != nil {
		l.Logger.Error(blobber.Baseurl, "Error creating rename request", err)
		return nil, err
	}
	httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
	l.Logger.Info(httpreq.URL.Path)
	ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 30))
	err = zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			l.Logger.Error("Copy : ", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			resp_body, _ := ioutil.ReadAll(resp.Body)
			l.Logger.Info("copy resp:", string(resp_body))
			req.Done()

			l.Logger.Info(blobber.Baseurl, " "+req.remotefilepath, " copied.")
			return nil
		}

		resp_body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			msg := string(resp_body)

			if len(msg) > 0 {
				l.Logger.Error(blobber.Baseurl, "Response: ", msg)
				return fmt.Errorf("Copy: %v %s", resp.StatusCode, msg)
			}
		}

		return fmt.Errorf("Copy: %v", resp.StatusCode)
	})

	if err != nil {
		return nil, err
	}
	return refEntity, nil
}

func (req *CopyRequest) ProcessCopy() error {
	num := len(req.blobbers)
	objectTreeRefs := make([]fileref.RefEntity, num)

	wait := make(chan CopyResult, num)

	wg := &sync.WaitGroup{}
	wg.Add(num)

	for i := 0; i < num; i++ {
		go func(blobberIdx int) {
			defer wg.Done()
			refEntity, err := req.copyBlobberObject(req.blobbers[blobberIdx])
			if err != nil {
				l.Logger.Error(err.Error())
			}

			wait <- CopyResult{
				BlobberIndex: blobberIdx,
				FileRef:      refEntity,
				Copied:       err == nil,
			}

		}(i)
	}
	wg.Wait()

	for i := 0; i < num; i++ {
		r := <-wait

		if !r.Copied {
			continue
		}

		objectTreeRefs[r.BlobberIndex] = r.FileRef
	}

	if !req.isConsensusOk() {
		return errors.New("Copy failed: Copy request failed. Operation failed.")
	}

	writeMarkerMutex, err := CreateWriteMarkerMutex(client.GetClient(), req.allocationObj)
	if err != nil {
		return fmt.Errorf("Copy failed: %s", err.Error())
	}
	err = writeMarkerMutex.Lock(context.TODO(), req.connectionID)
	defer writeMarkerMutex.Unlock(context.TODO(), req.connectionID) //nolint: errcheck
	if err != nil {
		return fmt.Errorf("Copy failed: %s", err.Error())
	}

	req.Reset()
	commitReqs := make([]*CommitRequest, 0, num)

	for pos, ref := range objectTreeRefs {
		if ref == nil {
			continue
		}

		wg.Add(1)
		//go req.prepareUpload(a, a.Blobbers[pos], req.file[c], req.uploadDataCh[c], req.wg)
		commitReq := &CommitRequest{}
		commitReq.allocationID = req.allocationID
		commitReq.allocationTx = req.allocationTx
		commitReq.blobber = req.blobbers[pos]
		newChange := &allocationchange.CopyFileChange{}
		newChange.DestPath = req.destPath
		newChange.ObjectTree = objectTreeRefs[pos]
		newChange.NumBlocks = 0
		newChange.Operation = constants.FileOperationCopy
		newChange.Size = 0
		commitReq.changes = append(commitReq.changes, newChange)
		commitReq.connectionID = req.connectionID
		commitReq.wg = wg
		commitReqs = append(commitReqs, commitReq)
		go AddCommitRequest(commitReq)
	}
	wg.Wait()

	for _, commitReq := range commitReqs {
		if commitReq.result != nil {
			if commitReq.result.Success {
				l.Logger.Info("Commit success", commitReq.blobber.Baseurl)
				req.consensus++
			} else {
				l.Logger.Info("Commit failed", commitReq.blobber.Baseurl, commitReq.result.ErrorMessage)
			}
		} else {
			l.Logger.Info("Commit result not set", commitReq.blobber.Baseurl)
		}
	}

	if !req.isConsensusOk() {
		return errors.New("Copy failed: Commit consensus failed")
	}
	return nil
}
