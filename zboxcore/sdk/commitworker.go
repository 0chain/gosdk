package sdk

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/0chain/common/core/util/wmpt"
	"github.com/0chain/errors"
	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/minio/sha256-simd"
)

type ReferencePathResult struct {
	*fileref.ReferencePath
	LatestWM *marker.WriteMarker `json:"latest_write_marker"`
	Version  string              `json:"version"`
}

type ReferencePathResultV2 struct {
	Path     []byte              `json:"path"`
	LatestWM *marker.WriteMarker `json:"latest_write_marker"`
	Version  string              `json:"version"`
}

type CommitResult struct {
	Success      bool   `json:"success"`
	ErrorMessage string `json:"error_msg,omitempty"`
}

func ErrorCommitResult(errMsg string) *CommitResult {
	result := &CommitResult{Success: false, ErrorMessage: errMsg}
	return result
}

func SuccessCommitResult() *CommitResult {
	result := &CommitResult{Success: true}
	return result
}

const MARKER_VERSION = "v2"

type CommitRequest struct {
	changes      []allocationchange.AllocationChange
	blobber      *blockchain.StorageNode
	allocationID string
	allocationTx string
	connectionID string
	sig          string
	wg           *sync.WaitGroup
	result       *CommitResult
	timestamp    int64
	blobberInd   uint64
}

type CommitRequestInterface interface {
	processCommit()
	blobberID() string
}

type CommitRequestV2 struct {
	changes         []allocationchange.AllocationChangeV2
	allocationObj   *Allocation
	connectionID    string
	sig             string
	wg              *sync.WaitGroup
	result          *CommitResult
	timestamp       int64
	consensusThresh int
	commitMask      zboxutil.Uint128
	changeIndex     uint64
	isRepair        bool
}

var (
	commitChan           map[string]chan CommitRequestInterface
	initCommitMutex      sync.Mutex
	errAlreadySuccessful = errors.New("alread_successful", "")
)

func InitCommitWorker(blobbers []*blockchain.StorageNode) {
	initCommitMutex.Lock()
	defer initCommitMutex.Unlock()
	if commitChan == nil {
		commitChan = make(map[string]chan CommitRequestInterface)
	}

	for _, blobber := range blobbers {
		if _, ok := commitChan[blobber.ID]; !ok {
			commitChan[blobber.ID] = make(chan CommitRequestInterface, 1)
			blobberChan := commitChan[blobber.ID]
			go startCommitWorker(blobberChan, blobber.ID)
		}
	}

}

func startCommitWorker(blobberChan chan CommitRequestInterface, blobberID string) {
	for {
		commitreq, open := <-blobberChan
		if !open {
			break
		}
		commitreq.processCommit()
	}
	initCommitMutex.Lock()
	defer initCommitMutex.Unlock()
	delete(commitChan, blobberID)
}

func (commitreq *CommitRequest) blobberID() string {
	return commitreq.blobber.ID
}

