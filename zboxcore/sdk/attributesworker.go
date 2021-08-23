package sdk

import (
	"context"
	"math/bits"
	"sync"

	blobbergrpc "github.com/0chain/blobber/code/go/0chain.net/blobbercore/blobbergrpc/proto"
	"github.com/0chain/gosdk/core/clients/blobberClient"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/zboxcore/fileref"

	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"

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

	return getObjectTreeFromBlobber(ar.allocationID, ar.allocationTx, ar.remotefilepath, blobber)
}

func (ar *AttributesRequest) updateBlobberObjectAttributes(
	blobber *blockchain.StorageNode, blobberIdx int) (
	re fileref.RefEntity, err error) {

	re, err = ar.getObjectTreeFromBlobber(ar.blobbers[blobberIdx])
	if err != nil {
		return
	}

	_, err = blobberClient.UpdateObjectAttributes(blobber.Baseurl, &blobbergrpc.UpdateObjectAttributesRequest{
		Path:         ar.remotefilepath,
		Allocation:   ar.allocationTx,
		ConnectionId: ar.connectionID,
		Attributes:   ar.attributes,
	})
	if err != nil {
		Logger.Error("could not update object attributes from blobber -" + blobber.Baseurl + " - " + err.Error())
		err = errors.Wrap(err, "update attribute failed")
		return
	}

	ar.consensus++
	ar.attributesMask |= (1 << uint32(blobberIdx))
	Logger.Info(blobber.Baseurl, " "+ar.remotefilepath,
		" attributes updated.")
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
		c, pos     = 0, 0
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
		change.Operation = allocationchange.UPDATE_ATTRS_OPERATION
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
