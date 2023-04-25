package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"time"

	"net/http"

	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type LatestPrevWriteMarker struct {
	LatestWM *marker.WriteMarker `json:"latest_write_marker"`
	PrevWM   *marker.WriteMarker `json:"prev_write_marker"`
}

type RollbackBlobber struct {
	blobber      *blockchain.StorageNode
	commitResult *CommitResult
	lpm          *LatestPrevWriteMarker
}

func GetWritemarker(allocID, id, baseUrl string) (*LatestPrevWriteMarker, error) {

	var lpm LatestPrevWriteMarker

	req, err := zboxutil.NewWritemarkerRequest(baseUrl, allocID)
	if err != nil {
		return nil, err
	}

	resp, err := zboxutil.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("writemarker error response %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &lpm)
	if err != nil {
		return nil, err
	}

	return &lpm, nil
}

func (rb *RollbackBlobber) processRollback(ctx context.Context, tx string) error {

	wm := &marker.WriteMarker{}
	wm.AllocationID = rb.lpm.LatestWM.AllocationID
	wm.Timestamp = rb.lpm.LatestWM.Timestamp
	wm.BlobberID = rb.lpm.LatestWM.BlobberID
	wm.ClientID = client.GetClientID()
	wm.Size = 0
	if rb.lpm.PrevWM != nil {
		wm.AllocationRoot = rb.lpm.PrevWM.AllocationRoot
		wm.PreviousAllocationRoot = rb.lpm.PrevWM.AllocationRoot
		wm.FileMetaRoot = rb.lpm.PrevWM.FileMetaRoot
	}
	err := wm.Sign()
	if err != nil {
		logger.Logger.Error("Signing writemarker failed: ", err)
		return err
	}
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)
	wmData, err := json.Marshal(wm)
	if err != nil {
		logger.Logger.Error("Creating writemarker failed: ", err)
		return err
	}
	connID := zboxutil.NewConnectionId()
	formWriter.WriteField("write_marker", string(wmData))
	formWriter.WriteField("connection_id", connID)
	formWriter.Close()

	req, err := zboxutil.NewRollbackRequest(rb.blobber.Baseurl, tx, body)
	if err != nil {
		logger.Logger.Error("Creating rollback request failed: ", err)
		return err
	}
	req.Header.Add("Content-Type", formWriter.FormDataContentType())

	logger.Logger.Info("Sending Rollback request to blobber: ", rb.blobber.Baseurl)

	var (
		resp           *http.Response
		shouldContinue bool
	)

	for retries := 0; retries < 3; retries++ {
		err, shouldContinue = func() (err error, shouldContinue bool) {
			reqCtx, ctxCncl := context.WithTimeout(ctx, DefaultUploadTimeOut)
			resp, err := zboxutil.Client.Do(req.WithContext(reqCtx))
			defer ctxCncl()
			if err != nil {
				logger.Logger.Error("Rollback request failed: ", err)
				return
			}

			if resp.Body != nil {
				defer resp.Body.Close()
			}

			var respBody []byte
			if resp.StatusCode == http.StatusOK {
				logger.Logger.Info(rb.blobber.Baseurl, connID, "rollbacked")
				return
			}

			if resp.StatusCode == http.StatusTooManyRequests {
				logger.Logger.Info(rb.blobber.Baseurl, connID, "got too many request error. Retrying")
				var r int
				r, err = zboxutil.GetRateLimitValue(resp)
				if err != nil {
					logger.Logger.Error(err)
					return
				}

				time.Sleep(time.Duration(r) * time.Second)
				shouldContinue = true
				return
			}

			respBody, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Logger.Error("Response read: ", err)
				return
			}

			err = thrown.New("commit_error",
				fmt.Sprintf("Got error response %s with status %d", respBody, resp.StatusCode))

			return
		}()
		if err != nil {
			logger.Logger.Error(err)
			return err
		}
		if shouldContinue {
			continue
		}
		return nil

	}

	return thrown.New("rolback_error", fmt.Sprintf("Rollback failed with response status %d", resp.StatusCode))
}
