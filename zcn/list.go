package zcn

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

	"github.com/0chain/gosdk/util"
)

type listDirData struct {
	AllocationRoot string                   `json:"allocation_root"`
	Meta           map[string]interface{}   `json:"meta_data"`
	Entities       []map[string]interface{} `json:"list"`
}

type listDirResponse struct {
	lsData listDirData
	idx    int
	err    error
}

type listFileData struct {
	Meta  map[string]interface{} `json:"meta"`
	Stats map[string]interface{} `json:"stats"`
}

type listFileResponse struct {
	lsData listFileData
	str    string
	idx    int
	err    error
}

func (obj *Allocation) listDirFromBlobber(blobber *util.Blobber, blobberIdx int, path string, rspCh chan<- *listDirResponse, wg *sync.WaitGroup) {
	wg.Done()
	var lR listDirData
	var err error
	lsRetFn := func() { rspCh <- &listDirResponse{lsData: lR, idx: blobberIdx, err: err} }
	defer lsRetFn()
	var req *http.Request
	req, err = util.NewListRequest(blobber.UrlRoot, obj.allocationId, obj.client, path)
	if err != nil {
		Logger.Error("Creating list req", err)
		return
	}
	ctx, cncl := context.WithTimeout(context.Background(), (time.Second * 30))
	err = httpDo(ctx, cncl, req, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("List:", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			Logger.Error("List response : ", resp.StatusCode)
		}
		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Logger.Error("List: Resp", err)
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("List error response: Status: %d - %s ", resp.StatusCode, string(resp_body))
		} else {
			err = json.Unmarshal(resp_body, &lR)
			if err != nil {
				Logger.Error("List json decode error: ", err)
				return err
			}
		}
		return nil
	})
}

func (obj *Allocation) getBlobberFileStat(blobber *util.Blobber, blobberIdx int, path string, rspCh chan<- *listFileResponse, wg *sync.WaitGroup) {
	wg.Done()
	var lR listFileData
	var s strings.Builder
	var err error
	lsRetFn := func() { rspCh <- &listFileResponse{lsData: lR, str: s.String(), idx: blobberIdx, err: err} }
	defer lsRetFn()
	var req *http.Request
	req, err = util.NewStatsRequest(blobber.UrlRoot, obj.allocationId, obj.client, path)
	if err != nil {
		Logger.Error(" Creating stats req: ", err)
		return
	}
	ctx, cncl := context.WithTimeout(context.Background(), (time.Second * 30))
	err = httpDo(ctx, cncl, req, func(resp *http.Response, err error) error {
		if err != nil {
			return fmt.Errorf(" Stats : %s", err.Error())
		}
		defer resp.Body.Close()
		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf(" Stats Resp: %s", err.Error())
		}
		s.WriteString(string(resp_body))
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf(" Stats error response: Status: %d - %s ", resp.StatusCode, string(resp_body))
		} else {
			err = json.Unmarshal(resp_body, &lR)
			if err != nil {
				return fmt.Errorf(" Stats json decode error: %s", err.Error())
			}
		}
		return nil
	})
}

func (obj *Allocation) getFileStatsFromBlobbers(path string) []*listFileResponse {
	numList := len(obj.blobbers)
	obj.bg.wg.Add(numList)
	rspCh := make(chan *listFileResponse, numList)
	for i := 0; i < numList; i++ {
		go obj.getBlobberFileStat(&obj.blobbers[i], i, path, rspCh, &obj.bg.wg)
	}
	obj.bg.wg.Wait()
	// Collect all the results.
	lR := make([]*listFileResponse, numList)
	for i := 0; i < numList; i++ {
		lR[i] = <-rspCh
	}
	return lR
}

func (obj *Allocation) getFileInfoFromBlobbers(path string) string {
	var s strings.Builder
	lR := obj.getFileStatsFromBlobbers(path)
	fmt.Fprintf(&s, "\n[")
	// Collect all the results.
	for i := 0; i < len(lR); i++ {
		if lR[i].err == nil {
			s.WriteString(fmt.Sprintf("{\n\"blobber\": {\"id\":\"%s\", \"url\":\"%s\"},\n\"data\":%s\n}",
				obj.blobbers[lR[i].idx].Id, obj.blobbers[lR[i].idx].UrlRoot, lR[i].str))
			if i != len(lR)-1 {
				s.WriteString(",\n")
			}
		} else {
			Logger.Error(obj.blobbers[lR[i].idx].UrlRoot, lR[i].err)
		}
	}
	fmt.Fprintf(&s, "]\n")
	return s.String()
}

func (obj *Allocation) getFileConsensusFromBlobbers(path string) uint32 {
	lR := obj.getFileStatsFromBlobbers(path)
	var selected *listFileResponse
	foundMask := uint32(0)
	for i := 0; i < len(lR); i++ {
		ti := lR[i]
		if ti.err != nil {
			continue
		}
		obj.consensus = 1
		for j := (i + 1); j < len(lR); j++ {
			tj := lR[j]
			if tj.err != nil {
				continue
			}
			if ti.lsData.Meta["actual_file_hash"] == tj.lsData.Meta["actual_file_hash"] {
				obj.consensus++
			}
		}
		if obj.isConsensusMin() {
			selected = ti
			break
		}
	}
	if selected == nil {
		Logger.Error("File consensus not found for ", path)
		return foundMask
	}
	for i := 0; i < len(lR); i++ {
		if selected.lsData.Meta["actual_file_hash"] == lR[i].lsData.Meta["actual_file_hash"] {
			foundMask |= (1 << uint32(lR[i].idx))
		}
	}
	return foundMask
}

