package sdk

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"time"

	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)


func getObjectTreeFromBlobber(ctx context.Context, allocationID string, remotefilepath string, blobber *blockchain.StorageNode) (fileref.RefEntity, error) {
	httpreq, err := zboxutil.NewObjectTreeRequest(blobber.Baseurl, allocationID, remotefilepath)
	if err != nil {
		Logger.Error(blobber.Baseurl, "Error creating object tree request", err)
		return nil, err
	}
	var lR ReferencePathResult
	ctx, cncl := context.WithTimeout(ctx, (time.Second * 30))
	err = zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("Object tree:", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			Logger.Error("Object tree response : ", resp.StatusCode)
		}
		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Logger.Error("Object tree: Resp", err)
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("Object tree error response: Status: %d - %s ", resp.StatusCode, string(resp_body))
		} else {
			Logger.Info("Object tree:", string(resp_body))
			err = json.Unmarshal(resp_body, &lR)
			if err != nil {
				Logger.Error("Object tree json decode error: ", err)
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