package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

func getObjectTreeFromBlobber(ctx context.Context, allocationID, allocationTx, sig string, remoteFilePath string, blobber *blockchain.StorageNode) (fileref.RefEntity, error) {
	httpreq, err := zboxutil.NewObjectTreeRequest(blobber.Baseurl, allocationID, allocationTx, sig, remoteFilePath)
	if err != nil {
		l.Logger.Error(blobber.Baseurl, "Error creating object tree request", err)
		return nil, err
	}
	var lR ReferencePathResult
	ctx, cncl := context.WithTimeout(ctx, (time.Second * 60))
	err = zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			l.Logger.Error("Object tree:", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			l.Logger.Error("Object tree response : ", resp.StatusCode)
		}
		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			l.Logger.Error("Object tree: Resp", err)
			return err
		}
		if resp.StatusCode != http.StatusOK {
			if resp.StatusCode == http.StatusNotFound {
				return errors.Throw(constants.ErrNotFound, remoteFilePath)
			}
			return errors.New(strconv.Itoa(resp.StatusCode), fmt.Sprintf("Object tree error response: Body: %s ", string(resp_body)))
		} else {
			err = json.Unmarshal(resp_body, &lR)
			if err != nil {
				l.Logger.Error("Object tree json decode error: ", err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return lR.GetRefFromObjectTree(allocationID)
}

func getAllocationDataFromBlobber(blobber *blockchain.StorageNode, allocationId string, allocationTx string, respCh chan<- *BlobberAllocationStats, wg *sync.WaitGroup) {
	defer wg.Done()
	httpreq, err := zboxutil.NewAllocationRequest(blobber.Baseurl, allocationId, allocationTx)
	if err != nil {
		l.Logger.Error(blobber.Baseurl, "Error creating allocation request", err)
		return
	}

	var result BlobberAllocationStats
	ctx, cncl := context.WithTimeout(context.Background(), (time.Second * 30))
	err = zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			l.Logger.Error("Get allocation :", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			l.Logger.Error("Get allocation response : ", resp.StatusCode)
		}
		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			l.Logger.Error("Get allocation: Resp", err)
			return err
		}

		err = json.Unmarshal(resp_body, &result)
		if err != nil {
			l.Logger.Error("Object tree json decode error: ", err)
			return err
		}
		return nil
	})
	if err != nil {
		return
	}

	result.BlobberID = blobber.ID
	result.BlobberURL = blobber.Baseurl
	respCh <- &result
}

type ProcessResult struct {
	BlobberIndex int
	FileRef      fileref.RefEntity
	Succeed      bool
}

var ErrFileNameTooLong = errors.New("invalid_parameter", "filename is longer than 150 characters")

func ValidateRemoteFileName(remotePath string) error {
	_, fileName := path.Split(remotePath)

	if len(fileName) > 150 {
		return ErrFileNameTooLong
	}

	return nil
}
