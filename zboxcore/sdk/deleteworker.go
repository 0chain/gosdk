package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/bits"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type DeleteRequest struct {
	allocationID   string
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

func (req *DeleteRequest) deleteBlobberFile(blobber *blockchain.StorageNode, blobberIdx int, listResponse *fileMetaResponse) {
	defer req.wg.Done()
	path, _ := filepath.Split(req.remotefilepath)
	if path != "/" {
		path = strings.TrimRight(path, "/")
	}

	if listResponse == nil || listResponse.fileref == nil {
		Logger.Error(blobber.Baseurl, req.remotefilepath, " File not found")
		return
	}
	dt := &marker.DeleteToken{}
	dt.FilePathHash = listResponse.fileref.PathHash
	dt.FileRefHash = listResponse.fileref.Hash
	dt.AllocationID = req.allocationID
	dt.Size = listResponse.fileref.Size + listResponse.fileref.ThumbnailSize
	dt.BlobberID = blobber.ID
	dt.Timestamp = common.Now()
	dt.ClientID = client.GetClientID()
	err := dt.Sign()
	if err != nil {
		Logger.Error(blobber.Baseurl, " Signing delete token", err)
		return
	}
	dtData, err := json.Marshal(dt)
	if err != nil {
		Logger.Error(blobber.Baseurl, " Creating json delete token", err)
		return
	}
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)
	// formData := deleteFormData{
	// 	ConnectionID: req.connectionID,
	// 	Filename:     listResponse.fileref.Name,
	// 	Path:         listResponse.fileref.Path,
	// }
	// var metaData []byte
	// metaData, err = json.Marshal(formData)
	// if err != nil {
	// 	Logger.Error(blobber.Baseurl, " creating delete formdata", err)
	// 	return
	// }
	//formWriter.WriteField("uploadMeta", string(metaData))
	_ = formWriter.WriteField("connection_id", req.connectionID)
	formWriter.WriteField("delete_token", string(dtData))
	formWriter.Close()
	httpreq, err := zboxutil.NewDeleteRequest(blobber.Baseurl, req.allocationID, body)
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

func (req *DeleteRequest) ProcessDelete() error {
	listReq := &ListRequest{remotefilepath: req.remotefilepath, allocationID: req.allocationID, blobbers: req.blobbers, ctx: req.ctx}
	var listResponses []*fileMetaResponse
	req.listMask, _, listResponses = listReq.getFileConsensusFromBlobbers()
	if req.listMask == 0 || len(listResponses) == 0 {
		return fmt.Errorf("No minimum consensus for file meta data of file")
	}
	req.deleteMask = uint32(0)
	req.consensus = 0
	numDeletes := bits.OnesCount32(req.listMask)
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numDeletes)

	c, pos := 0, 0
	for i := req.listMask; i != 0; i &= ^(1 << uint32(pos)) {
		pos = bits.TrailingZeros32(i)
		go req.deleteBlobberFile(req.blobbers[pos], pos, listResponses[pos])
		//go obj.downloadBlobberBlock(&obj.blobbers[pos], pos, path, blockNum, rspCh, isPathHash, authTicket)
		c++
	}
	req.wg.Wait()
	// if !req.isConsensusOk() {
	// 	return fmt.Errorf("Delete failed: Success_rate:%2f, expected:%2f", req.getConsensusRate(), req.getConsensusRequiredForOk())
	// }

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
		commitReq.blobber = req.blobbers[pos]
		newChange := &allocationchange.DeleteFileChange{}
		newChange.File = listResponses[pos].fileref
		newChange.NumBlocks = newChange.File.NumBlocks
		newChange.Operation = allocationchange.DELETE_OPERATION
		newChange.Size = newChange.File.Size
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
		return fmt.Errorf("Upload failed: Commit consensus failed")
	}
	return nil
}