func (commitreq *CommitRequest) processCommit() {
	defer commitreq.wg.Done()
	start := time.Now()
	l.Logger.Debug("received a commit request")
	paths := make([]string, 0)
	for _, change := range commitreq.changes {
		paths = append(paths, change.GetAffectedPath()...)
	}
	if len(paths) == 0 {
		l.Logger.Debug("Nothing to commit")
		commitreq.result = SuccessCommitResult()
		return
	}
	var req *http.Request
	var lR ReferencePathResult
	req, err := zboxutil.NewReferencePathRequest(commitreq.blobber.Baseurl, commitreq.allocationID, commitreq.allocationTx, commitreq.sig, paths)
	if err != nil {
		l.Logger.Error("Creating ref path req", err)
		return
	}
	ctx, cncl := context.WithTimeout(context.Background(), (time.Second * 30))
	err = zboxutil.HttpDo(ctx, cncl, req, func(resp *http.Response, err error) error {
		if err != nil {
			l.Logger.Error("Ref path error:", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			l.Logger.Error("Ref path response : ", resp.StatusCode)
		}
		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			l.Logger.Error("Ref path: Resp", err)
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return errors.New(
				strconv.Itoa(resp.StatusCode),
				fmt.Sprintf("Reference path error response: Status: %d - %s ",
					resp.StatusCode, string(resp_body)))
		}
		err = json.Unmarshal(resp_body, &lR)
		if err != nil {
			l.Logger.Error("Reference path json decode error: ", err)
			return err
		}
		return nil
	})

	if err != nil {
		commitreq.result = ErrorCommitResult(err.Error())
		return
	}
	rootRef, err := lR.GetDirTree(commitreq.allocationID)

	if err != nil {
		commitreq.result = ErrorCommitResult(err.Error())
		return
	}
	hasher := sha256.New()
	if lR.LatestWM != nil {
		err = lR.LatestWM.VerifySignature(client.GetClientPublicKey())
		if err != nil {
			e := errors.New("signature_verification_failed", err.Error())
			commitreq.result = ErrorCommitResult(e.Error())
			return
		}
		if commitreq.timestamp <= lR.LatestWM.Timestamp {
			commitreq.timestamp = lR.LatestWM.Timestamp + 1
		}

		rootRef.CalculateHash()
		prevAllocationRoot := rootRef.Hash
		if prevAllocationRoot != lR.LatestWM.AllocationRoot {
			l.Logger.Error("Allocation root from latest writemarker mismatch. Expected: " + prevAllocationRoot + " got: " + lR.LatestWM.AllocationRoot)
			errMsg := fmt.Sprintf(
				"calculated allocation root mismatch from blobber %s. Expected: %s, Got: %s",
				commitreq.blobber.Baseurl, prevAllocationRoot, lR.LatestWM.AllocationRoot)
			commitreq.result = ErrorCommitResult(errMsg)
			return
		}
		if lR.LatestWM.ChainHash != "" {
			prevChainHash, err := hex.DecodeString(lR.LatestWM.ChainHash)
			if err != nil {
				commitreq.result = ErrorCommitResult(err.Error())
				return
			}
			hasher.Write(prevChainHash) //nolint:errcheck
		}
	}

	var size int64
	fileIDMeta := make(map[string]string)

	for _, change := range commitreq.changes {
		err = change.ProcessChange(rootRef, fileIDMeta)
		if err != nil {
			if !errors.Is(err, allocationchange.ErrRefNotFound) {
				commitreq.result = ErrorCommitResult(err.Error())
				return
			}
		} else {
			size += change.GetSize()
		}
	}
	rootRef.CalculateHash()
	var chainHash string
	if lR.Version == MARKER_VERSION {
		decodedHash, _ := hex.DecodeString(rootRef.Hash)
		hasher.Write(decodedHash) //nolint:errcheck
		chainHash = hex.EncodeToString(hasher.Sum(nil))
	}
	err = commitreq.commitBlobber(rootRef, chainHash, lR.LatestWM, size, fileIDMeta)
	if err != nil {
		commitreq.result = ErrorCommitResult(err.Error())
		return
	}
	l.Logger.Debug("[commitBlobber]", time.Since(start).Milliseconds())
	commitreq.result = SuccessCommitResult()
}

