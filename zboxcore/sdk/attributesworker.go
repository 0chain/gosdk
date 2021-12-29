package sdk

import (
	"bytes"
	"context"
	"io/ioutil"
	"math/bits"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	"errors"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/zboxcore/fileref"

	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/zboxutil"

	. "github.com/0chain/gosdk/zboxcore/logger"
)

type AttributesRequest struct {
	allocationID   string                    //
	allocationTx   string                    //
	blobbers       []*blockchain.StorageNode //
	remotefilepath string                    // path (not hash)
	Attributes     fileref.Attributes        // new attributes
	attributes     string                    // new attributes (JSON)
	attributesMask uint32                    //
	connectionID   string                    //
	Consensus                                //
	ctx            context.Context           //
	wg             *sync.WaitGroup           //
}

func (ar *AttributesRequest) getObjectTreeFromBlobber(
	blobber *blockchain.StorageNode) (fileref.RefEntity, error) {

	return getObjectTreeFromBlobber(ar.ctx, ar.allocationID, ar.allocationTx,
		ar.remotefilepath, blobber)
}

func (ar *AttributesRequest) updateBlobberObjectAttributes(
	blobber *blockchain.StorageNode, blobberIdx int) (
	re fileref.RefEntity, err error) {

	re, err = ar.getObjectTreeFromBlobber(ar.blobbers[blobberIdx])
	if err != nil {
		return
	}

	var (
		body bytes.Buffer
		form = multipart.NewWriter(&body)
	)

	form.WriteField("connection_id", ar.connectionID)
	form.WriteField("path", ar.remotefilepath)
	form.WriteField("attributes", ar.attributes)

	form.Close()

	var httpreq *http.Request
	httpreq, err = zboxutil.NewAttributesRequest(blobber.Baseurl,
		ar.allocationTx, &body)
	if err != nil {
		Logger.Error(blobber.Baseurl,
			"Error creating update attributes request", err)
		return
	}

	httpreq.Header.Add("Content-Type", form.FormDataContentType())

	var ctx, cncl = context.WithTimeout(ar.ctx, (time.Second * 30))
	defer cncl()

	err = zboxutil.HttpDo(ctx, cncl, httpreq,
		func(resp *http.Response, err error) error {
			if err != nil {
				Logger.Error("Request error: ", err)
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				ar.consensus++
				ar.attributesMask |= (1 << uint32(blobberIdx))
				Logger.Info(blobber.Baseurl, " "+ar.remotefilepath,
					" attributes updated.")
				return nil
			}

			var respBody []byte
			if respBody, err = ioutil.ReadAll(resp.Body); err != nil {
				Logger.Error(blobber.Baseurl, "Reading response: ", err)
				return nil
			}

			Logger.Error(blobber.Baseurl, "Response error: ",
				string(respBody))
			return nil
		})

	return
}

func (ar *AttributesRequest) ProcessAttributes() (err error) {

	var (
		numList        = len(ar.blobbers)
		objectTreeRefs = make([]fileref.RefEntity, numList)
	)

	ar.wg = &sync.WaitGroup{}
	ar.wg.Add(numList)

	for i := 0; i < numList; i++ {
		go func(bidx int) {
			defer ar.wg.Done()
			var re, err = ar.updateBlobberObjectAttributes(ar.blobbers[bidx],
				bidx)
			if err != nil {
				Logger.Error(err.Error())
				return
			}
			objectTreeRefs[bidx] = re
		}(i)
	}
	ar.wg.Wait()

	if !ar.isConsensusOk() {
		return errors.New("Update attributes failed: request failed, operation failed")
	}

	ar.consensus = 0

	var wg sync.WaitGroup
	wg.Add(bits.OnesCount32(ar.attributesMask))

	var (
		commitReqs = make([]*CommitRequest, bits.OnesCount32(ar.attributesMask))
		c, pos     int
	)

	for i := ar.attributesMask; i != 0; i &= ^(1 << uint32(pos)) {
		pos = bits.TrailingZeros32(i)
		var commitReq CommitRequest
		commitReq.allocationID = ar.allocationID
		commitReq.allocationTx = ar.allocationTx
		commitReq.blobber = ar.blobbers[pos]
		var change = new(allocationchange.AttributesChange)
		change.AllocationID = ar.allocationID
		change.ConnectionID = ar.connectionID
		change.Path = ar.remotefilepath
		change.Attributes = ar.Attributes
		change.NumBlocks = 0
		change.Size = 0
		change.Operation = constants.FileOperationUpdateAttrs
		commitReq.changes = append(commitReq.changes, change)
		commitReq.connectionID = ar.connectionID
		commitReq.wg = &wg
		commitReqs[c] = &commitReq
		go AddCommitRequest(&commitReq)
		c++
	}
	wg.Wait()

	for _, commitReq := range commitReqs {
		if commitReq.result != nil {
			if commitReq.result.Success {
				Logger.Info("Commit success", commitReq.blobber.Baseurl)
				ar.consensus++
			} else {
				Logger.Info("Commit failed", commitReq.blobber.Baseurl,
					commitReq.result.ErrorMessage)
			}
		} else {
			Logger.Info("Commit result not set", commitReq.blobber.Baseurl)
		}
	}

	if !ar.isConsensusOk() {
		return errors.New("Delete failed: Commit consensus failed")
	}

	return nil
}
