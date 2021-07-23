package sdk

import (
	blobbergrpc "github.com/0chain/blobber/code/go/0chain.net/blobbercore/blobbergrpc/proto"
	"net/http"
	"sync"

	"github.com/0chain/gosdk/core/clients/blobberClient"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	. "github.com/0chain/gosdk/zboxcore/logger"
)

type CollaboratorRequest struct {
	a              *Allocation
	path           string
	collaboratorID string
	wg             *sync.WaitGroup
}

func (req *CollaboratorRequest) CollaboratorHandler(blobber *blockchain.StorageNode, blobberIdx int, method string, rspCh chan<- bool) {
	defer req.wg.Done()

	_, err := blobberClient.Collaborator(blobber.Baseurl, &blobbergrpc.CollaboratorRequest{
		Allocation: req.a.Tx,
		CollabId:   req.collaboratorID,
		Method:     method,
		Path:       req.path,
	})
	if err != nil {
		Logger.Error("CollaboratorHandler Error: ", err)
		rspCh <- false
		return
	}

	rspCh <- true
	return
}

func (req *CollaboratorRequest) UpdateCollaboratorToBlobbers() bool {
	numList := len(req.a.Blobbers)
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numList)
	rspCh := make(chan bool, numList)
	for i := 0; i < numList; i++ {
		go req.CollaboratorHandler(req.a.Blobbers[i], i, http.MethodPost, rspCh)
	}
	req.wg.Wait()
	count := 0
	for i := 0; i < numList; i++ {
		resp := <-rspCh
		if resp {
			count++
		}
	}
	return count == numList
}

func (req *CollaboratorRequest) RemoveCollaboratorFromBlobbers() bool {
	numList := len(req.a.Blobbers)
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numList)
	rspCh := make(chan bool, numList)
	for i := 0; i < numList; i++ {
		go req.CollaboratorHandler(req.a.Blobbers[i], i, http.MethodDelete, rspCh)
	}
	req.wg.Wait()
	count := 0
	for i := 0; i < numList; i++ {
		resp := <-rspCh
		if resp {
			count++
		}
	}
	return count == numList
}
