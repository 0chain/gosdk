package sdk

import (
	"bytes"
	"context"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	"github.com/0chain/gosdk/zboxcore/blockchain"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type DirRequest struct {
	allocationID string
	name         string
	ctx          context.Context
	action       string // create, del
	wg           *sync.WaitGroup
	connectionID string
	Consensus
}

func (req *DirRequest) ProcessDir(a *Allocation) error {
	numList := len(a.Blobbers)
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numList)

	Logger.Info("Start creating dir for blobbers")
	for i := 0; i < numList; i++ {
		go func(blobberIdx int) {
			defer req.wg.Done()
			err := req.createDirInBlobber(a.Blobbers[blobberIdx])
			if err != nil {
				Logger.Error(err.Error())
				return
			}
		}(i)
	}
	req.wg.Wait()

	return nil
}

func (req *DirRequest) createDirInBlobber(blobber *blockchain.StorageNode) error {
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)
	formWriter.WriteField("connection_id", req.connectionID)
	formWriter.WriteField("dir_path", req.name)
	formWriter.Close()
	httpreq, err := zboxutil.NewCreateDirRequest(blobber.Baseurl, req.allocationID, body)
	if err != nil {
		Logger.Error(blobber.Baseurl, "Error creating dir request", err)
		return err
	}
	Logger.Debug(httpreq.URL)

	httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
	ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 30))

	err = zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("Copy : ", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			resp_body, _ := ioutil.ReadAll(resp.Body)
			Logger.Info("copy resp:", string(resp_body))
			req.consensus++
			Logger.Info(blobber.Baseurl, " "+req.name, " created.")
		} else {
			resp_body, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				Logger.Error(blobber.Baseurl, "Response: ", string(resp_body))
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
