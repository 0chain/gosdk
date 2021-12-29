package sdk

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	"github.com/0chain/gosdk/zboxcore/blockchain"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type CollaboratorRequest struct {
	a              *Allocation
	path           string
	collaboratorID string
	wg             *sync.WaitGroup
}

func (req *CollaboratorRequest) UpdateCollaboratorToBlobbers() bool {
	numList := len(req.a.Blobbers)
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numList)
	rspCh := make(chan bool, numList)
	for i := 0; i < numList; i++ {
		go req.updateCollaboratorToBlobber(req.a.Blobbers[i], i, rspCh)
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

func (req *CollaboratorRequest) updateCollaboratorToBlobber(blobber *blockchain.StorageNode, blobberIdx int, rspCh chan<- bool) {

	defer req.wg.Done()
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)

	formWriter.WriteField("path", req.path)
	formWriter.WriteField("collab_id", req.collaboratorID)

	formWriter.Close()
	httpreq, err := zboxutil.NewCollaboratorRequest(blobber.Baseurl, req.a.Tx, body)
	if err != nil {
		Logger.Error("Update collaborator request error: ", err.Error())
		return
	}

	httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
	ctx, cncl := context.WithTimeout(req.a.ctx, (time.Second * 30))
	zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("Update Collaborator : ", err)
			rspCh <- false
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			rspCh <- true
		} else {
			rspCh <- false
		}
		return err
	})
}

func (req *CollaboratorRequest) RemoveCollaboratorFromBlobbers() bool {
	numList := len(req.a.Blobbers)
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numList)
	rspCh := make(chan bool, numList)
	for i := 0; i < numList; i++ {
		go req.removeCollaboratorFromBlobber(req.a.Blobbers[i], i, rspCh)
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

func (req *CollaboratorRequest) removeCollaboratorFromBlobber(blobber *blockchain.StorageNode, blobberIdx int, rspCh chan<- bool) {

	defer req.wg.Done()
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)

	formWriter.WriteField("path", req.path)
	formWriter.WriteField("collab_id", req.collaboratorID)

	formWriter.Close()
	httpreq, err := zboxutil.DeleteCollaboratorRequest(blobber.Baseurl, req.a.Tx, body)
	if err != nil {
		Logger.Error("Delete collaborator request error: ", err.Error())
		return
	}

	httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
	ctx, cncl := context.WithTimeout(req.a.ctx, (time.Second * 30))

	zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("Delete Collaborator : ", err)
			rspCh <- false
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			rspCh <- true
		} else {
			rspCh <- false
		}

		return err
	})
}
