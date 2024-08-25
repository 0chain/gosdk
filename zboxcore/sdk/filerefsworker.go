package sdk

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/logger"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type ObjectTreeResult struct {
	TotalPages int64  `json:"total_pages"`
	OffsetPath string `json:"offset_path"`
	OffsetDate string `json:"offset_date"`
	Refs       []ORef `json:"refs"`
}

const INVALID_PATH = "invalid_path"

type ObjectTreeRequest struct {
	allocationID   string
	allocationTx   string
	blobbers       []*blockchain.StorageNode
	authToken      string
	pathHash       string
	remotefilepath string
	pageLimit      int // numbers of refs that will be returned by blobber at max
	level          int
	fileType       string
	refType        string
	offsetPath     string
	updatedDate    string // must have "2006-01-02T15:04:05.99999Z07:00" format
	offsetDate     string // must have "2006-01-02T15:04:05.99999Z07:00" format
	reqMask        zboxutil.Uint128
	ctx            context.Context
	singleBlobber  bool
	Consensus
}

type oTreeResponse struct {
	oTResult *ObjectTreeResult
	err      error
	hash     string
	idx      int
}

var errTooManyRequests = errors.New("too_many_requests", "Too many requests")

type ObjectTreeRequestOption func(*ObjectTreeRequest)

func WithObjectContext(ctx context.Context) ObjectTreeRequestOption {
	return func(o *ObjectTreeRequest) {
		o.ctx = ctx
	}
}

func WithObjectMask(mask zboxutil.Uint128) ObjectTreeRequestOption {
	return func(o *ObjectTreeRequest) {
		o.reqMask = mask
	}
}

func WithObjectConsensusThresh(thresh int) ObjectTreeRequestOption {
	return func(o *ObjectTreeRequest) {
		o.consensusThresh = thresh
	}
}

func WithSingleBlobber(singleBlobber bool) ObjectTreeRequestOption {
	return func(o *ObjectTreeRequest) {
		o.singleBlobber = singleBlobber
	}
}

// Paginated tree should not be collected as this will stall the client
// It should rather be handled by application that uses gosdk
func (o *ObjectTreeRequest) GetRefs() (*ObjectTreeResult, error) {
	activeCount := o.reqMask.CountOnes()
	oTreeResponses := make([]oTreeResponse, activeCount)
	respChan := make(chan *oTreeResponse, activeCount)
	if o.singleBlobber {
		var respErr error
		for i := 0; i < activeCount; i++ {
			var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
			num := rnd.Intn(activeCount)
			var blob *blockchain.StorageNode
			var pos uint64
			for i := o.reqMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
				pos = uint64(i.TrailingZeros())
				num--
				if num < 0 {
					blob = o.blobbers[pos]
					break
				}
			}
			l.Logger.Debug(fmt.Sprintf("Getting file refs for path %v from blobber %v", o.remotefilepath, blob.Baseurl))
			idx := num
			baseURL := blob.Baseurl
			go o.getFileRefs(baseURL, respChan, idx)
			select {
			case <-o.ctx.Done():
				return nil, o.ctx.Err()
			case oTreeResponse := <-respChan:
				if oTreeResponse.err != nil {
					respErr = oTreeResponse.err
				} else {
					return oTreeResponse.oTResult, nil
				}
			}
		}
		return nil, respErr
	}
	var pos uint64
	for i := o.reqMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		blob := o.blobbers[pos]
		l.Logger.Debug(fmt.Sprintf("Getting file refs for path %v from blobber %v", o.remotefilepath, blob.Baseurl))
		idx := int(pos)
		baseURL := blob.Baseurl
		go o.getFileRefs(baseURL, respChan, idx)
	}

	hashCount := make(map[string]int)
	hashRefsMap := make(map[string]*ObjectTreeResult)
	oTreeResponseErrors := make([]error, activeCount)
	var successCount int
	for i := 0; i < activeCount; i++ {
		select {
		case <-o.ctx.Done():
			return nil, o.ctx.Err()
		case oTreeResponse := <-respChan:
			oTreeResponseErrors[oTreeResponse.idx] = oTreeResponse.err
			if oTreeResponse.err != nil {
				if code, _ := zboxutil.GetErrorMessageCode(oTreeResponse.err.Error()); code != INVALID_PATH {
					l.Logger.Error("Error while getting file refs from blobber:", oTreeResponse.err)
				}
				continue
			}
			oTreeResponses[oTreeResponse.idx] = *oTreeResponse
			successCount++
			hash := oTreeResponse.hash
			if _, ok := hashCount[hash]; ok {
				hashCount[hash]++
			} else {
				hashCount[hash]++
				hashRefsMap[hash] = oTreeResponse.oTResult
			}
			if hashCount[hash] == o.consensusThresh {
				return oTreeResponse.oTResult, nil
			}
		}
	}
	l.Logger.Error("no consensus found: ", o.remotefilepath, " pageLimit: ", o.pageLimit, " offsetPath: ", o.offsetPath)
	var selected *ObjectTreeResult
	if successCount < o.consensusThresh {
		majorError := zboxutil.MajorError(oTreeResponseErrors)
		majorErrorMsg := ""
		if majorError != nil {
			majorErrorMsg = majorError.Error()
		}
		if code, _ := zboxutil.GetErrorMessageCode(majorErrorMsg); code == INVALID_PATH {
			return &ObjectTreeResult{}, nil
		} else {
			return nil, majorError
		}
	}
	// build the object tree result by using consensus on individual refs
	refHash := make(map[string]int)
	selected = &ObjectTreeResult{}
	minPage := int64(math.MaxInt64)
	for _, oTreeResponse := range oTreeResponses {
		if oTreeResponse.err != nil || oTreeResponse.oTResult == nil {
			continue
		}
		if oTreeResponse.oTResult.TotalPages < minPage {
			minPage = oTreeResponse.oTResult.TotalPages
			selected.TotalPages = minPage
		}
		for _, ref := range oTreeResponse.oTResult.Refs {
			if refHash[ref.FileMetaHash] == o.consensusThresh {
				continue
			}
			refHash[ref.FileMetaHash] += 1
			if refHash[ref.FileMetaHash] == o.consensusThresh {
				selected.Refs = append(selected.Refs, ref)
			}
		}
	}
	if len(selected.Refs) > 0 {
		for _, oTreeResponse := range oTreeResponses {
			if oTreeResponse.err != nil || oTreeResponse.oTResult == nil {
				continue
			}
			if len(selected.Refs) != len(oTreeResponse.oTResult.Refs) {
				l.Logger.Error("Consensus failed for refs: ", o.blobbers[oTreeResponse.idx].Baseurl)
			}
		}
		selected.OffsetPath = selected.Refs[len(selected.Refs)-1].Path
		return selected, nil
	}
	return nil, errors.New("consensus_failed", "Refs consensus is less than consensus threshold")
}

