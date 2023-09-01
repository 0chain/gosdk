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

	l.Logger.Info("ENTRY GetRoundFromSharders", zap.Any("sharders", sharders))
	result := make(chan *util.GetResponse, len(sharders))

	l.Logger.Info("ENTRY 2 GetRoundFromSharders", zap.Any("sharders", sharders), zap.Any("result", result))

	// getMinShardersVerify
	var numSharders = len(sharders) // overwrite, use all
	queryFromSharders(sharders, fmt.Sprintf("%v", CURRENT_ROUND), result)

	l.Logger.Info("ENTRY 3 GetRoundFromSharders", zap.Any("sharders", sharders), zap.Any("result", result))

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

			l.Logger.Info("1 GetRoundFromSharders", zap.Any("sharders", sharders), zap.Any("numSharders", numSharders), zap.Any("result", result), zap.Any("rsp", rsp))
			if rsp.StatusCode != http.StatusOK {
				l.Logger.Info("2 GetRoundFromSharders", zap.Any("sharders", sharders), zap.Any("numSharders", numSharders), zap.Any("result", result), zap.Any("rsp", rsp))
				continue
			}

			l.Logger.Info("3 GetRoundFromSharders", zap.Any("sharders", sharders), zap.Any("numSharders", numSharders), zap.Any("result", result), zap.Any("rsp", rsp))

			var respRound int64
			err := json.Unmarshal([]byte(rsp.Body), &respRound)

			l.Logger.Info("4 GetRoundFromSharders", zap.Any("sharders", sharders), zap.Any("numSharders", numSharders), zap.Any("result", result), zap.Any("rsp", rsp), zap.Any("respRound", respRound))

			if err != nil {
				l.Logger.Info("5 GetRoundFromSharders", zap.Any("sharders", sharders), zap.Any("numSharders", numSharders), zap.Any("result", result), zap.Any("rsp", rsp), zap.Any("respRound", respRound))
				continue
			}

			l.Logger.Info("6 GetRoundFromSharders", zap.Any("sharders", sharders), zap.Any("numSharders", numSharders), zap.Any("result", result), zap.Any("respRound", respRound))

			rounds = append(rounds, respRound)

			l.Logger.Info("7 GetRoundFromSharders", zap.Any("sharders", sharders), zap.Any("numSharders", numSharders), zap.Any("result", result), zap.Any("respRound", respRound))

			sort.Slice(rounds, func(i, j int) bool {
				return false
			})

			l.Logger.Info("8 GetRoundFromSharders", zap.Any("sharders", sharders), zap.Any("numSharders", numSharders), zap.Any("result", result), zap.Any("respRound", respRound))

			medianRound := rounds[len(rounds)/2]

			l.Logger.Info("9 GetRoundFromSharders", zap.Any("sharders", sharders), zap.Any("numSharders", numSharders), zap.Any("result", result), zap.Any("respRound", respRound), zap.Any("medianRound", medianRound))

			roundMap[medianRound]++

			l.Logger.Info("10 GetRoundFromSharders", zap.Any("sharders", sharders), zap.Any("numSharders", numSharders), zap.Any("result", result), zap.Any("respRound", respRound), zap.Any("medianRound", medianRound), zap.Any("roundMap", roundMap))

			if roundMap[medianRound] > consensus {

				l.Logger.Info("11 GetRoundFromSharders", zap.Any("sharders", sharders), zap.Any("numSharders", numSharders), zap.Any("result", result), zap.Any("respRound", respRound), zap.Any("medianRound", medianRound), zap.Any("roundMap", roundMap), zap.Any("consensus", consensus))

				consensus = roundMap[medianRound]

				l.Logger.Info("12 GetRoundFromSharders", zap.Any("sharders", sharders), zap.Any("numSharders", numSharders), zap.Any("result", result), zap.Any("respRound", respRound), zap.Any("medianRound", medianRound), zap.Any("roundMap", roundMap), zap.Any("consensus", consensus))

				round = medianRound

				l.Logger.Info("13 GetRoundFromSharders", zap.Any("sharders", sharders), zap.Any("numSharders", numSharders), zap.Any("result", result), zap.Any("respRound", respRound), zap.Any("medianRound", medianRound), zap.Any("roundMap", roundMap), zap.Any("consensus", consensus), zap.Any("round", round))

				rate := consensus * 100 / int64(len(sharders))

				l.Logger.Info("14 GetRoundFromSharders", zap.Any("sharders", sharders), zap.Any("numSharders", numSharders), zap.Any("result", result), zap.Any("respRound", respRound), zap.Any("medianRound", medianRound), zap.Any("roundMap", roundMap), zap.Any("consensus", consensus), zap.Any("round", round), zap.Any("rate", rate))

				if rate >= int64(consensusThresh) {

					l.Logger.Info("15 GetRoundFromSharders", zap.Any("sharders", sharders), zap.Any("numSharders", numSharders), zap.Any("result", result), zap.Any("respRound", respRound), zap.Any("medianRound", medianRound), zap.Any("roundMap", roundMap), zap.Any("consensus", consensus), zap.Any("round", round), zap.Any("rate", rate))

					return round, nil
				}

				l.Logger.Info("16 GetRoundFromSharders", zap.Any("sharders", sharders), zap.Any("numSharders", numSharders), zap.Any("result", result), zap.Any("respRound", respRound), zap.Any("medianRound", medianRound), zap.Any("roundMap", roundMap), zap.Any("consensus", consensus), zap.Any("round", round), zap.Any("rate", rate))
			}
		}
	}

	l.Logger.Info("17 GetRoundFromSharders", zap.Any("sharders", sharders), zap.Any("numSharders", numSharders), zap.Any("round", round))

	return round, nil
}

