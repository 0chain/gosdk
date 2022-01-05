package sdk

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"github.com/0chain/gosdk/zboxcore/blockchain"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"golang.org/x/sync/errgroup"
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
	errs := new(errgroup.Group)
	for i := 0; i < numList; i++ {
		i := i
		errs.Go(func() error {
			err := req.createDirInBlobber(a.Blobbers[i])
			if err != nil {
				Logger.Error(err.Error())
				return err
			}
			return err
		})
	}
	err := errs.Wait()

	if err == nil {
		Logger.Info("Directory created successfully.")
	}

	return err
}

func (req *DirRequest) createDirInBlobber(blobber *blockchain.StorageNode) error {
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)
	formWriter.WriteField("connection_id", req.connectionID)
	formWriter.WriteField("dir_path", filepath.ToSlash(filepath.Join("/", req.name)))
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
				Logger.Error(blobber.Baseurl, "Response: ", string(resp_body))
				return errors.New(string(resp_body))
			}
		}
		return err
	})

	if err != nil {
		return err
	}

	return nil
}
