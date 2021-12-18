package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type AllocRenameRequest struct {
	allocationID    string
	allocationTx    string
	blobbers        []*blockchain.StorageNode
	name            string
	authToken       *marker.AuthTicket
	ctx             context.Context
	wg              *sync.WaitGroup
	connectionID    string
	allocRenameMask uint32
	Consensus
}

func (req *AllocRenameRequest) allocationRenameInBlobber(blobber *blockchain.StorageNode, blobberIdx int) (*Allocation, error) {
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)

	_ = formWriter.WriteField("connection_id", req.connectionID)
	formWriter.WriteField("name", req.name)
	formWriter.Close()

	httpreq, err := zboxutil.NewAllocRenameRequest(blobber.Baseurl, req.allocationTx, body)
	if err != nil {
		Logger.Error("Allocation Rename request error: ", err)
		return nil, err
	}
	httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())

	var alloc *Allocation

	ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 30))
	err = zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("Allocation Rename : ", err)
			return err
		}
		defer resp.Body.Close()
		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "Error: Resp")
		}
		Logger.Debug("Allocation Rename result:", string(resp_body))
		if resp.StatusCode == http.StatusOK {
			alloc = &Allocation{}
			err = json.Unmarshal(resp_body, alloc)
			if err != nil {
				return errors.Wrap(err, "allocation parsing error:")
			}
			req.consensus++

			return nil
		}

		return fmt.Errorf("error from server allocation rename")

	})

	return alloc, err
}

func (req *AllocRenameRequest) AllocationRename() error {
	numList := len(req.blobbers)
	objectTreeRefs := make([]Allocation, numList)
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numList)
	for i := 0; i < numList; i++ {
		go func(blobberIdx int) {
			defer req.wg.Done()
			alloc, err := req.allocationRenameInBlobber(req.blobbers[blobberIdx], blobberIdx)
			if err != nil {
				Logger.Error(err.Error())
				return
			}
			if alloc != nil {
				objectTreeRefs[blobberIdx] = *alloc
			}
		}(i)
	}
	req.wg.Wait()

	if !req.isConsensusOk() {
		return fmt.Errorf("Alloc Rename failed: allocation rename failed. Operation failed.")
	}

	return nil

}

func (req *AllocRenameRequest) commitAllocationRenameInBlobber(blobber *blockchain.StorageNode, blobberIdx int) (*Allocation, error) {
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)

	_ = formWriter.WriteField("connection_id", req.connectionID)
	formWriter.WriteField("name", req.name)
	formWriter.Close()

	httpreq, err := zboxutil.NewCommitAllocRenameRequest(blobber.Baseurl, req.allocationTx, body)
	if err != nil {
		Logger.Error("Commit Allocation Rename request error: ", err)
		return nil, err
	}
	httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())

	var alloc *Allocation

	ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 30))
	err = zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("Commit Allocation Rename : ", err)
			return err
		}
		defer resp.Body.Close()
		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "Error: Resp")
		}
		Logger.Debug("Commit Allocation Rename result:", string(resp_body))
		if resp.StatusCode == http.StatusOK {
			alloc = &Allocation{}
			err = json.Unmarshal(resp_body, alloc)
			if err != nil {
				return errors.Wrap(err, "allocation parsing error:")
			}
			req.consensus++

			return nil
		}

		return fmt.Errorf("error from server commit allocation rename")

	})

	return alloc, err
}

func (req *AllocRenameRequest) CommitAllocationRename() error {
	numList := len(req.blobbers)
	objectTreeRefs := make([]Allocation, numList)
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numList)
	for i := 0; i < numList; i++ {
		go func(blobberIdx int) {
			defer req.wg.Done()
			alloc, err := req.commitAllocationRenameInBlobber(req.blobbers[blobberIdx], blobberIdx)
			if err != nil {
				Logger.Error(err.Error())
				return
			}
			if alloc != nil {
				objectTreeRefs[blobberIdx] = *alloc
			}
		}(i)
	}
	req.wg.Wait()

	if !req.isConsensusOk() {
		return fmt.Errorf("Commit Alloc Rename failed: allocation rename failed. Operation failed.")
	}

	return nil

}