func queryFromSharders(sharders []string, query string,
	result chan *util.GetResponse) {

	l.Logger.Info("queryFromSharders", zap.Any("sharders", sharders), zap.Any("query", query), zap.Any("result", result))

	queryFromShardersContext(context.Background(), sharders, query, result)
}

func queryFromShardersContext(ctx context.Context, sharders []string,
	query string, result chan *util.GetResponse) {

	l.Logger.Info("queryFromShardersContext", zap.Any("sharders", sharders), zap.Any("query", query), zap.Any("result", result))

	for _, sharder := range util.Shuffle(sharders) {
		go func(sharderurl string) {
			l.Logger.Info("2 queryFromShardersContext", zap.Any("sharderurl", sharderurl), zap.Any("query", query), zap.Any("result", result))

			url := fmt.Sprintf("%v%v", sharderurl, query)
			req, err := util.NewHTTPGetRequestContext(ctx, url)

			l.Logger.Info("3 queryFromShardersContext", zap.Any("sharderurl", sharderurl), zap.Any("query", query), zap.Any("result", result), zap.Any("req", req), zap.Any("err", err))

			if err != nil {
				return
			}

			l.Logger.Info("4 queryFromShardersContext", zap.Any("sharderurl", sharderurl), zap.Any("query", query), zap.Any("result", result), zap.Any("req", req), zap.Any("err", err))

			res, err := req.Get()

			l.Logger.Info("5 queryFromShardersContext", zap.Any("sharderurl", sharderurl), zap.Any("query", query), zap.Any("result", result), zap.Any("req", req), zap.Any("err", err), zap.Any("res", res))

			if err != nil {
				return
			}

			l.Logger.Info("6 queryFromShardersContext", zap.Any("sharderurl", sharderurl), zap.Any("query", query), zap.Any("result", result), zap.Any("req", req), zap.Any("err", err), zap.Any("res", res))

			result <- res

			l.Logger.Info("7 queryFromShardersContext", zap.Any("sharderurl", sharderurl), zap.Any("query", query), zap.Any("result", result), zap.Any("req", req), zap.Any("err", err), zap.Any("res", res))
		}(sharder)
	}
}
