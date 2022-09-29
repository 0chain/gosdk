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

type RenameRequest struct {
	allocationObj  *Allocation
	allocationID   string
	allocationTx   string
	blobbers       []*blockchain.StorageNode
	remotefilepath string
	newName        string
	ctx            context.Context

	connectionID string
	consensus    Consensus
}

type RenameResult struct {
	BlobberIndex int
	FileRef      fileref.RefEntity
	Renamed      bool
}

func (req *RenameRequest) getObjectTreeFromBlobber(blobber *blockchain.StorageNode) (fileref.RefEntity, error) {
	return getObjectTreeFromBlobber(req.ctx, req.allocationID, req.allocationTx, req.remotefilepath, blobber)
}

func (req *RenameRequest) renameBlobberObject(blobber *blockchain.StorageNode) (fileref.RefEntity, error) {
	refEntity, err := req.getObjectTreeFromBlobber(blobber)
	if err != nil {
		return nil, err
	}

	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)

	_ = formWriter.WriteField("connection_id", req.connectionID)
	formWriter.WriteField("path", req.remotefilepath)
	formWriter.WriteField("new_name", req.newName)

	formWriter.Close()
	httpreq, err := zboxutil.NewRenameRequest(blobber.Baseurl, req.allocationTx, body)
	if err != nil {
		l.Logger.Error(blobber.Baseurl, "Error creating rename request", err)
		return nil, err
	}
	httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
	ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 30))
	err = zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			l.Logger.Error("Rename : ", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			req.consensus.Done()

			l.Logger.Info(blobber.Baseurl, " "+req.remotefilepath, " renamed.")
			return nil
		}
		resp_body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			l.Logger.Error(blobber.Baseurl, "Response: ", string(resp_body))

			return fmt.Errorf("Rename: %v %s", resp.StatusCode, string(resp_body))
		}

		return fmt.Errorf("Rename: %v", resp.StatusCode)
	})
	if err != nil {
		return nil, err
	}
	return refEntity, nil
}

func (req *RenameRequest) ProcessRename() error {
	num := len(req.blobbers)
	objectTreeRefs := make([]fileref.RefEntity, num)

	wait := make(chan RenameResult, num)

	wg := &sync.WaitGroup{}
	wg.Add(num)

	for i := 0; i < num; i++ {
		go func(blobberIdx int) {
			defer wg.Done()
			refEntity, err := req.renameBlobberObject(req.blobbers[blobberIdx])

			if err != nil {
				l.Logger.Error(err.Error())
				return
			}

			wait <- RenameResult{
				BlobberIndex: blobberIdx,
				FileRef:      refEntity,
				Renamed:      err == nil,
			}
		}(i)
	}
	wg.Wait()

	if !req.consensus.isConsensusOk() {
		return errors.New("Rename failed: Rename request failed. Operation failed.")
	}

	writeMarkerMutex, err := CreateWriteMarkerMutex(client.GetClient(), req.allocationObj)
	if err != nil {
		return fmt.Errorf("rename failed: %s", err.Error())
	}
	err = writeMarkerMutex.Lock(context.TODO(), req.connectionID)
	defer writeMarkerMutex.Unlock(context.TODO(), req.connectionID) //nolint: errcheck
	if err != nil {
		return fmt.Errorf("rename failed: %s", err.Error())
	}

	req.consensus.Reset()

	commitReqs := make([]*CommitRequest, 0, num)

	for pos, ref := range objectTreeRefs {
		if ref == nil {
			continue
		}

		wg.Add(1)
		commitReq := &CommitRequest{}
		commitReq.allocationID = req.allocationID
		commitReq.allocationTx = req.allocationTx
		commitReq.blobber = req.blobbers[pos]
		newChange := &allocationchange.RenameFileChange{}
		newChange.NewName = req.newName
		newChange.ObjectTree = objectTreeRefs[pos]
		newChange.NumBlocks = 0
		newChange.Operation = constants.FileOperationRename
		newChange.Size = 0
		commitReq.changes = append(commitReq.changes, newChange)
		commitReq.connectionID = req.connectionID
		commitReq.wg = wg
		commitReqs = append(commitReqs, commitReq)
		go AddCommitRequest(commitReq)
	}
	wg.Wait()

	var errMessages string
	for _, commitReq := range commitReqs {
		if commitReq.result != nil {
			if commitReq.result.Success {
				l.Logger.Info("Commit success", commitReq.blobber.Baseurl)
				req.consensus.Done()
			} else {
				errMessages += commitReq.result.ErrorMessage + "\t"
				l.Logger.Info("Commit failed", commitReq.blobber.Baseurl, commitReq.result.ErrorMessage)
			}
		} else {
			l.Logger.Info("Commit result not set", commitReq.blobber.Baseurl)
		}
	}

	if !req.consensus.isConsensusOk() {
		return errors.New("rename failed: Commit consensus failed. Error: " + errMessages)
	}
	return nil
}
