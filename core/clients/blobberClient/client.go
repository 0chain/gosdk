package blobberClient

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"

	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/blobbergrpc"
	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/convert"
	blobbercommon "github.com/0chain/blobber/code/go/0chain.net/core/common"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/zboxcore/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const GRPCPort = 7031

func newBlobberGRPCClient(urlRaw string) (blobbergrpc.BlobberClient, error) {
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

func Commit(url string, req *blobbergrpc.CommitRequest) ([]byte, error) {
	clientSignature, err := client.Sign(encryption.Hash(req.Allocation))
	if err != nil {
		return nil, err
	}

	grpcCtx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		blobbercommon.ClientHeader:          client.GetClientID(),
		blobbercommon.ClientKeyHeader:       client.GetClientPublicKey(),
		blobbercommon.ClientSignatureHeader: clientSignature,
	}))

	blobberClient, err := newBlobberGRPCClient(url)
	if err != nil {
		return nil, err
	}

	commitResp, err := blobberClient.Commit(grpcCtx, req)
	if err != nil {
		return nil, err
	}

	return json.Marshal(convert.CommitWriteResponseHandler(commitResp))
}

func GetAllocation(url string, req *blobbergrpc.GetAllocationRequest) ([]byte, error) {
	grpcCtx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		blobbercommon.ClientHeader:          client.GetClientID(),
		blobbercommon.ClientKeyHeader:       client.GetClientPublicKey(),
		blobbercommon.ClientSignatureHeader: "",
	}))

	blobberClient, err := newBlobberGRPCClient(url)
	if err != nil {
		return nil, err
	}

	getAllocationResp, err := blobberClient.GetAllocation(grpcCtx, req)
	if err != nil {
		return nil, err
	}

	return json.Marshal(convert.GetAllocationResponseHandler(getAllocationResp))
}

func GetObjectTree(url string, req *blobbergrpc.GetObjectTreeRequest) ([]byte, error) {

	blobberClient, err := newBlobberGRPCClient(url)
	if err != nil {
		return nil, err
	}

	clientSignature, err := client.Sign(encryption.Hash(req.Allocation))
	if err != nil {
		return nil, err
	}

	grpcCtx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		blobbercommon.ClientHeader:          client.GetClientID(),
		blobbercommon.ClientKeyHeader:       client.GetClientPublicKey(),
		blobbercommon.ClientSignatureHeader: clientSignature,
	}))

	getObjectTreeResp, err := blobberClient.GetObjectTree(grpcCtx, req)
	if err != nil {
		return nil, err
	}

	return json.Marshal(convert.GetObjectTreeResponseHandler(getObjectTreeResp))
}

func GetReferencePath(url string, req *blobbergrpc.GetReferencePathRequest) ([]byte, error) {

	blobberClient, err := newBlobberGRPCClient(url)
	if err != nil {
		return nil, err
	}

	clientSignature, err := client.Sign(encryption.Hash(req.Allocation))
	if err != nil {
		return nil, err
	}

	grpcCtx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		blobbercommon.ClientHeader:          client.GetClientID(),
		blobbercommon.ClientKeyHeader:       client.GetClientPublicKey(),
		blobbercommon.ClientSignatureHeader: clientSignature,
	}))

	getReferencePathResp, err := blobberClient.GetReferencePath(grpcCtx, req)
	if err != nil {
		return nil, err
	}

	return json.Marshal(convert.GetReferencePathResponseHandler(getReferencePathResp))
}

func ListEntities(url string, req *blobbergrpc.ListEntitiesRequest) ([]byte, error) {
	blobberClient, err := newBlobberGRPCClient(url)
	if err != nil {
		return nil, err
	}

	grpcCtx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		blobbercommon.ClientHeader:          client.GetClientID(),
		blobbercommon.ClientKeyHeader:       client.GetClientPublicKey(),
		blobbercommon.ClientSignatureHeader: "",
	}))

	listEntitiesResp, err := blobberClient.ListEntities(grpcCtx, req)
	if err != nil {
		return nil, err
	}

	return json.Marshal(convert.ListEntitesResponseHandler(listEntitiesResp))
}

func GetFileStats(url string, req *blobbergrpc.GetFileStatsRequest) ([]byte, error) {
	blobberClient, err := newBlobberGRPCClient(url)
	if err != nil {
		return nil, err
	}

	clientSignature, err := client.Sign(encryption.Hash(req.Allocation))
	if err != nil {
		return nil, err
	}

	grpcCtx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		blobbercommon.ClientHeader:          client.GetClientID(),
		blobbercommon.ClientKeyHeader:       client.GetClientPublicKey(),
		blobbercommon.ClientSignatureHeader: clientSignature,
	}))

	getFileStatsResp, err := blobberClient.GetFileStats(grpcCtx, req)
	if err != nil {
		return nil, err
	}

	return json.Marshal(convert.GetFileStatsResponseHandler(getFileStatsResp))
}

func GetFileMetaData(url string, req *blobbergrpc.GetFileMetaDataRequest) ([]byte, error) {
	blobberClient, err := newBlobberGRPCClient(url)
	if err != nil {
		return nil, err
	}

	grpcCtx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		blobbercommon.ClientHeader:          client.GetClientID(),
		blobbercommon.ClientKeyHeader:       client.GetClientPublicKey(),
		blobbercommon.ClientSignatureHeader: "",
	}))

	getFileMetaDataResp, err := blobberClient.GetFileMetaData(grpcCtx, req)
	if err != nil {
		return nil, err
	}

	return json.Marshal(convert.GetFileMetaDataResponseHandler(getFileMetaDataResp))
}

func CalculateHash(url string, req *blobbergrpc.CalculateHashRequest) ([]byte, error) {
	blobberClient, err := newBlobberGRPCClient(url)
	if err != nil {
		return nil, err
	}

	clientSignature, err := client.Sign(encryption.Hash(req.Allocation))
	if err != nil {
		return nil, err
	}

	grpcCtx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		blobbercommon.ClientHeader:          client.GetClientID(),
		blobbercommon.ClientKeyHeader:       client.GetClientPublicKey(),
		blobbercommon.ClientSignatureHeader: clientSignature,
	}))

	calculateHashResp, err := blobberClient.CalculateHash(grpcCtx, req)
	if err != nil {
		return nil, err
	}

	return json.Marshal(convert.GetCalculateHashResponseHandler(calculateHashResp))
}
