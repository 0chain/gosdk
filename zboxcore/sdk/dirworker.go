package sdk

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/0chain/gosdk/zboxcore/blockchain"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type DirRequest struct {
	allocationID string
	name         string
	ctx          context.Context
	action       string // create, del
	connectionID string
	wg           *sync.WaitGroup
	Consensus
}

func (req *DirRequest) ProcessDir(a *Allocation) error {
	numList := len(a.Blobbers)
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numList)

	l.Logger.Info("Start creating dir for blobbers")

	for i := 0; i < numList; i++ {
		go func(blobberIdx int) {
			defer req.wg.Done()

			err := req.createDirInBlobber(a.Blobbers[blobberIdx])
			if err != nil {
				l.Logger.Error(err.Error())
				return
			}
		}(i)
	}

	req.wg.Wait()

	if !req.isConsensusOk() {
		return errors.New("Directory creation failed due to consensus not met")
	}

	return nil
}

func (req *DirRequest) createDirInBlobber(blobber *blockchain.StorageNode) error {
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)
	formWriter.WriteField("connection_id", req.connectionID)

	dirPath := filepath.ToSlash(filepath.Join("/", req.name))
	formWriter.WriteField("dir_path", dirPath)

	formWriter.Close()
	httpreq, err := zboxutil.NewCreateDirRequest(blobber.Baseurl, req.allocationID, body)
	if err != nil {
		l.Logger.Error(blobber.Baseurl, "Error creating dir request", err)
		return err
	}

	httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
	ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 30))

	err = zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			l.Logger.Error("createdir : ", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			resp_body, _ := ioutil.ReadAll(resp.Body)
			l.Logger.Info("createdir resp:", string(resp_body))
			req.consensus++
			l.Logger.Info(blobber.Baseurl, " "+req.name, " created.")
		} else {
			resp_body, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				msg := strings.TrimSpace(string(resp_body))
				l.Logger.Error(blobber.Baseurl, "Response: ", msg)
				return errors.New(msg)
			}
		}
		return err
	})

	if err != nil {
		return err
	}

	return nil
}
