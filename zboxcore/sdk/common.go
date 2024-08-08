package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

const alreadyExists = "file already exists"

func getObjectTreeFromBlobber(ctx context.Context, allocationID, allocationTx string, remoteFilePath string, blobber *blockchain.StorageNode) (fileref.RefEntity, error) {
	httpreq, err := zboxutil.NewObjectTreeRequest(blobber.Baseurl, allocationID, allocationTx, remoteFilePath)
	if err != nil {
		l.Logger.Error(blobber.Baseurl, "Error creating object tree request", err)
		return nil, err
	}
	var lR ReferencePathResult
	ctx, cncl := context.WithTimeout(ctx, (time.Second * 60))
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

var ErrFileNameTooLong = errors.New("invalid_parameter", "filename is longer than 150 characters")

func ValidateRemoteFileName(remotePath string) error {
	_, fileName := path.Split(remotePath)

	if len(fileName) > 150 {
		return ErrFileNameTooLong
	}

	return nil
}

type subDirRequest struct {
	opType          string
	remotefilepath  string
	destPath        string
	allocationObj   *Allocation
	ctx             context.Context
	consensusThresh int
	mask            zboxutil.Uint128
}

func (req *subDirRequest) processSubDirectories() error {
	var (
		offsetPath string
		pathLevel  int
	)

	for {
		oResult, err := req.allocationObj.GetRefs(req.remotefilepath, offsetPath, "", "", fileref.FILE, fileref.REGULAR, 0, getRefPageLimit, WithObjectContext(req.ctx), WithObjectConsensusThresh(req.consensusThresh), WithSingleBlobber(true), WithObjectMask(req.mask))
		if err != nil {
			return err
		}
		if len(oResult.Refs) == 0 {
			break
		}
		ops := make([]OperationRequest, 0, len(oResult.Refs))
		for _, ref := range oResult.Refs {
			opMask := req.mask
			if ref.Type == fileref.DIRECTORY {
				continue
			}
			if ref.PathLevel > pathLevel {
				pathLevel = ref.PathLevel
			}
			destPath := filepath.Dir(strings.Replace(ref.Path, req.remotefilepath, req.destPath, 1))
			op := OperationRequest{
				OperationType: req.opType,
				RemotePath:    ref.Path,
				DestPath:      destPath,
				Mask:          &opMask,
			}
			ops = append(ops, op)
		}
		err = req.allocationObj.DoMultiOperation(ops)
		if err != nil {
			return err
		}
		offsetPath = oResult.Refs[len(oResult.Refs)-1].Path
		if len(oResult.Refs) < getRefPageLimit {
			break
		}
	}

	offsetPath = ""
	level := len(strings.Split(strings.TrimSuffix(req.remotefilepath, "/"), "/"))
	if pathLevel == 0 {
		pathLevel = level + 1
	}

	for pathLevel > level {
		oResult, err := req.allocationObj.GetRefs(req.remotefilepath, offsetPath, "", "", fileref.DIRECTORY, fileref.REGULAR, pathLevel, getRefPageLimit, WithObjectContext(req.ctx), WithObjectMask(req.mask), WithObjectConsensusThresh(req.consensusThresh), WithSingleBlobber(true))
		if err != nil {
			return err
		}
		if len(oResult.Refs) == 0 {
			pathLevel--
		} else {
			ops := make([]OperationRequest, 0, len(oResult.Refs))
			for _, ref := range oResult.Refs {
				opMask := req.mask
				if ref.Type == fileref.FILE {
					continue
				}
				destPath := filepath.Dir(strings.Replace(ref.Path, req.remotefilepath, req.destPath, 1))
				op := OperationRequest{
					OperationType: req.opType,
					RemotePath:    ref.Path,
					DestPath:      destPath,
					Mask:          &opMask,
				}
				ops = append(ops, op)
			}
			err = req.allocationObj.DoMultiOperation(ops)
			if err != nil {
				return err
			}
			offsetPath = oResult.Refs[len(oResult.Refs)-1].Path
			if len(oResult.Refs) < getRefPageLimit {
				pathLevel--
			}
		}
	}

	return nil
}