func (req *CommitRequest) commitBlobber(
	rootRef *fileref.Ref, chainHash string, latestWM *marker.WriteMarker, size int64,
	fileIDMeta map[string]string) (err error) {

	fileIDMetaData, err := json.Marshal(fileIDMeta)
	if err != nil {
		l.Logger.Error("Marshalling inode metadata failed: ", err)
		return err
	}

	wm := &marker.WriteMarker{}
	wm.AllocationRoot = rootRef.Hash
	wm.ChainSize = size
	if latestWM != nil {
		wm.PreviousAllocationRoot = latestWM.AllocationRoot
		wm.ChainSize += latestWM.ChainSize
	} else {
		wm.PreviousAllocationRoot = ""
	}
	if wm.AllocationRoot == wm.PreviousAllocationRoot {
		l.Logger.Debug("Allocation root and previous allocation root are same")
		return nil
	}
	wm.ChainHash = chainHash
	wm.FileMetaRoot = rootRef.FileMetaHash
	wm.AllocationID = req.allocationID
	wm.Size = size
	wm.BlobberID = req.blobber.ID
	wm.Timestamp = req.timestamp
	wm.ClientID = client.GetClientID()
	err = wm.Sign()
	if err != nil {
		l.Logger.Error("Signing writemarker failed: ", err)
		return err
	}
	wmData, err := json.Marshal(wm)
	if err != nil {
		l.Logger.Error("Creating writemarker failed: ", err)
		return err
	}

	l.Logger.Debug("Committing to blobber." + req.blobber.Baseurl)
	var (
		resp           *http.Response
		shouldContinue bool
	)
	for retries := 0; retries < 6; retries++ {
		err, shouldContinue = func() (err error, shouldContinue bool) {
			body := new(bytes.Buffer)
			formWriter, err := getFormWritter(req.connectionID, wmData, fileIDMetaData, body)
			if err != nil {
				l.Logger.Error("Creating form writer failed: ", err)
				return
			}
			httpreq, err := zboxutil.NewCommitRequest(req.blobber.Baseurl, req.allocationID, req.allocationTx, body, 0)
			if err != nil {
				l.Logger.Error("Error creating commit req: ", err)
				return
			}
			httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
			reqCtx, ctxCncl := context.WithTimeout(context.Background(), time.Second*60)
			resp, err = zboxutil.Client.Do(httpreq.WithContext(reqCtx))
			defer ctxCncl()

			if err != nil {
				logger.Logger.Error("Commit: ", err)
				return
			}

			if resp.Body != nil {
				defer resp.Body.Close()
			}

			var respBody []byte
			respBody, err = io.ReadAll(resp.Body)
			if err != nil {
				logger.Logger.Error("Response read: ", err)
				return
			}
			if resp.StatusCode == http.StatusOK {
				logger.Logger.Debug(req.blobber.Baseurl, " committed")
				return
			}

			if resp.StatusCode == http.StatusTooManyRequests {
				logger.Logger.Debug(req.blobber.Baseurl,
					" got too many request error. Retrying")

				var r int
				r, err = zboxutil.GetRateLimitValue(resp)
				if err != nil {
					logger.Logger.Error(err)
					return
				}

				time.Sleep(time.Duration(r) * time.Second)
				shouldContinue = true
				return
			}

			if strings.Contains(string(respBody), "pending_markers:") {
				logger.Logger.Debug("Commit pending for blobber ",
					req.blobber.Baseurl, " Retrying")
				time.Sleep(5 * time.Second)
				shouldContinue = true
				return
			}

			if strings.Contains(string(respBody), "chain_length_exceeded") {
				l.Logger.Error("Chain length exceeded for blobber ",
					req.blobber.Baseurl, " Retrying")
				time.Sleep(5 * time.Second)
				shouldContinue = true
				return
			}

			err = thrown.New("commit_error",
				fmt.Sprintf("Got error response %s with status %d", respBody, resp.StatusCode))
			return
		}()
		if shouldContinue {
			continue
		}
		return
	}
	return thrown.New("commit_error", fmt.Sprintf("Commit failed with response status %d", resp.StatusCode))
}

func AddCommitRequest(req CommitRequestInterface) {
	commitChan[req.blobberID()] <- req
}

func (commitReq *CommitRequestV2) blobberID() string {
	var pos uint64
	for i := commitReq.commitMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		commitReq.changeIndex = pos
		return commitReq.allocationObj.Blobbers[pos].ID
	}
	// we should never reach here
	return ""
}

type refPathResp struct {
	trie *wmpt.WeightedMerkleTrie
	pos  uint64
	err  error
}

