package sdk

import (
	"context"
	"math/bits"
	"sync"

	blobbergrpc "github.com/0chain/blobber/code/go/0chain.net/blobbercore/blobbergrpc/proto"
	"github.com/0chain/gosdk/core/clients/blobberClient"
	"github.com/0chain/gosdk/core/common/errors"
	"github.com/0chain/gosdk/zboxcore/fileref"

	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	. "github.com/0chain/gosdk/zboxcore/logger"
)

type RenameRequest struct {
	allocationID   string
	allocationTx   string
	blobbers       []*blockchain.StorageNode
	remotefilepath string
	newName        string
	ctx            context.Context
	wg             *sync.WaitGroup
	renameMask     uint32
	connectionID   string
	Consensus
}

func (req *RenameRequest) getObjectTreeFromBlobber(blobber *blockchain.StorageNode) (fileref.RefEntity, error) {
	return getObjectTreeFromBlobber(req.allocationID, req.allocationTx, req.remotefilepath, blobber)
}

func (req *RenameRequest) renameBlobberObject(blobber *blockchain.StorageNode, blobberIdx int) (fileref.RefEntity, error) {
	refEntity, err := req.getObjectTreeFromBlobber(req.blobbers[blobberIdx])
	if err != nil {
		return nil, err
	}

	err = blobberClient.RenameObject(blobber.Baseurl, &blobbergrpc.RenameObjectRequest{
		Path:         req.remotefilepath,
		Allocation:   req.allocationTx,
		ConnectionId: req.connectionID,
		NewName:      req.newName,
	})
	if err != nil {
		Logger.Error("could not rename object from blobber -" + blobber.Baseurl + " - " + err.Error())
		return nil, err
	}

	req.consensus++
	req.renameMask |= (1 << uint32(blobberIdx))
	Logger.Info(blobber.Baseurl, " "+req.remotefilepath, " renamed.")

	return refEntity, nil
}

func (req *RenameRequest) ProcessRename() error {
	numList := len(req.blobbers)
	objectTreeRefs := make([]fileref.RefEntity, numList)
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numList)
	for i := 0; i < numList; i++ {
		go func(blobberIdx int) {
			defer req.wg.Done()
			refEntity, err := req.renameBlobberObject(req.blobbers[blobberIdx], blobberIdx)
			if err != nil {
				Logger.Error(err.Error())
				return
			}
			objectTreeRefs[blobberIdx] = refEntity
		}(i)
	}
	req.wg.Wait()

	if !req.isConsensusOk() {
		return errors.New("Rename failed: Rename request failed. Operation failed.")
	}

	req.consensus = 0
	wg := &sync.WaitGroup{}
	wg.Add(bits.OnesCount32(req.renameMask))
	commitReqs := make([]*CommitRequest, bits.OnesCount32(req.renameMask))
	c, pos := 0, 0
	for i := req.renameMask; i != 0; i &= ^(1 << uint32(pos)) {
		pos = bits.TrailingZeros32(i)
		//go req.prepareUpload(a, a.Blobbers[pos], req.file[c], req.uploadDataCh[c], req.wg)
		commitReq := &CommitRequest{}
		commitReq.allocationID = req.allocationID
		commitReq.allocationTx = req.allocationTx
		commitReq.blobber = req.blobbers[pos]
		newChange := &allocationchange.RenameFileChange{}
		newChange.NewName = req.newName
		newChange.ObjectTree = objectTreeRefs[pos]
		newChange.NumBlocks = 0
		newChange.Operation = allocationchange.RENAME_OPERATION
		newChange.Size = 0
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
