package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

const CHUNK_SIZE = 64 * 1024

type ListRequest struct {
	allocationID       string
	allocationTx       string
	blobbers           []*blockchain.StorageNode
	remotefilepathhash string
	remotefilepath     string
	authToken          *marker.AuthTicket
	ctx                context.Context
	wg                 *sync.WaitGroup
	Consensus
}

type listResponse struct {
	ref         *fileref.Ref
	responseStr string
	blobberIdx  int
	err         error
}

type ListResult struct {
	Name            string             `json:"name"`
	Path            string             `json:"path,omitempty"`
	Type            string             `json:"type"`
	Size            int64              `json:"size"`
	Hash            string             `json:"hash,omitempty"`
	MimeType        string             `json:"mimetype,omitempty"`
	NumBlocks       int64              `json:"num_blocks"`
	LookupHash      string             `json:"lookup_hash"`
	EncryptionKey   string             `json:"encryption_key"`
	Attributes      fileref.Attributes `json:"attributes"`
	ActualSize      int64              `json:"actual_size"`
	ActualNumBlocks int64              `json:"actual_num_blocks"`
	CreatedAt       string             `json:"created_at"`
	UpdatedAt       string             `json:"updated_at"`
	Children        []*ListResult      `json:"list"`
	Consensus       `json:"-"`
}

func (req *ListRequest) getListInfoFromBlobber(blobber *blockchain.StorageNode, blobberIdx int, rspCh chan<- *listResponse) {
	defer req.wg.Done()
	//body := new(bytes.Buffer)
	//formWriter := multipart.NewWriter(body)

	ref := &fileref.Ref{}
	var s strings.Builder
	var err error
	listRetFn := func() {
		rspCh <- &listResponse{ref: ref, responseStr: s.String(), blobberIdx: blobberIdx, err: err}
	}
	defer listRetFn()

	if len(req.remotefilepath) > 0 {
		req.remotefilepathhash = fileref.GetReferenceLookup(req.allocationID, req.remotefilepath)
	}
	//formWriter.WriteField("path_hash", req.remotefilepathhash)
	//Logger.Info("Path hash for list dir: ", req.remotefilepathhash)

	authTokenBytes := make([]byte, 0)
	if req.authToken != nil {
		authTokenBytes, err = json.Marshal(req.authToken)
		if err != nil {
			Logger.Error(blobber.Baseurl, " creating auth token bytes", err)
			return
		}
		//formWriter.WriteField("auth_token", string(authTokenBytes))
	}

	//formWriter.Close()
	httpreq, err := zboxutil.NewListRequest(blobber.Baseurl, req.allocationTx, req.remotefilepath, req.remotefilepathhash, string(authTokenBytes))
	if err != nil {
		Logger.Error("List info request error: ", err.Error())
		return
	}

	//httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
	ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 30))
	err = zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("List : ", err)
			return err
		}
		defer resp.Body.Close()
		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "Error: Resp")
		}
		s.WriteString(string(resp_body))
		Logger.Debug("List result:", string(resp_body))
		if resp.StatusCode == http.StatusOK {
			listResult := &fileref.ListResult{}
			err = json.Unmarshal(resp_body, listResult)
			if err != nil {
				return errors.Wrap(err, "list entities response parse error:")
			}
			ref, err = listResult.GetDirTree(req.allocationID)
			if err != nil {
				return errors.Wrap(err, "error getting the dir tree from list response:")
			}
			return nil
		}

		return fmt.Errorf("error from server list response: %s", s.String())

	})
}

func (req *ListRequest) getlistFromBlobbers() []*listResponse {
	numList := len(req.blobbers)
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numList)
	rspCh := make(chan *listResponse, numList)
	for i := 0; i < numList; i++ {
		go req.getListInfoFromBlobber(req.blobbers[i], i, rspCh)
	}
	req.wg.Wait()
	listInfos := make([]*listResponse, len(req.blobbers))
	for i := 0; i < numList; i++ {
		listInfos[i] = <-rspCh
	}
	return listInfos
}

func (req *ListRequest) GetListFromBlobbers() *ListResult {
	lR := req.getlistFromBlobbers()
	result := &ListResult{}
	selected := make(map[string]*ListResult)
	childResultMap := make(map[string]*ListResult)
	for i := 0; i < len(lR); i++ {
		req.consensus = 0
		ti := lR[i]
		if ti.err != nil || ti.ref == nil {
			continue
		}

		result.Name = ti.ref.Name
		result.Path = ti.ref.Path
		result.Type = ti.ref.Type
		result.CreatedAt = ti.ref.CreatedAt
		result.UpdatedAt = ti.ref.UpdatedAt
		result.LookupHash = ti.ref.LookupHash
		result.Attributes = ti.ref.Attributes
		result.ActualSize = ti.ref.ActualSize
		result.ActualNumBlocks = 0
		if result.Type == fileref.DIRECTORY {
			result.Size = -1
		}
		if ti.ref.ActualSize > 0 {
			result.ActualNumBlocks = ti.ref.ActualSize / CHUNK_SIZE
		}

		for _, child := range lR[i].ref.Children {
			actualHash := encryption.Hash(child.GetLookupHash())
			if child.GetType() == fileref.FILE {
				actualHash = encryption.Hash(child.GetLookupHash() + ":" + (child.(*fileref.FileRef)).ActualFileHash)
			}
			var childResult *ListResult
			if _, ok := childResultMap[actualHash]; !ok {
				childResult = &ListResult{
					Name:       child.GetName(),
					Path:       child.GetPath(),
					Type:       child.GetType(),
					CreatedAt:  child.GetCreatedAt(),
					UpdatedAt:  child.GetUpdatedAt(),
					Attributes: child.GetAttributes(),
				}
				childResult.LookupHash = child.GetLookupHash()
				childResult.consensus = 0
				childResult.consensusThresh = req.consensusThresh
				childResult.fullconsensus = req.fullconsensus
				childResultMap[actualHash] = childResult
			}
			childResult = childResultMap[actualHash]
			childResult.consensus++
			if child.GetType() == fileref.FILE {
				childResult.Hash = (child.(*fileref.FileRef)).ActualFileHash
				childResult.MimeType = (child.(*fileref.FileRef)).MimeType
				childResult.EncryptionKey = (child.(*fileref.FileRef)).EncryptedKey
				childResult.ActualSize = (child.(*fileref.FileRef)).ActualFileSize
				if childResult.ActualSize > 0 {
					childResult.ActualNumBlocks = childResult.ActualSize / CHUNK_SIZE
				}
			}
			childResult.Size += child.GetSize()
			childResult.NumBlocks += child.GetNumBlocks()
			if childResult.isConsensusOk() {
				if _, ok := selected[child.GetLookupHash()]; !ok {
					result.Children = append(result.Children, childResult)
					selected[child.GetLookupHash()] = childResult
				}
			}
		}

		for _, child := range result.Children {
			result.NumBlocks += child.NumBlocks
			result.Size += child.Size
		}
	}
	return result
}
