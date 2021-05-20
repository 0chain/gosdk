package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"sync"

	blobbercommon "github.com/0chain/blobber/code/go/0chain.net/core/common"
	"github.com/0chain/gosdk/zboxcore/client"
	"google.golang.org/grpc/metadata"

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

	clientSignature, err := client.Sign(allocationTx)
	if err != nil {
		return nil, err
	}

	grpcCtx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		blobbercommon.ClientHeader:          client.GetClientID(),
		blobbercommon.ClientKeyHeader:       client.GetClientPublicKey(),
		blobbercommon.ClientSignatureHeader: clientSignature,
	}))

	getObjectTreeResp, err := blobberClient.GetObjectTree(grpcCtx, &blobbergrpc.GetObjectTreeRequest{
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

	grpcCtx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		blobbercommon.ClientHeader:          client.GetClientID(),
		blobbercommon.ClientKeyHeader:       client.GetClientPublicKey(),
		blobbercommon.ClientSignatureHeader: "",
	}))

	blobberClient, err := NewBlobberGRPCClient(blobber.Baseurl)
	if err != nil {
		return
	}

	getAllocationResp, err := blobberClient.GetAllocation(grpcCtx, &blobbergrpc.GetAllocationRequest{
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
