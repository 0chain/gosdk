package sdk

import (
	"encoding/json"
	blobbergrpc "github.com/0chain/blobber/code/go/0chain.net/blobbercore/blobbergrpc/proto"
	"strings"
	"sync"
	"time"

	"github.com/0chain/gosdk/core/clients/blobberClient"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"
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

	respRaw, err := blobberClient.GetFileStats(blobber.Baseurl, &blobbergrpc.GetFileStatsRequest{
		PathHash:   req.remotefilepathhash,
		Allocation: req.allocationTx,
	})
	if err != nil {
		logger.Logger.Error("could not get file stats from blobber -" + blobber.Baseurl + " - " + err.Error())
		return
	}
	s.WriteString(string(respRaw))
	err = json.Unmarshal(respRaw, &fileStats)
	if err != nil {
		return
	}

	if len(fileStats.WriteMarkerRedeemTxn) > 0 {
		fileStats.BlockchainAware = true
	} else {
		fileStats.BlockchainAware = false
	}
	fileStats.BlobberID = blobber.ID
	fileStats.BlobberURL = blobber.Baseurl
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
	fileInfos := make(map[string]*FileStats)
	for i := 0; i < numList; i++ {
		ch := <-rspCh
		fileInfos[req.blobbers[i].ID] = ch.filestats
	}
	return fileInfos
}
