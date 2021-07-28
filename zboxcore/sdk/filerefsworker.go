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
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"gorm.io/datatypes"
)

type ObjectTreeResult struct {
	Page          int64               `json:"page"`
	TotalPages    int64               `json:"total_pages"`
	NewOffsetPath string              `json:"offsetPath"`
	Refs          []ORef              `json:"refs"`
	LatestWM      *marker.WriteMarker `json:"latest_write_marker"`
}

type ObjectTreeRequest struct {
	allocationID       string
	allocationTx       string
	blobbers           []*blockchain.StorageNode
	remotefilepathhash string
	remotefilepath     string
	pageLimit          int
	level              int
	_type              string
	offsetPath         string
	updatedDate        string
	offsetDate         string
	authToken          *marker.AuthTicket
	ctx                context.Context
	wg                 *sync.WaitGroup
	Consensus
}

type oTreeChan struct {
	dataCh chan *ObjectTreeResult
	errCh  chan error
}

//Paginated tree should not be collected as this will stall the client
//It should rather be handled by application that uses gosdk
func (o *ObjectTreeRequest) GetRefs() (*ObjectTreeResult, error) {
	totalBlobbersCount := len(o.blobbers)
	oTreeChans := make([]oTreeChan, totalBlobbersCount)
	o.wg.Add(totalBlobbersCount)
	for i, blob := range o.blobbers {
		Logger.Info(fmt.Sprintf("Getting file refs for path %v from blobber %v", o.remotefilepath, blob.Baseurl))
		go o.getFileRefs(&oTreeChans[i], blob.Baseurl)
	}
	o.wg.Wait()
	//TODO Check for consensus and send the result
	refsMap := make(map[string]map[string]interface{})
	for _, oTreeChan := range oTreeChans {
		oTreeResult := <-oTreeChan.dataCh
		err := <-oTreeChan.errCh

		if err != nil {
			continue
		}

		for _, ref := range oTreeResult.Refs {
			if _, ok := refsMap[ref.LookupHash]; !ok {
				//Consensus work left
			}
		}
	}
	oTreeResult := <-oTreeChans[0].dataCh
	err := <-oTreeChans[0].errCh

	return oTreeResult, err
}

func (o *ObjectTreeRequest) getFileRefs(oTreechan *oTreeChan, bUrl string) {
	defer o.wg.Done()
	oReq, err := zboxutil.NewRefsRequest(bUrl, o.allocationID, o.remotefilepath, o.offsetPath, o.updatedDate, o.offsetDate, o._type, o.level, o.pageLimit)
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
	oTreechan.dataCh <- &oResult
}

type ORef struct {
	ID                  int64          `json:"id"`
	Type                string         `json:"type"`
	AllocationID        string         `json:"allocation_id"`
	LookupHash          string         `json:"lookup_hash"`
	Name                string         `json:"name"`
	Path                string         `json:"path"`
	Hash                string         `json:"hash"`
	NumBlocks           int64          `json:"num_blocks"`
	PathHash            string         `json:"path_hash"`
	ParentPath          string         `json:"parent_path"`
	PathLevel           int            `json:"level"`
	CustomMeta          string         `json:"custom_meta"`
	ContentHash         string         `json:"content_hash"`
	Size                int64          `json:"size"`
	MerkleRoot          string         `json:"merkle_root"`
	ActualFileSize      int64          `json:"actual_file_size"`
	ActualFileHash      string         `json:"actual_file_hash"`
	MimeType            string         `json:"mimetype"`
	WriteMarker         string         `json:"write_marker"`
	ThumbnailSize       int64          `json:"thumbnail_size"`
	ThumbnailHash       string         `json:"thumbnail_hash"`
	ActualThumbnailSize int64          `json:"actual_thumbnail_size"`
	ActualThumbnailHash string         `json:"actual_thumbnail_hash"`
	EncryptedKey        string         `json:"encrypted_key"`
	Attributes          datatypes.JSON `json:"attributes"`
	OnCloud             bool           `json:"on_cloud"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}