func (o *ObjectTreeRequest) getFileRefs(bUrl string, respChan chan *oTreeResponse, idx int) {
	oTR := &oTreeResponse{
		idx: idx,
	}
	defer func() {
		respChan <- oTR
	}()

	oResult := ObjectTreeResult{}
	for i := 0; i < 3; i++ {
		oReq, err := zboxutil.NewRefsRequest(
			bUrl,
			o.allocationID,
			o.allocationTx,
			o.remotefilepath,
			o.pathHash,
			o.authToken,
			o.offsetPath,
			o.updatedDate,
			o.offsetDate,
			o.fileType,
			o.refType,
			o.level,
			o.pageLimit,
		)
		if err != nil {
			oTR.err = err
			return
		}
		ctx, cncl := context.WithTimeout(o.ctx, 2*time.Minute)
		defer cncl()
		err = zboxutil.HttpDo(ctx, cncl, oReq, func(resp *http.Response, err error) error {
			if err != nil {
				l.Logger.Error(err)
				return err
			}
			defer resp.Body.Close()
			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				l.Logger.Error(err)
				return err
			}
			if resp.StatusCode == http.StatusOK {
				err := json.Unmarshal(respBody, &oResult)
				if err != nil {
					l.Logger.Error(err)
					return err
				}
				return nil
			} else {
				if resp.StatusCode == http.StatusTooManyRequests {
					l.Logger.Error("Too many requests")
					r, err := zboxutil.GetRateLimitValue(resp)
					if err != nil {
						logger.Logger.Error(err)
						return err
					}

					time.Sleep(time.Duration(r) * time.Second)
					return errTooManyRequests
				}

				return errors.New("response_error", fmt.Sprintf("got status %d, err: %s", resp.StatusCode, respBody))
			}
		})
		if err != nil {
			if err == errTooManyRequests && i < 2 {
				continue
			}
			oTR.err = err
			return
		}
		break
	}
	oTR.oTResult = &oResult
	similarFieldRefs := make([]byte, 0, 32*len(oResult.Refs))
	for _, ref := range oResult.Refs {
		decodeBytes, _ := hex.DecodeString(ref.SimilarField.FileMetaHash)
		similarFieldRefs = append(similarFieldRefs, decodeBytes...)
	}
	oTR.hash = zboxutil.GetRefsHash(similarFieldRefs)
}

