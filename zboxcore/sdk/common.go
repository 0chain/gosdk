package sdk

import (
	"context"
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"github.com/0chain/gosdk/core/util"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"path"
	"sort"
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

const (
	CURRENT_ROUND = "/v1/current-round"
)

func getObjectTreeFromBlobber(ctx context.Context, allocationID, allocationTx string, remoteFilePath string, blobber *blockchain.StorageNode) (fileref.RefEntity, error) {
	httpreq, err := zboxutil.NewObjectTreeRequest(blobber.Baseurl, allocationID, allocationTx, remoteFilePath)
	if err != nil {
		l.Logger.Error(blobber.Baseurl, "Error creating object tree request", err)
		return nil, err
	}
	var lR ReferencePathResult
	ctx, cncl := context.WithTimeout(ctx, (time.Second * 30))
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
			l.Logger.Info("Object tree:", string(resp_body))
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

var ErrFileNameTooLong = errors.New("invalid_parameter", "filename is longer than 100 characters")

func ValidateRemoteFileName(remotePath string) error {
	_, fileName := path.Split(remotePath)

	if len(fileName) > 100 {
		return ErrFileNameTooLong
	}

	return nil
}

func GetRoundFromSharders() (int64, error) {
	sharders := blockchain.GetSharders()

	result := make(chan *util.GetResponse, len(sharders))
	// getMinShardersVerify
	var numSharders = len(sharders) // overwrite, use all
	queryFromSharders(sharders, fmt.Sprintf("%v", CURRENT_ROUND), result)

	fmt.Println("GetRoundFromSharders", zap.Any("sharders", sharders), zap.Any("numSharders", numSharders), zap.Any("result", result))
	const consensusThresh = float32(25.0)

	var rounds []int64

	consensus := int64(0)
	roundMap := make(map[int64]int64)

	round := int64(0)

	waitTimeC := time.After(10 * time.Second)
	for i := 0; i < numSharders; i++ {
		select {
		case <-waitTimeC:
			return 0, stdErrors.New("get round failed. consensus not reached")
		case rsp := <-result:
			if rsp.StatusCode != http.StatusOK {
				continue
			}

			var respRound int64
			err := json.Unmarshal([]byte(rsp.Body), &respRound)
			if err != nil {
				continue
			}

			fmt.Println("GetRoundFromSharders", zap.Any("sharders", sharders), zap.Any("numSharders", numSharders), zap.Any("result", result), zap.Any("respRound", respRound))

			rounds = append(rounds, respRound)
			sort.Slice(rounds, func(i, j int) bool {
				return false
			})

			medianRound := rounds[len(rounds)/2]
			roundMap[medianRound]++

			if roundMap[medianRound] > consensus {
				consensus = roundMap[medianRound]
				round = medianRound

				rate := consensus * 100 / int64(len(sharders))
				if rate >= int64(consensusThresh) {
					return round, nil
				}
			}
		}
	}

	return 0, stdErrors.New("get round failed, consensus not reached")
}

func queryFromSharders(sharders []string, query string,
	result chan *util.GetResponse) {

	fmt.Println("queryFromSharders", zap.Any("sharders", sharders), zap.Any("query", query), zap.Any("result", result))

	queryFromShardersContext(context.Background(), sharders, query, result)
}

func queryFromShardersContext(ctx context.Context, sharders []string,
	query string, result chan *util.GetResponse) {

	fmt.Println("queryFromShardersContext", zap.Any("sharders", sharders), zap.Any("query", query), zap.Any("result", result))

	for _, sharder := range util.Shuffle(sharders) {
		go func(sharderurl string) {

			fmt.Println("2 queryFromShardersContext", zap.Any("sharderurl", sharderurl), zap.Any("query", query), zap.Any("result", result))

			//Logger.Info("Query from ", sharderurl+query)
			url := fmt.Sprintf("%v%v", sharderurl, query)
			req, err := util.NewHTTPGetRequestContext(ctx, url)
			if err != nil {
				fmt.Println("new get request failed. ", zap.Any("url", url), zap.Error(err))
				return
			}
			res, err := req.Get()
			if err != nil {
				fmt.Println("2 new get request failed. ", zap.Any("url", url), zap.Error(err))
				return
			}

			l.Logger.Error("round info parse error.", res)

			fmt.Println("3 new get request failed. ", zap.Any("url", url), zap.Error(err), zap.Any("res", res.Body))
			result <- res
		}(sharder)
	}
}
