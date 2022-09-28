package sdk

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"errors"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type DeleteRequest struct {
	allocationObj  *Allocation
	allocationID   string
	allocationTx   string
	blobbers       []*blockchain.StorageNode
	remoteFilePath string
	ctx            context.Context

	connectionID string
	consensus    Consensus
}

type DeleteResult struct {
	BlobberIndex int
	FileRef      fileref.RefEntity
	Deleted      bool
}

func (req *DeleteRequest) deleteBlobberFile(blobber *blockchain.StorageNode) error {

	query := &url.Values{}

	query.Add("connection_id", req.connectionID)
	query.Add("path", req.remoteFilePath)

	httpreq, err := zboxutil.NewDeleteRequest(blobber.Baseurl, req.allocationTx, query)
	if err != nil {
		l.Logger.Error(blobber.Baseurl, "Error creating delete request", err)
		return err
	}

	ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 30))
	return zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			l.Logger.Error("Delete : ", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			l.Logger.Info(blobber.Baseurl, " "+req.remoteFilePath, " deleted.")
			return nil
		}

		if resp.StatusCode == http.StatusNoContent {
			l.Logger.Info(blobber.Baseurl, " "+req.remoteFilePath, " not available in blobber.")
			return nil
		}

		resp_body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			l.Logger.Error(blobber.Baseurl, "Response: ", string(resp_body))
		}

		return fmt.Errorf("delete: %s", resp.Status)
	})

}

func (req *DeleteRequest) getObjectTreeFromBlobber(blobber *blockchain.StorageNode) (fileref.RefEntity, error) {
	return getObjectTreeFromBlobber(req.ctx, req.allocationID, req.allocationTx, req.remoteFilePath, blobber)
}

func (req *DeleteRequest) deleteFileFromBlobber(b *blockchain.StorageNode) (fileref.RefEntity, error) {

	refEntity, err := req.getObjectTreeFromBlobber(b)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			req.consensus.Done()
			return nil, nil
		}

		return nil, err
	}

	err = req.deleteBlobberFile(b)
	if err != nil {
		return nil, err
	}

	req.consensus.Done()
	return refEntity, nil
}

func (req *DeleteRequest) ProcessDelete() error {
	num := len(req.blobbers)
	numNotFound := 0
	objectTreeRefs := make([]fileref.RefEntity, num)

	wait := make(chan DeleteResult, num)

	wg := sync.WaitGroup{}
	wg.Add(num)

	for i := 0; i < num; i++ {
		go func(blobberIdx int) {
			defer wg.Done()

			fr, err := req.deleteFileFromBlobber(req.blobbers[blobberIdx])
			if err == nil {

				wait <- DeleteResult{
					BlobberIndex: blobberIdx,
					FileRef:      fr,
					Deleted:      true,
				}

				return
			}

			//it was removed from the blobber
			if errors.Is(err, constants.ErrNotFound) {

				wait <- DeleteResult{
					BlobberIndex: blobberIdx,
					FileRef:      nil,
					Deleted:      true,
				}

				return
			}

			wait <- DeleteResult{
				BlobberIndex: blobberIdx,
			}

			l.Logger.Error(err.Error())
		}(i)
	}

	wg.Wait()

	for i := 0; i < num; i++ {
		r := <-wait

		if !r.Deleted {
			continue
		}

		// it was deleted
		if r.FileRef == nil {
			numNotFound++
			continue
		}

		objectTreeRefs[r.BlobberIndex] = r.FileRef
	}

	if !req.consensus.isConsensusOk() {
		return fmt.Errorf("Delete failed: Success_rate:%2f, expected:%2f", req.consensus.getConsensusRate(), req.consensus.getConsensusRequiredForOk())
	}

	writeMarkerMutex, err := CreateWriteMarkerMutex(client.GetClient(), req.allocationObj)
	if err != nil {
		return fmt.Errorf("Delete failed: %s", err.Error())
	}
	err = writeMarkerMutex.Lock(context.TODO(), req.connectionID)
	defer writeMarkerMutex.Unlock(context.TODO(), req.connectionID) //nolint: errcheck
	if err != nil {
		return fmt.Errorf("Delete failed: %s", err.Error())
	}

	req.consensus.Reset()
	req.consensus.consensus = float32(numNotFound)

	commitReqs := make([]*CommitRequest, 0, numNotFound)

	for pos, ref := range objectTreeRefs {
		if ref == nil {
			continue
		}
		wg.Add(1)
		commitReq := &CommitRequest{}
		commitReq.allocationID = req.allocationID
		commitReq.allocationTx = req.allocationTx
		commitReq.blobber = req.blobbers[pos]
		newChange := &allocationchange.DeleteFileChange{}
		newChange.ObjectTree = ref
		newChange.NumBlocks = newChange.ObjectTree.GetNumBlocks()
		newChange.Operation = constants.FileOperationDelete
		newChange.Size = newChange.ObjectTree.GetSize()
		commitReq.changes = append(commitReq.changes, newChange)
		commitReq.connectionID = req.connectionID
		commitReq.wg = &wg
		commitReqs = append(commitReqs, commitReq)
		go AddCommitRequest(commitReq)
	}
	wg.Wait()

	for _, commitReq := range commitReqs {
		if commitReq.result != nil {
			if commitReq.result.Success {
				l.Logger.Info("Commit success", commitReq.blobber.Baseurl)
				req.consensus.Done()
			} else {
				l.Logger.Info("Commit failed", commitReq.blobber.Baseurl, commitReq.result.ErrorMessage)
			}
		} else {
			l.Logger.Info("Commit result not set", commitReq.blobber.Baseurl)
		}
	}

	if !req.consensus.isConsensusOk() {
		return errors.New("Delete failed: Commit consensus failed")
	}
	return nil
}