// Blobber response will be different from each other so we should only consider similar fields
// i.e. we cannot calculate hash of response and have consensus on it
type ORef struct {
	SimilarField
	ID        int64            `json:"id"`
	CreatedAt common.Timestamp `json:"created_at"`
	UpdatedAt common.Timestamp `json:"updated_at"`
	Err       error            `json:"-"`
}

type SimilarField struct {
	FileID              string `json:"file_id"`
	FileMetaHash        string `json:"file_meta_hash"`
	Type                string `json:"type"`
	AllocationID        string `json:"allocation_id"`
	LookupHash          string `json:"lookup_hash"`
	Name                string `json:"name"`
	Path                string `json:"path"`
	PathHash            string `json:"path_hash"`
	ParentPath          string `json:"parent_path"`
	PathLevel           int    `json:"level"`
	Size                int64  `json:"size"`
	EncryptedKey        string `json:"encrypted_key"`
	EncryptedKeyPoint   string `json:"encrypted_key_point"`
	ActualFileSize      int64  `json:"actual_file_size"`
	ActualFileHash      string `json:"actual_file_hash"`
	MimeType            string `json:"mimetype"`
	ActualThumbnailSize int64  `json:"actual_thumbnail_size"`
	ActualThumbnailHash string `json:"actual_thumbnail_hash"`
	CustomMeta          string `json:"custom_meta"`
}

type RecentlyAddedRefRequest struct {
	ctx          context.Context
	allocationID string
	allocationTx string
	blobbers     []*blockchain.StorageNode
	fromDate     int64
	offset       int64
	pageLimit    int
	wg           *sync.WaitGroup
	Consensus
}

type RecentlyAddedRefResult struct {
	Offset int    `json:"offset"`
	Refs   []ORef `json:"refs"`
}

type RecentlyAddedRefResponse struct {
	Result *RecentlyAddedRefResult
	err    error
}

func (r *RecentlyAddedRefRequest) GetRecentlyAddedRefs() (*RecentlyAddedRefResult, error) {
	totalBlobbers := len(r.blobbers)
	responses := make([]*RecentlyAddedRefResponse, totalBlobbers)
	for i := range responses {
		responses[i] = &RecentlyAddedRefResponse{}
	}
	r.wg.Add(totalBlobbers)

	for i, blob := range r.blobbers {
		go r.getRecentlyAddedRefs(responses[i], blob.Baseurl)
	}
	r.wg.Wait()

	hashCount := make(map[string]int)
	hashRefsMap := make(map[string]*RecentlyAddedRefResult)

	for _, response := range responses {
		if response.err != nil {
			l.Logger.Error(response.err)
			continue
		}

		var similarFieldRefs []SimilarField
		for _, ref := range response.Result.Refs {
			similarFieldRefs = append(similarFieldRefs, ref.SimilarField)
		}

		refsMarshall, err := json.Marshal(similarFieldRefs)
		if err != nil {
			l.Logger.Error(err)
			continue
		}

		hash := zboxutil.GetRefsHash(refsMarshall)
		if _, ok := hashCount[hash]; ok {
			hashCount[hash]++
		} else {
			hashCount[hash]++
			hashRefsMap[hash] = response.Result
		}
	}

	var selected *RecentlyAddedRefResult
	for k, v := range hashCount {
		if v >= r.consensusThresh {
			selected = hashRefsMap[k]
			break
		}
	}

	if selected == nil {
		return nil, errors.New("consensus_failed", "Refs consensus is less than consensus threshold")
	}
	return selected, nil
}

func (r *RecentlyAddedRefRequest) getRecentlyAddedRefs(resp *RecentlyAddedRefResponse, bUrl string) {
	defer r.wg.Done()
	req, err := zboxutil.NewRecentlyAddedRefsRequest(bUrl, r.allocationID, r.allocationTx, r.fromDate, r.offset, r.pageLimit)
	if err != nil {
		resp.err = err
		return
	}

	result := RecentlyAddedRefResult{}
	ctx, cncl := context.WithTimeout(r.ctx, time.Second*30)
	err = zboxutil.HttpDo(ctx, cncl, req, func(hResp *http.Response, err error) error {
		if err != nil {
			l.Logger.Error(err)
			return err
		}
		defer hResp.Body.Close()
		body, err := ioutil.ReadAll(hResp.Body)
		if err != nil {
			l.Logger.Error(err)
			return err
		}
		if hResp.StatusCode != http.StatusOK {
			return fmt.Errorf("Want code %d, got %d. Message: %s",
				http.StatusOK, hResp.StatusCode, string(body))
		}
		err = json.Unmarshal(body, &result)
		if err != nil {
			l.Logger.Error(err)
		}
		return err

	})
	if err != nil {
		resp.err = err
		return
	}
	resp.Result = &result
}
