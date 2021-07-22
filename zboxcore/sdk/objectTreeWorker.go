package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type ObjectTreeResult struct {
	Page          int64                    `json:"page"`
	TotalPages    int64                    `json:"total_pages"`
	NewOffsetPath string                   `json:"offsetPath"`
	Refs          []fileref.FileRef        `json:"refs"`
	LatestWM      *writemarker.WriteMarker `json:"latest_write_marker"`
}

type ObjectTreeRequest struct {
	allocationID       string
	allocationTx       string
	blobbers           []*blockchain.StorageNode
	remotefilepathhash string
	remotefilepath     string
	page               int
	offsetPath         string
	authToken          *marker.AuthTicket
	ctx                context.Context
	wg                 *sync.WaitGroup
	Consensus
}

type oTreeChan struct {
	dataCh chan *[]fileref.FileRef
	errCh  chan error
}

func (o *ObjectTreeRequest) GetObjectTree() (*[]fileref.FileRef, error) {
	totalBlobbersCount := len(o.blobbers)
	oTreeChans := make([]oTreeChan, totalBlobbersCount)
	o.wg.Add(totalBlobbersCount)
	for i, blob := range o.blobbers {
		Logger.Info(fmt.Sprintf("Getting page %v of file refs for path %v from blobber %v", o.page, o.remotefilepath, blob.Baseurl))
		go o.getFileRefs(&oTreeChans[i], blob.Baseurl)
	}
	o.wg.Wait()
	//TODO Check for consensus and send the result
	refsMap := make(map[string]map[string]interface{})
	for _, oTreeChan := range oTreeChans {
		refs := <-oTreeChan.dataCh
		err := <-oTreeChan.errCh

		if err != nil {
			continue
		}

		for _, ref := range *refs {
			if _, ok := refsMap[ref.LookupHash]; !ok {
				//Consensus work left
			}
		}
	}
	data := <-oTreeChans[0].dataCh
	err := <-oTreeChans[0].errCh

	return data, err
}

func (o *ObjectTreeRequest) getFileRefs(oTreechan *oTreeChan, bUrl string) {
	defer o.wg.Done()
	oReq, err := zboxutil.NewPaginatedObjectTreeRequest(bUrl, o.allocationID, o.remotefilepath, o.offsetPath, o.page)
	if err != nil {
		oTreechan.errCh <- err
		return
	}
	oResult := ObjectTreeResult{}
	ctx, cncl := context.WithTimeout(o.ctx, time.Second*30)
	err = zboxutil.HttpDo(ctx, cncl, oReq, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("ObjectTree: ", err)
			return err
		}
		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Logger.Error("ObjectTree: Error ", err, "while reading response from ", bUrl)
			return err
		}
		if resp.StatusCode == http.StatusOK {
			err := json.Unmarshal(respBody, &oResult)
			if err != nil {
				Logger.Error("ObjectTree: Error ", err, "while unmarshalling response from ", bUrl)
				return err
			}
			return nil
		} else {
			Logger.Error(bUrl, "ObjectTree Response: ", string(respBody))
		}
		return nil
	})
	if err != nil {
		oTreechan.errCh <- err
		return
	}
	oTreechan.dataCh <- &oResult.Refs
}
