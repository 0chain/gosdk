package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type FileStats struct {
	Name                     string    `json:"name"`
	Size                     int64     `json:"size"`
	PathHash                 string    `json:"path_hash"`
	Path                     string    `json:"path"`
	NumBlocks                int64     `json:"num_of_blocks"`
	NumUpdates               int64     `json:"num_of_updates"`
	NumBlockDownloads        int64     `json:"num_of_block_downloads"`
	SuccessChallenges        int64     `json:"num_of_challenges"`
	FailedChallenges         int64     `json:"num_of_failed_challenges"`
	LastChallengeResponseTxn string    `json:"last_challenge_txn"`
	WriteMarkerRedeemTxn     string    `json:"write_marker_txn"`
	BlobberID                string    `json:"blobber_id"`
	BlobberURL               string    `json:"blobber_url"`
	BlockchainAware          bool      `json:"blockchain_aware"`
	CreatedAt                time.Time `json:"CreatedAt"`
}

type fileStatsResponse struct {
	filestats   *FileStats
	responseStr string
	blobberIdx  int
	err         error
}

func (req *ListRequest) getFileStatsInfoFromBlobber(blobber *blockchain.StorageNode, blobberIdx int, rspCh chan<- *fileStatsResponse) {
	defer req.wg.Done()
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)

	var fileStats *FileStats
	var s strings.Builder
	var err error
	fileMetaRetFn := func() {
		rspCh <- &fileStatsResponse{filestats: fileStats, responseStr: s.String(), blobberIdx: blobberIdx, err: err}
	}
	defer fileMetaRetFn()
	if len(req.remotefilepath) > 0 {
		req.remotefilepathhash = fileref.GetReferenceLookup(req.allocationID, req.remotefilepath)
	}
	formWriter.WriteField("path_hash", req.remotefilepathhash)

	formWriter.Close()
	httpreq, err := zboxutil.NewFileStatsRequest(blobber.Baseurl, req.allocationTx, body)
	if err != nil {
		Logger.Error("File meta info request error: ", err.Error())
		return
	}

	httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
	ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 30))
	err = zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("GetFileStats : ", err)
			return err
		}
		defer resp.Body.Close()
		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Error: Resp : %s", err.Error())
		}
		s.WriteString(string(resp_body))
		if resp.StatusCode == http.StatusOK {
			err = json.Unmarshal(resp_body, &fileStats)
			if err != nil {
				return fmt.Errorf("file stats response parse error: %s", err.Error())
			}
			if len(fileStats.WriteMarkerRedeemTxn) > 0 {
				fileStats.BlockchainAware = true
			} else {
				fileStats.BlockchainAware = false
			}
			fileStats.BlobberID = blobber.ID
			fileStats.BlobberURL = blobber.Baseurl
			return nil
		}
		return err
	})
}

func (req *ListRequest) getFileStatsFromBlobbers() map[string]*FileStats {
	numList := len(req.blobbers)
	//fmt.Printf("%v\n", req.blobbers)
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numList)
	rspCh := make(chan *fileStatsResponse, numList)
	for i := 0; i < numList; i++ {
		go req.getFileStatsInfoFromBlobber(req.blobbers[i], i, rspCh)
	}
	req.wg.Wait()
	var fileInfos map[string]*FileStats
	for i := 0; i < numList; i++ {
		ch := <-rspCh
		if ch.filestats == nil {
			continue
		}
		if fileInfos == nil {
			fileInfos = make(map[string]*FileStats)
		}
		fileInfos[req.blobbers[i].ID] = ch.filestats
	}
	return fileInfos
}
