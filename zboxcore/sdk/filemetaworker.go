package sdk

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	blobbercommon "github.com/0chain/blobber/code/go/0chain.net/core/common"
	"github.com/0chain/gosdk/zboxcore/client"
	"google.golang.org/grpc/metadata"

	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/handler"

	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/blobbergrpc"

	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type fileMetaResponse struct {
	fileref     *fileref.FileRef
	responseStr string
	blobberIdx  int
	err         error
}

func (req *ListRequest) getFileMetaInfoFromBlobber(blobber *blockchain.StorageNode, blobberIdx int, rspCh chan<- *fileMetaResponse) {
	defer req.wg.Done()

	var fileRef *fileref.FileRef
	var s strings.Builder
	var err error
	grpcReq := &blobbergrpc.GetFileMetaDataRequest{
		Allocation: req.allocationTx,
	}

	fileMetaRetFn := func() {
		rspCh <- &fileMetaResponse{fileref: fileRef, responseStr: s.String(), blobberIdx: blobberIdx, err: err}
	}
	defer fileMetaRetFn()
	if len(req.remotefilepath) > 0 {
		req.remotefilepathhash = fileref.GetReferenceLookup(req.allocationID, req.remotefilepath)
	}
	grpcReq.PathHash = req.remotefilepathhash
	if req.authToken != nil {
		authTokenBytes, err := json.Marshal(req.authToken)
		if err != nil {
			Logger.Error(blobber.Baseurl, " creating auth token bytes", err)
			return
		}
		grpcReq.AuthToken = string(authTokenBytes)
	}

	blobberClient, err := NewBlobberGRPCClient(blobber.Baseurl)
	if err != nil {
		return
	}

	grpcCtx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		blobbercommon.ClientHeader:          client.GetClientID(),
		blobbercommon.ClientKeyHeader:       client.GetClientPublicKey(),
		blobbercommon.ClientSignatureHeader: "",
	}))

	getFileMetaDataResp, err := blobberClient.GetFileMetaData(grpcCtx, grpcReq)
	if err != nil {
		return
	}

	respRaw, err := json.Marshal(handler.GetFileMetaDataResponseHandler(getFileMetaDataResp))
	if err != nil {
		return
	}
	s.WriteString(string(respRaw))

	err = json.Unmarshal(respRaw, &fileRef)
	if err != nil {
		return
	}
}

func (req *ListRequest) getFileMetaFromBlobbers() []*fileMetaResponse {
	numList := len(req.blobbers)
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numList)
	rspCh := make(chan *fileMetaResponse, numList)
	for i := 0; i < numList; i++ {
		go req.getFileMetaInfoFromBlobber(req.blobbers[i], i, rspCh)
	}
	req.wg.Wait()
	fileInfos := make([]*fileMetaResponse, len(req.blobbers))
	for i := 0; i < numList; i++ {
		ch := <-rspCh
		fileInfos[ch.blobberIdx] = ch
	}
	return fileInfos
}

func (req *ListRequest) getFileConsensusFromBlobbers() (zboxutil.Uint128, *fileref.FileRef, []*fileMetaResponse) {
	lR := req.getFileMetaFromBlobbers()
	var selected *fileMetaResponse
	foundMask := zboxutil.NewUint128(0)
	req.consensus = 0
	retMap := make(map[string]float32)
	for i := 0; i < len(lR); i++ {
		ti := lR[i]
		if ti.err != nil || ti.fileref == nil {
			continue
		}
		actualHash := ti.fileref.ActualFileHash
		retMap[actualHash]++
		if retMap[actualHash] > req.consensus {
			req.consensus = retMap[actualHash]
			selected = ti
		}
		if req.isConsensusOk() {
			selected = ti
			break
		} else {
			selected = nil
		}
	}
	if selected == nil {
		Logger.Error("File consensus not found for ", req.remotefilepath)
		return foundMask, nil, nil
	}

	for i := 0; i < len(lR); i++ {
		if lR[i].fileref != nil && selected.fileref.ActualFileHash == lR[i].fileref.ActualFileHash {
			shift := zboxutil.NewUint128(1).Lsh(uint64(lR[i].blobberIdx))
			foundMask = foundMask.Or(shift)
		}
	}
	return foundMask, selected.fileref, lR
}