func (commitReq *CommitRequestV2) processCommit() {
	defer commitReq.wg.Done()
	l.Logger.Debug("received a commit request")
	paths := make([]string, 0)
	changeIndex := commitReq.changeIndex
	for i := 0; i < len(commitReq.changes); i++ {
		lookupHash := commitReq.changes[i].GetLookupHash(changeIndex)
		if lookupHash != "" {
			paths = append(paths, lookupHash)
		} else {
			commitReq.changes[i] = nil
		}
	}

	var (
		pos     uint64
		mu      = &sync.Mutex{}
		success bool
	)
	now := time.Now()
	respChan := make(chan refPathResp, commitReq.commitMask.CountOnes())
	for i := commitReq.commitMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		go func(ind uint64) {
			blobber := commitReq.allocationObj.Blobbers[ind]
			trie, err := getReferencePathV2(blobber, commitReq.allocationObj.ID, commitReq.allocationObj.Tx, commitReq.sig, paths, &success, mu)
			resp := refPathResp{
				trie: trie,
				err:  err,
				pos:  ind,
			}
			respChan <- resp
		}(pos)
	}

	var (
		trie *wmpt.WeightedMerkleTrie
		err  error
	)

	for {
		resp := <-respChan
		if resp.err == nil {
			trie = resp.trie
			latestWM := commitReq.allocationObj.Blobbers[resp.pos].LatestWM
			if latestWM != nil && commitReq.timestamp <= latestWM.Timestamp {
				commitReq.timestamp = latestWM.Timestamp + 1
			}
			break
		} else if resp.err != errAlreadySuccessful {
			commitReq.commitMask = commitReq.commitMask.And(zboxutil.NewUint128(1).Lsh(resp.pos).Not())
			if commitReq.commitMask.CountOnes() < commitReq.consensusThresh {
				commitReq.result = ErrorCommitResult("Failed to get reference path " + resp.err.Error())
				return
			}
		}
	}

	if trie == nil {
		commitReq.result = ErrorCommitResult("Failed to get reference path")
		return
	}
	if commitReq.commitMask.CountOnes() < commitReq.consensusThresh {
		commitReq.result = ErrorCommitResult("Failed to get reference path")
		return
	}
	elapsedGetRefPath := time.Since(now)

	for _, change := range commitReq.changes {
		if change == nil {
			continue
		}
		err = change.ProcessChangeV2(trie, changeIndex)
		if err != nil && err != wmpt.ErrNotFound {
			l.Logger.Error("Error processing change ", err)
			commitReq.result = ErrorCommitResult("Failed to process change " + err.Error())
			return
		}
	}
	rootHash := trie.GetRoot().CalcHash()
	rootWeight := trie.Weight()
	pos = 0
	elapsedProcessChanges := time.Since(now) - elapsedGetRefPath
	wg := sync.WaitGroup{}
	errSlice := make([]error, commitReq.commitMask.CountOnes())
	counter := 0
	for i := commitReq.commitMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		blobber := commitReq.allocationObj.Blobbers[pos]
		blobberPos := pos
		wg.Add(1)
		go func(ind int) {
			defer wg.Done()
			err = commitReq.commitBlobber(rootHash, rootWeight, blobberPos, blobber)
			if err != nil {
				l.Logger.Error("Error committing to blobber", err)
				errSlice[ind] = err
				mu.Lock()
				commitReq.commitMask = commitReq.commitMask.And(zboxutil.NewUint128(1).Lsh(pos).Not())
				mu.Unlock()
				return
			}
		}(counter)
		counter++
	}
	wg.Wait()
	elapsedCommit := time.Since(now) - elapsedProcessChanges - elapsedGetRefPath
	if commitReq.commitMask.CountOnes() < commitReq.consensusThresh {
		err = zboxutil.MajorError(errSlice)
		if err == nil {
			err = errors.New("consensus_not_met", fmt.Sprintf("Successfully committed to %d blobbers, but required %d", commitReq.commitMask.CountOnes(), commitReq.consensusThresh))
		}
		commitReq.result = ErrorCommitResult(err.Error())
		return
	}
	if !commitReq.isRepair {
		commitReq.allocationObj.allocationRoot = hex.EncodeToString(rootHash)
	}
	l.Logger.Info("[commit] ", "elapsedGetRefPath ", elapsedGetRefPath.Milliseconds(), " elapsedProcessChanges ", elapsedProcessChanges.Milliseconds(), " elapsedCommit ", elapsedCommit.Milliseconds(), " total ", time.Since(now).Milliseconds())
	commitReq.result = SuccessCommitResult()
}

