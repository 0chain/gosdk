package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"sync"

	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/handler"

	"google.golang.org/grpc"

	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/blobbergrpc"

	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
)

func getObjectTreeFromBlobber(ctx context.Context, allocationID, allocationTx string, remotefilepath string, blobber *blockchain.StorageNode) (fileref.RefEntity, error) {
	blobberClient, err := NewBlobberGRPCClient(blobber.Baseurl)
	if err != nil {
		return nil, err
	}

	getObjectTreeResp, err := blobberClient.GetObjectTree(context.Background(), &blobbergrpc.GetObjectTreeRequest{
		Context: &blobbergrpc.RequestContext{
			Client:          "",
			ClientKey:       "",
			Allocation:      allocationTx,
			ClientSignature: "",
		},
		Path:       remotefilepath,
		Allocation: allocationTx,
	})
	if err != nil {
		return nil, err
	}

	var lR ReferencePathResult
	respRaw, err := json.Marshal(handler.GetObjectTreeResponseHandler(getObjectTreeResp))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(respRaw, &lR)
	if err != nil {
		return nil, err
	}

	return lR.GetRefFromObjectTree(allocationID)
}

const GRPCPort = 7777

func NewBlobberGRPCClient(urlRaw string) (blobbergrpc.BlobberClient, error) {
	u, err := url.Parse(urlRaw)
	if err != nil {
		return nil, err
	}
	host, _, _ := net.SplitHostPort(u.Host)

	cc, err := grpc.Dial(host+":"+fmt.Sprint(GRPCPort), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return blobbergrpc.NewBlobberClient(cc), nil
}

func getAllocationDataFromBlobber(blobber *blockchain.StorageNode, allocationTx string, respCh chan<- *BlobberAllocationStats, wg *sync.WaitGroup) {
	defer wg.Done()
	blobberClient, err := NewBlobberGRPCClient(blobber.Baseurl)
	if err != nil {
		return
	}

	getAllocationResp, err := blobberClient.GetAllocation(context.Background(), &blobbergrpc.GetAllocationRequest{
		Context: &blobbergrpc.RequestContext{
			Client:          "",
			ClientKey:       "",
			Allocation:      allocationTx,
			ClientSignature: "",
		},
		Id: allocationTx,
	})
	if err != nil {
		return
	}

	respRaw, err := json.Marshal(handler.GetAllocationResponseHandler(getAllocationResp))
	if err != nil {
		return
	}

	var result BlobberAllocationStats
	err = json.Unmarshal(respRaw, &result)
	if err != nil {
		return
	}

	result.BlobberID = blobber.ID
	result.BlobberURL = blobber.Baseurl
	respCh <- &result
	return
}
