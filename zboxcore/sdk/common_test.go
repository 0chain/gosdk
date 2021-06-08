package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

)

const (
	testData   = "testdata"
	walletString = `{"client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","client_key":"eeb0c33325cbee0fb58bc09962f69a44d0b22ac2824a063eb1002273347e601a4612e6fea7e1a1ae62e0e3b7f1301c4de8a855bae86ebfa6e9dbbb41c3e39c24","keys":[{"public_key":"eeb0c33325cbee0fb58bc09962f69a44d0b22ac2824a063eb1002273347e601a4612e6fea7e1a1ae62e0e3b7f1301c4de8a855bae86ebfa6e9dbbb41c3e39c24","private_key":"0800424da684ff94ac8af3ccc3e024a8d16bb9054237e79feffc486efda6e210"}],"mnemonics":"neck hurt glow action goose gadget meat ankle patch boy truth convince glass grief cloth sunny evil puppy decorate language okay burst replace cigar","version":"1.0","date_created":"2021-03-13 01:42:49.625529496 +0700 +07 m=+1.336039112"}`
	allocationString = `{"id":"69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f","tx":"69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f","data_shards":2,"parity_shards":2,"size":2147483648,"expiration_date":1617542537,"owner_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","owner_public_key":"eeb0c33325cbee0fb58bc09962f69a44d0b22ac2824a063eb1002273347e601a4612e6fea7e1a1ae62e0e3b7f1301c4de8a855bae86ebfa6e9dbbb41c3e39c24","payer_id":"","blobbers":[{"id":"${blobber_id_1}","url":"${blobber_url_1}"},{"id":"${blobber_id_2}","url":"${blobber_url_2}"},{"id":"${blobber_id_3}","url":"${blobber_url_3}"},{"id":"${blobber_id_4}","url":"${blobber_url_4}"}],"stats":{"used_size":0,"num_of_writes":0,"num_of_reads":0,"total_challenges":0,"num_open_challenges":0,"num_success_challenges":0,"num_failed_challenges":0,"latest_closed_challenge":""},"time_unit":172800000000000,"blobber_details":[{"blobber_id":"${blobber_id_1}","size":357913942,"terms":{"read_price":344362696,"write_price":172181348,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"min_lock_demand":86090674,"spent":0,"penalty":0,"read_reward":0,"returned":0,"challenge_reward":0,"final_reward":0},{"blobber_id":"${blobber_id_2}","size":357913942,"terms":{"read_price":344362696,"write_price":172181348,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"min_lock_demand":86090674,"spent":0,"penalty":0,"read_reward":0,"returned":0,"challenge_reward":0,"final_reward":0},{"blobber_id":"${blobber_id_3}","size":357913942,"terms":{"read_price":100000000,"write_price":100000000,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"min_lock_demand":50000000,"spent":0,"penalty":0,"read_reward":0,"returned":0,"challenge_reward":0,"final_reward":0},{"blobber_id":"${blobber_id_4}","size":357913942,"terms":{"read_price":312180015,"write_price":156090007,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"min_lock_demand":78045003,"spent":0,"penalty":0,"read_reward":0,"returned":0,"challenge_reward":0,"final_reward":0}],"read_price_range":{"min":0,"max":10000000000},"write_price_range":{"min":0,"max":10000000000},"challenge_completion_time":120000000000,"start_time":1614950537}`
	textPlainContentType = "text/plain"
)

func blobberIDMask(idx int) string {
	return fmt.Sprintf("${blobber_id_%v}", idx)
}

func blobberURLMask(idx int) string {
	return fmt.Sprintf("${blobber_url_%v}", idx)
}

type nodeConfig struct {
	BlockWorker       string   `yaml:"block_worker"`
	PreferredBlobbers []string `yaml:"preferred_blobbers"`
	SignScheme        string   `yaml:"signature_scheme"`
	ChainID           string   `yaml:"chain_id"`
}

var nodeCfg = nodeConfig{
	BlockWorker: "https://one.devnet-0chain.net/dns",
	SignScheme: "bls0chain",
}

func setupMockInitStorageSDK() error {
	clientConfig := walletString
	blockWorker := nodeCfg.BlockWorker
	preferredBlobbers := nodeCfg.PreferredBlobbers
	signScheme := nodeCfg.SignScheme
	chainID := nodeCfg.ChainID

	return InitStorageSDK(clientConfig, blockWorker, chainID, signScheme, preferredBlobbers)
}