func (obj *Allocation) getDir(blobber *util.Blobber, blobberIdx int, path string) (listDirData, error) {
	var wg sync.WaitGroup
	rspCh := make(chan *listDirResponse, 1)
	wg.Add(1)
	obj.listDirFromBlobber(blobber, blobberIdx, path, rspCh, &wg)
	wg.Wait()
	result := <-rspCh
	return result.lsData, result.err
}

func (obj *Allocation) syncDir(path string) error {
	numList := len(obj.blobbers)
	obj.bg.wg.Add(numList)
	rspCh := make(chan *listDirResponse, numList)
	for i := 0; i < numList; i++ {
		go obj.listDirFromBlobber(&obj.blobbers[i], i, path, rspCh, &obj.bg.wg)
	}
	obj.bg.wg.Wait()
	var err error
	lR := make([]*listDirResponse, numList)
	// Collect all the results.
	for i := 0; i < numList; i++ {
		lR[i] = <-rspCh
		if lR[i].err == nil {
			blobber := &obj.blobbers[lR[i].idx]
			for _, child := range lR[i].lsData.Entities {
				if child["type"] == "d" {
					util.AddDir(&blobber.ConnObj.DirTree, child["path"].(string))
				} else {
					_, err = util.InsertFile(&blobber.ConnObj.DirTree, child["path"].(string), child["hash"].(string), int64(child["size"].(float64)))
					if err != nil {
						Logger.Error("Add dir/file:", child["path"].(string), err)
					}
				}
			}
			// Calculate hash
			_ = util.CalculateDirHash(&blobber.ConnObj.DirTree)
			blobber.ConnObj.DirTree.Hash = lR[i].lsData.AllocationRoot
			blobber.DirTree = blobber.ConnObj.DirTree
		} else {
			Logger.Error(obj.blobbers[lR[i].idx].UrlRoot, lR[i].err)
		}
	}
	// Check the majority
	var selected *listDirResponse
	for i := 0; i < numList; i++ {
		ti := lR[i]
		if ti.err != nil {
			continue
		}
		obj.consensus = 1
		for j := (i + 1); j < numList; j++ {
			tj := lR[j]
			if tj.err != nil {
				continue
			}
			if len(ti.lsData.Entities) == len(tj.lsData.Entities) &&
				ti.lsData.Meta["path_hash"].(string) == tj.lsData.Meta["path_hash"].(string) {
				obj.consensus++
			}
		}
		if obj.isConsensusOk() {
			selected = ti
			break
		}
	}
	if selected == nil {
		return fmt.Errorf("Common info for %s path not found from blobbers", path)
	}
	// Update the master dir structure
	for _, child := range selected.lsData.Entities {
		if child["type"] == "d" {
			util.AddDir(&obj.dirTree, child["path"].(string))
			obj.syncDir(child["path"].(string))
		} else {
			fl, err := util.InsertFile(&obj.dirTree, child["path"].(string), child["actual_file_hash"].(string), int64(child["actual_file_size"].(float64)))
			if err != nil {
				Logger.Error("Error inserting file to master dir structure: ", err.Error())
			} else {
				fl.Meta[fileMetaBlobberCount] = obj.consensus
			}
		}
	}

	return nil
}

func (obj *Allocation) syncAllBlobbers() {
	// Reset the dir
	obj.dirTree = util.NewDirTree()
	// Create new connection object tree
	for i := 0; i < len(obj.blobbers); i++ {
		obj.blobbers[i].ConnObj.DirTree = util.NewDirTree()
	}
	obj.syncDir("/")
}

func (obj *Allocation) getFileMetaInfoFromBlobber(blobber *util.Blobber, remotePath string, authToken *authTicket, isPathHash bool) (*util.FileDirInfo, error) {
	defer obj.wg.Done()
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)

	authTokenBytes, err := json.Marshal(authToken)
	if err != nil {
		Logger.Error(blobber.UrlRoot, " creating auth token bytes", err)
		return nil, err
	}
	if isPathHash {
		formWriter.WriteField("path_hash", remotePath)
	} else {
		formWriter.WriteField("path", remotePath)
	}

	if authToken != nil {
		formWriter.WriteField("auth_token", string(authTokenBytes))
	}

	formWriter.Close()
	req, err := util.NewFileMetaRequest(blobber.UrlRoot, obj.allocationId, obj.client, body)
	if err != nil {
		return nil, fmt.Errorf("File meta info request error: %s", err.Error())
	}

	req.Header.Add("Content-Type", formWriter.FormDataContentType())
	ctx, cncl := context.WithTimeout(context.Background(), (time.Second * 30))
	var retFileInfo *util.FileDirInfo
	err = httpDo(ctx, cncl, req, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("GetFileMeta : ", err)
			return err
		}
		defer resp.Body.Close()
		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Error: Resp : %s", err.Error())
		}
		meta_data := make(map[string]interface{})
		if resp.StatusCode == http.StatusOK {
			fmt.Println("received from blobber " + string(resp_body))
			err = json.Unmarshal(resp_body, &meta_data)
			if err != nil {
				return fmt.Errorf("file meta data response parse error: %s", err.Error())
			}
			fInfo := &util.FileDirInfo{
				Type: meta_data["type"].(string),
				Name: meta_data["name"].(string),
				Hash: meta_data["actual_file_hash"].(string),
				Size: int64(meta_data["actual_file_size"].(float64)),
			}
			retFileInfo = fInfo
			return nil
		}

		return fmt.Errorf("%s Response Error: %s", blobber.UrlRoot, string(resp_body))
	})

	return retFileInfo, err
}

func (obj *Allocation) GetFileStats(remotePath string) string {
	return obj.getFileInfoFromBlobbers(remotePath)
}
