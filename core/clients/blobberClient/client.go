package blobberClient

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"

	blobbergrpc "github.com/0chain/blobber/code/go/0chain.net/blobbercore/blobbergrpc/proto"

	"google.golang.org/grpc/encoding/gzip"

	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/convert"
	blobbercommon "github.com/0chain/blobber/code/go/0chain.net/core/common"
	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/zboxcore/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const GRPCPort = 31501

func newBlobberGRPCClient(urlRaw string) (blobbergrpc.BlobberServiceClient, error) {
	u, err := url.Parse(urlRaw)
	if err != nil {
		return nil, err
	}
	host, _, _ := net.SplitHostPort(u.Host)

	cc, err := grpc.Dial(host+":"+fmt.Sprint(GRPCPort), grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return blobbergrpc.NewBlobberServiceClient(cc), nil
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

func CommitMetaTxn(url string, req *blobbergrpc.CommitMetaTxnRequest) ([]byte, error) {
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

	commitMetaResp, err := blobberClient.CommitMetaTxn(grpcCtx, req)
	if err != nil {
		return nil, err
	}

	return json.Marshal(convert.GetCommitMetaTxnHandlerResponse(commitMetaResp))
}

func Collaborator(url string, req *blobbergrpc.CollaboratorRequest) ([]byte, error) {
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

	collaboratorResp, err := blobberClient.Collaborator(grpcCtx, req)
	if err != nil {
		return nil, err
	}

	return json.Marshal(convert.CollaboratorResponse(collaboratorResp))
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

func CopyObject(url string, req *blobbergrpc.CopyObjectRequest) ([]byte, error) {
	blobberClient, err := newBlobberGRPCClient(url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create blobber grpc client")
	}

	clientSignature, err := client.Sign(encryption.Hash(req.Allocation))
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate client signature")
	}

	grpcCtx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		blobbercommon.ClientHeader:          client.GetClientID(),
		blobbercommon.ClientKeyHeader:       client.GetClientPublicKey(),
		blobbercommon.ClientSignatureHeader: clientSignature,
	}))

	copyObjectResponse, err := blobberClient.CopyObject(grpcCtx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to CopyObject")
	}
	return json.Marshal(convert.CopyObjectResponseHandler(copyObjectResponse))
}
