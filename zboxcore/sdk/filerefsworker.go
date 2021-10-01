package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type ObjectTreeResult struct {
	// raw           []byte
	TotalPages int64               `json:"total_pages"`
	OffsetPath string              `json:"offset_path"`
	OffsetDate string              `json:"offset_date"`
	Refs       []ORef              `json:"refs"`
	LatestWM   *marker.WriteMarker `json:"latest_write_marker"`
}

type ObjectTreeRequest struct {
	allocationID       string
	allocationTx       string
	blobbers           []*blockchain.StorageNode
	remotefilepathhash string
	remotefilepath     string
	pageLimit          int
	level              int
	fileType           string
	refType            string
	offsetPath         string
	updatedDate        string //must have "2006-01-02T15:04:05.99999Z07:00" format
	offsetDate         string //must have "2006-01-02T15:04:05.99999Z07:00" format
	authToken          *marker.AuthTicket
	ctx                context.Context
	wg                 *sync.WaitGroup
	Consensus
}

type oTreeResponse struct {
	oTResult     *ObjectTreeResult
	responseHash string
	err          error
}

//Paginated tree should not be collected as this will stall the client
//It should rather be handled by application that uses gosdk
func (o *ObjectTreeRequest) GetRefs() (*ObjectTreeResult, error) {
	totalBlobbersCount := len(o.blobbers)
	oTreeResponses := make([]oTreeResponse, totalBlobbersCount)
	o.wg.Add(totalBlobbersCount)
	Logger.Info(fmt.Sprintf("Total blobbers count: %v", totalBlobbersCount))
	for i, blob := range o.blobbers {
		Logger.Info(fmt.Sprintf("Getting file refs for path %v from blobber %v", o.remotefilepath, blob.Baseurl))
		go o.getFileRefs(&oTreeResponses[i], blob.Baseurl)
	}
	o.wg.Wait()
	hashCount := make(map[string]uint8)
	hashRefsMap := make(map[string]*ObjectTreeResult)

	for _, oTreeResponse := range oTreeResponses {
		if oTreeResponse.err != nil {
			continue
		}
		var similarFieldRefs []SimilarField
		for _, ref := range oTreeResponse.oTResult.Refs {
			similarFieldRefs = append(similarFieldRefs, ref.SimilarField)
		}
		refsMarshall, err := json.Marshal(similarFieldRefs)
		if err != nil {
			continue
		}
		hash := zboxutil.GetRefsHash(refsMarshall)

		if _, ok := hashCount[hash]; ok {
			hashCount[hash]++
		} else {
			hashCount[hash]++
			hashRefsMap[hash] = oTreeResponse.oTResult
		}
	}

	var selected *ObjectTreeResult
	Logger.Info(fmt.Sprintf("Consensus threshold: %v, fullconsensus: %v", o.consensusThresh, o.fullconsensus))
	Logger.Info(fmt.Sprint("Hash count map: ", hashCount))
	for k, v := range hashCount {
		if float32(v)/o.fullconsensus >= o.consensusThresh {
			selected = hashRefsMap[k]
			break
		}
	}

	if selected != nil {
		return selected, nil
	}
	return nil, errors.New("consensus_failed", "Refs consensus is less than consensus threshold")

}

func (o *ObjectTreeRequest) getFileRefs(oTR *oTreeResponse, bUrl string) {
	defer o.wg.Done()
	oReq, err := zboxutil.NewRefsRequest(bUrl, o.allocationID, o.remotefilepath, o.offsetPath, o.updatedDate, o.offsetDate, o.fileType, o.refType, o.level, o.pageLimit)
	if err != nil {
		oTR.err = err
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
			return err
		}
	})
	if err != nil {
		oTR.err = err
		return
	}
	oTR.oTResult = &oResult
	Logger.Info("Gottcha result")
}

// type ORef struct {
// 	ID                  int64          `json:"id"`
// 	Type                string         `json:"type"`
// 	AllocationID        string         `json:"allocation_id"`
// 	LookupHash          string         `json:"lookup_hash"`
// 	Name                string         `json:"name"`
// 	Path                string         `json:"path"`
// 	Hash                string         `json:"hash"`
// 	NumBlocks           int64          `json:"num_blocks"`
// 	PathHash            string         `json:"path_hash"`
// 	ParentPath          string         `json:"parent_path"`
// 	PathLevel           int            `json:"level"`
// 	CustomMeta          string         `json:"custom_meta"`
// 	ContentHash         string         `json:"content_hash"`
// 	Size                int64          `json:"size"`
// 	MerkleRoot          string         `json:"merkle_root"`
// 	ActualFileSize      int64          `json:"actual_file_size"`
// 	ActualFileHash      string         `json:"actual_file_hash"`
// 	MimeType            string         `json:"mimetype"`
// 	WriteMarker         string         `json:"write_marker"`
// 	ThumbnailSize       int64          `json:"thumbnail_size"`
// 	ThumbnailHash       string         `json:"thumbnail_hash"`
// 	ActualThumbnailSize int64          `json:"actual_thumbnail_size"`
// 	ActualThumbnailHash string         `json:"actual_thumbnail_hash"`
// 	EncryptedKey        string         `json:"encrypted_key"`
// 	Attributes          datatypes.JSON `json:"attributes"`
// 	OnCloud             bool           `json:"on_cloud"`
// 	CreatedAt           time.Time      `json:"created_at"`
// 	UpdatedAt           time.Time      `json:"updated_at"`
// }

// Blobber response will be different from each other so we should only consider similar fields
// i.e. we cannot calculate hash of response and have consensus on it
type ORef struct {
	SimilarField
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"` //It cannot be considered for consensus calculation as blobbers can have
	UpdatedAt time.Time `json:"updated_at"` //minor difference and will fail in concensus
}

type SimilarField struct {
	Type                string `json:"type"`
	AllocationID        string `json:"allocation_id"`
	LookupHash          string `json:"lookup_hash"`
	Name                string `json:"name"`
	Path                string `json:"path"`
	PathHash            string `json:"path_hash"`
	ParentPath          string `json:"parent_path"`
	PathLevel           int    `json:"level"`
	ActualFileSize      int64  `json:"actual_file_size"`
	ActualFileHash      string `json:"actual_file_hash"`
	MimeType            string `json:"mimetype"`
	ActualThumbnailSize int64  `json:"actual_thumbnail_size"`
	ActualThumbnailHash string `json:"actual_thumbnail_hash"`
}
