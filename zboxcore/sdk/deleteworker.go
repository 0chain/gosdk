package sdk

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/bits"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"errors"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type DeleteRequest struct {
	allocationID   string
	allocationTx   string
	blobbers       []*blockchain.StorageNode
	remotefilepath string
	ctx            context.Context
	wg             *sync.WaitGroup
	listMask       uint32
	deleteMask     uint32
	connectionID   string
	Consensus
}

type deleteFormData struct {
	ConnectionID string `json:"connection_id"`
	Filename     string `json:"filename"`
	Path         string `json:"filepath"`
}

func (req *DeleteRequest) deleteBlobberFile(blobber *blockchain.StorageNode, blobberIdx int, objectTree fileref.RefEntity) {
	defer req.wg.Done()
	path, _ := filepath.Split(req.remotefilepath)
	if path != "/" {
		path = strings.TrimRight(path, "/")
	}

	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)

	_ = formWriter.WriteField("connection_id", req.connectionID)
	_ = formWriter.WriteField("path", req.remotefilepath)
	formWriter.Close()
	httpreq, err := zboxutil.NewDeleteRequest(blobber.Baseurl, req.allocationTx, body)
	if err != nil {
		Logger.Error(blobber.Baseurl, "Error creating delete request", err)
		return
	}
	httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
	ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 30))
	_ = zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("Delete : ", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			req.consensus++
			req.deleteMask |= (1 << uint32(blobberIdx))
			Logger.Info(blobber.Baseurl, " "+req.remotefilepath, " deleted.")
		} else {
			resp_body, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				Logger.Error(blobber.Baseurl, "Response: ", string(resp_body))
			}
		}
		return nil
	})
}

func (req *DeleteRequest) getObjectTreeFromBlobber(blobber *blockchain.StorageNode) (fileref.RefEntity, error) {
	return getObjectTreeFromBlobber(req.ctx, req.allocationID, req.allocationTx, req.remotefilepath, blobber)
}

func (req *DeleteRequest) ProcessDelete() error {
	numList := len(req.blobbers)
	objectTreeRefs := make([]fileref.RefEntity, numList)
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numList)
	for i := 0; i < numList; i++ {
		go func(blobberIdx int) {
			defer req.wg.Done()
			refEntity, err := req.getObjectTreeFromBlobber(req.blobbers[blobberIdx])
			if err != nil {
				Logger.Error(err.Error())
				return
			}
			req.consensus++
			req.listMask |= (1 << uint32(blobberIdx))
			objectTreeRefs[blobberIdx] = refEntity
		}(i)
	}
	req.wg.Wait()

	req.deleteMask = uint32(0)
	req.consensus = 0
	numDeletes := bits.OnesCount32(req.listMask)
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numDeletes)

	c, pos := 0, 0
	for i := req.listMask; i != 0; i &= ^(1 << uint32(pos)) {
		pos = bits.TrailingZeros32(i)
		go req.deleteBlobberFile(req.blobbers[pos], pos, objectTreeRefs[pos])
		//go obj.downloadBlobberBlock(&obj.blobbers[pos], pos, path, blockNum, rspCh, isPathHash, authTicket)
		c++
	}
	req.wg.Wait()

	if !req.isConsensusOk() {
		return fmt.Errorf("Delete failed: Success_rate:%2f, expected:%2f", req.getConsensusRate(), req.getConsensusRequiredForOk())
	}

	req.consensus = 0
	wg := &sync.WaitGroup{}
	wg.Add(bits.OnesCount32(req.deleteMask))
	commitReqs := make([]*CommitRequest, bits.OnesCount32(req.deleteMask))
	c, pos = 0, 0
	for i := req.deleteMask; i != 0; i &= ^(1 << uint32(pos)) {
		pos = bits.TrailingZeros32(i)
		//go req.prepareUpload(a, a.Blobbers[pos], req.file[c], req.uploadDataCh[c], req.wg)
		commitReq := &CommitRequest{}
		commitReq.allocationID = req.allocationID
		commitReq.allocationTx = req.allocationTx
		commitReq.blobber = req.blobbers[pos]
		newChange := &allocationchange.DeleteFileChange{}
		newChange.ObjectTree = objectTreeRefs[pos]
		newChange.NumBlocks = newChange.ObjectTree.GetNumBlocks()
		newChange.Operation = constants.FileOperationDelete
		newChange.Size = newChange.ObjectTree.GetSize()
		commitReq.changes = append(commitReq.changes, newChange)
		commitReq.connectionID = req.connectionID
		commitReq.wg = wg
		commitReqs[c] = commitReq
		go AddCommitRequest(commitReq)
		c++
	}
	wg.Wait()

	for _, commitReq := range commitReqs {
		if commitReq.result != nil {
			if commitReq.result.Success {
				Logger.Info("Commit success", commitReq.blobber.Baseurl)
				req.consensus++
			} else {
				Logger.Info("Commit failed", commitReq.blobber.Baseurl, commitReq.result.ErrorMessage)
			}
		} else {
			Logger.Info("Commit result not set", commitReq.blobber.Baseurl)
		}
	}

	if !req.isConsensusOk() {
		return errors.New("Delete failed: Commit consensus failed")
	}
	return nil
}