func (req *CommitRequestV2) commitBlobber(rootHash []byte, rootWeight, changeIndex uint64, blobber *blockchain.StorageNode) (err error) {
	hasher := sha256.New()
	var prevChainSize int64
	if blobber.LatestWM != nil {
		prevChainHash, err := hex.DecodeString(blobber.LatestWM.ChainHash)
		if err != nil {
			l.Logger.Error("Error decoding prev chain hash", err)
			return err
		}
		hasher.Write(prevChainHash) //nolint:errcheck
		prevChainSize = numBlocks(blobber.LatestWM.ChainSize)
	}
	hasher.Write(rootHash) //nolint:errcheck
	chainHash := hex.EncodeToString(hasher.Sum(nil))
	allocationRoot := hex.EncodeToString(rootHash)
	wm := &marker.WriteMarker{}
	wm.AllocationRoot = allocationRoot
	wm.Size = (int64(rootWeight) - prevChainSize) * CHUNK_SIZE
	wm.ChainHash = chainHash
	wm.ChainSize = int64(rootWeight) * CHUNK_SIZE
	if blobber.LatestWM != nil {
		wm.PreviousAllocationRoot = blobber.LatestWM.AllocationRoot
	}
	wm.BlobberID = blobber.ID
	wm.Timestamp = req.timestamp
	wm.AllocationID = req.allocationObj.ID
	wm.FileMetaRoot = allocationRoot
	wm.ClientID = client.GetClientID()
	err = wm.Sign()
	if err != nil {
		l.Logger.Error("Error signing writemarker", err)
		return err
	}
	wmData, err := json.Marshal(wm)
	if err != nil {
		l.Logger.Error("Error marshalling writemarker data", err)
		return err
	}

	err = submitWriteMarker(wmData, nil, blobber, req.connectionID, req.allocationObj.ID, req.allocationObj.Tx, req.allocationObj.StorageVersion)
	if err != nil {
		l.Logger.Error("Error submitting writemarker", err)
		return err
	}
	blobber.LatestWM = wm
	blobber.AllocationRoot = allocationRoot
	return
}

func getFormWritter(connectionID string, wmData, fileIDMetaData []byte, body *bytes.Buffer) (*multipart.Writer, error) {
	formWriter := multipart.NewWriter(body)
	err := formWriter.WriteField("connection_id", connectionID)
	if err != nil {
		return nil, err
	}

	err = formWriter.WriteField("write_marker", string(wmData))
	if err != nil {
		return nil, err
	}
	if len(fileIDMetaData) > 0 {
		err = formWriter.WriteField("file_id_meta", string(fileIDMetaData))
		if err != nil {
			return nil, err
		}
	}
	formWriter.Close()
	return formWriter, nil
}

