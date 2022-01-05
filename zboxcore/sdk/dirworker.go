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
	connectionID string
	Consensus
}

func (req *DirRequest) ProcessDir(a *Allocation) error {
	numList := len(a.Blobbers)

	Logger.Info("Start creating dir for blobbers")

	await := make(chan error, numList)

	for i := 0; i < numList; i++ {
		go func(blobberIdx int) {
			err := req.createDirInBlobber(a.Blobbers[blobberIdx])
			if err != nil {
				Logger.Error(err.Error())
			}
			await <- err
		}(i)
	}

	msgList := make([]string, 0, numList)
	for i := 0; i < numList; i++ {
		err := <-await
		if err != nil {
			msgList = append(msgList, err.Error())
		}
	}

	if len(msgList) > 0 {
		return errors.New(strings.Join(msgList, ", "))
	}

	return nil
}

func (req *DirRequest) createDirInBlobber(blobber *blockchain.StorageNode) error {
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)
	formWriter.WriteField("connection_id", req.connectionID)

	dirPath := filepath.Join("/", filepath.ToSlash(req.name))
	formWriter.WriteField("dir_path", dirPath)

	formWriter.Close()
	httpreq, err := zboxutil.NewCreateDirRequest(blobber.Baseurl, req.allocationID, body)
	if err != nil {
		Logger.Error(blobber.Baseurl, "Error creating dir request", err)
		return err
	}

	httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
	ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 30))

	err = zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("createdir : ", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			resp_body, _ := ioutil.ReadAll(resp.Body)
			Logger.Info("createdir resp:", string(resp_body))
			req.consensus++
			Logger.Info(blobber.Baseurl, " "+req.name, " created.")
		} else {
			resp_body, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				msg := strings.TrimSpace(string(resp_body))
				Logger.Error(blobber.Baseurl, "Response: ", msg)
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