type writeFile struct {
	FileName string
	IsDir    bool
	Content  []byte
	Child    []*writeFile
}

var (
	downloadSuccessFileChan chan []*writeFile
	commitResultChan        chan *CommitResult
)

func willDownloadSuccessFiles(wf ...*writeFile) {
	downloadSuccessFileChan <- wf
}

func deleteFiles(wfs ...*writeFile) {
	for _, wf := range wfs {
		os.Remove(wf.FileName)
	}
}

func willReturnCommitResult(c *CommitResult) {
	commitResultChan <- c
}

func writeFiles(wfs ...*writeFile) {
	for _, wf := range wfs {
		var err error
		if wf.IsDir {
			err = os.MkdirAll(wf.FileName, 0755)
			if len(wf.Child) > 0 {
				writeFiles(wf.Child...)
			}
		} else {
			err = ioutil.WriteFile(wf.FileName, wf.Content, 0644)
		}
		if err != nil {
			break
		}
	}
}

func setupMockAllocation() (allocation *Allocation, cncl func(), err error) {
	allocationBytes := []byte(allocationString)
	err = json.Unmarshal(allocationBytes, &allocation)
	if err != nil {
		return nil, nil, err
	}
	allocation.uploadChan = make(chan *UploadRequest, 10)
	allocation.downloadChan = make(chan *DownloadRequest, 10)
	allocation.repairChan = make(chan *RepairRequest, 1)
	allocation.ctx, allocation.ctxCancelF = context.WithCancel(context.Background())
	allocation.uploadProgressMap = make(map[string]*UploadRequest)
	allocation.downloadProgressMap = make(map[string]*DownloadRequest)
	allocation.mutex = &sync.Mutex{}

	// init mock test commit worker
	commitChan = make(map[string]chan *CommitRequest)
	commitResultChan = make(chan *CommitResult)

	var commitResult *CommitResult
	for _, blobber := range allocation.Blobbers {
		if _, ok := commitChan[blobber.ID]; !ok {
			commitChan[blobber.ID] = make(chan *CommitRequest, 1)
			blobberChan := commitChan[blobber.ID]
			go func(c <-chan *CommitRequest, blID string) {
				for true {
					cm := <-c
					if cm != nil {
						cm.result = commitResult
						if cm.wg != nil {
							cm.wg.Done()
						}
					}
				}
			}(blobberChan, blobber.ID)
		}
	}

	downloadSuccessFileChan = make(chan []*writeFile, 5)
	var downloadWriteFiles []*writeFile

	// init mock test dispatcher, commit result, download success
	go func() {
		for true {
			select {
			case <-allocation.ctx.Done():
				return
			case commitResult = <-commitResultChan:
			case wfs := <-downloadSuccessFileChan:
				downloadWriteFiles = wfs
			case uploadReq := <-allocation.uploadChan:
				if uploadReq.completedCallback != nil {
					uploadReq.completedCallback(uploadReq.filepath)
				}
				if uploadReq.statusCallback != nil {
					uploadReq.statusCallback.Completed(allocation.ID, uploadReq.filepath, uploadReq.filemeta.Name, uploadReq.filemeta.MimeType, int(uploadReq.filemeta.Size), OpUpload)
				}
				if uploadReq.wg != nil {
					uploadReq.wg.Done()
				}
			case downloadReq := <-allocation.downloadChan:
				if len(downloadWriteFiles) > 0 {
					writeFiles(downloadWriteFiles...)
				}
				if downloadReq.completedCallback != nil {
					downloadReq.completedCallback(downloadReq.remotefilepath, downloadReq.remotefilepathhash)
				}
				if downloadReq.statusCallback != nil {
					downloadReq.statusCallback.Completed(allocation.ID, downloadReq.localpath, "1.txt", "application/octet-stream", 3, OpDownload)
				}
				if downloadReq.wg != nil {
					downloadReq.wg.Done()
				}
			case repairReq := <-allocation.repairChan:
				if repairReq.completedCallback != nil {
					repairReq.completedCallback()
				}
				if repairReq.wg != nil {
					repairReq.wg.Done()
				}
			}
		}
	}()
	allocation.initialized = true

	return allocation, func() {
		for _, commitResultChan := range commitChan {
			close(commitResultChan)
		}
		close(downloadSuccessFileChan)
	}, err
}