func getReferencePathV2(blobber *blockchain.StorageNode, allocationID, allocationTx, sig string, paths []string, success *bool, mu *sync.Mutex) (*wmpt.WeightedMerkleTrie, error) {
	if len(paths) == 0 {
		var node wmpt.Node
		if blobber.LatestWM != nil {
			decodedRoot, _ := hex.DecodeString(blobber.LatestWM.AllocationRoot)
			node = wmpt.NewHashNode(decodedRoot, uint64(numBlocks(blobber.LatestWM.ChainSize)))
		}
		trie := wmpt.New(node, nil)
		return trie, nil
	}
	now := time.Now()
	req, err := zboxutil.NewReferencePathRequestV2(blobber.Baseurl, allocationID, allocationTx, sig, paths, false)
	if err != nil {
		l.Logger.Error("Creating ref path req", err)
		return nil, err
	}
	var lR ReferencePathResultV2
	ctx, cncl := context.WithTimeout(context.Background(), (time.Second * 30))
	err = zboxutil.HttpDo(ctx, cncl, req, func(resp *http.Response, err error) error {
		if err != nil {
			l.Logger.Error("Ref path error:", err)
			return err
		}
		defer resp.Body.Close()
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			l.Logger.Error("Ref path: Resp", err)
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return errors.New(
				strconv.Itoa(resp.StatusCode),
				fmt.Sprintf("Reference path error response: Status: %d - %s ",
					resp.StatusCode, string(respBody)))
		}
		err = json.Unmarshal(respBody, &lR)
		if err != nil {
			l.Logger.Error("Reference path json decode error: ", err)
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	elapsedRefPath := time.Since(now)
	mu.Lock()
	defer mu.Unlock()
	if *success {
		return nil, errAlreadySuccessful
	}
	trie := wmpt.New(nil, nil)
	if lR.LatestWM != nil {
		err = lR.LatestWM.VerifySignature(client.GetClientPublicKey())
		if err != nil {
			return nil, errors.New("signature_verification_failed", err.Error())
		}
		err = trie.Deserialize(lR.Path)
		if err != nil {
			l.Logger.Error("Error deserializing trie", err)
			return nil, err
		}
		l.Logger.Info("[getReferencePathV2] elapsedRefPath ", elapsedRefPath.Milliseconds(), " elapsedDeserialize ", (time.Since(now) - elapsedRefPath).Milliseconds())
		chainBlocks := numBlocks(lR.LatestWM.ChainSize)
		if trie.Weight() != uint64(chainBlocks) {
			return nil, errors.New("chain_length_mismatch", fmt.Sprintf("Expected chain length %d, got %d", chainBlocks, trie.Weight()))
		}
		if hex.EncodeToString(trie.Root()) != lR.LatestWM.AllocationRoot {
			return nil, errors.New("allocation_root_mismatch", fmt.Sprintf("Expected allocation root %s, got %s", lR.LatestWM.AllocationRoot, hex.EncodeToString(trie.Root())))
		}
	}
	*success = true
	return trie, nil
}

func submitWriteMarker(wmData, metaData []byte, blobber *blockchain.StorageNode, connectionID, allocationID, allocationTx string, apiVersion int) (err error) {
	var (
		resp           *http.Response
		shouldContinue bool
	)
	for retries := 0; retries < 6; retries++ {
		err, shouldContinue = func() (err error, shouldContinue bool) {
			body := new(bytes.Buffer)
			formWriter, err := getFormWritter(connectionID, wmData, metaData, body)
			if err != nil {
				l.Logger.Error("Creating form writer failed: ", err)
				return
			}
			httpreq, err := zboxutil.NewCommitRequest(blobber.Baseurl, allocationID, allocationTx, body, apiVersion)
			if err != nil {
				l.Logger.Error("Error creating commit req: ", err)
				return
			}
			httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
			reqCtx, ctxCncl := context.WithTimeout(context.Background(), time.Second*60)
			resp, err = zboxutil.Client.Do(httpreq.WithContext(reqCtx))
			defer ctxCncl()

			if err != nil {
				logger.Logger.Error("Commit: ", err)
				return
			}

			if resp.Body != nil {
				defer resp.Body.Close()
			}

			var respBody []byte
			respBody, err = io.ReadAll(resp.Body)
			if err != nil {
				logger.Logger.Error("Response read: ", err)
				return
			}
			if resp.StatusCode == http.StatusOK {
				logger.Logger.Debug(blobber.Baseurl, " committed")
				return
			}

			if resp.StatusCode == http.StatusTooManyRequests {
				logger.Logger.Debug(blobber.Baseurl,
					" got too many request error. Retrying")

				var r int
				r, err = zboxutil.GetRateLimitValue(resp)
				if err != nil {
					logger.Logger.Error(err)
					return
				}

				time.Sleep(time.Duration(r) * time.Second)
				shouldContinue = true
				return
			}

			if strings.Contains(string(respBody), "pending_markers:") {
				logger.Logger.Debug("Commit pending for blobber ",
					blobber.Baseurl, " Retrying")
				time.Sleep(5 * time.Second)
				shouldContinue = true
				return
			}

			if strings.Contains(string(respBody), "chain_length_exceeded") {
				l.Logger.Error("Chain length exceeded for blobber ",
					blobber.Baseurl, " Retrying")
				time.Sleep(5 * time.Second)
				shouldContinue = true
				return
			}

			err = thrown.New("commit_error",
				fmt.Sprintf("Got error response %s with status %d", respBody, resp.StatusCode))
			return
		}()
		if shouldContinue {
			continue
		}
		return
	}
	return thrown.New("commit_error", fmt.Sprintf("Commit failed with response status %d", resp.StatusCode))
}

func numBlocks(size int64) int64 {
	return int64(math.Ceil(float64(size*1.0) / CHUNK_SIZE))
}
