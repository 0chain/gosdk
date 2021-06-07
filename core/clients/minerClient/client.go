package minerClient

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"

	"github.com/0chain/0chain/code/go/0chain.net/chaincore/chain"
	"github.com/0chain/0chain/code/go/0chain.net/miner/minerGRPC"
	"google.golang.org/grpc"
)

const GRPCPort = 7031

func newMinerGRPCClient(urlRaw string) (minerGRPC.MinerClient, error) {
	u, err := url.Parse(urlRaw)
	if err != nil {
		return nil, err
	}
	host, _, _ := net.SplitHostPort(u.Host)

	cc, err := grpc.Dial(host+":"+fmt.Sprint(GRPCPort), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return minerGRPC.NewMinerClient(cc), nil
}

func WhoAmI(url string, req *minerGRPC.WhoAmIRequest) ([]byte, error) {
	minerClient, err := newMinerGRPCClient(url)
	if err != nil {
		return nil, err
	}

	whoAmIResponse, err := minerClient.WhoAmI(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return []byte(whoAmIResponse.Data), nil
}

func GetLatestFinalizedBlockSummary(url string, req *minerGRPC.GetLatestFinalizedBlockSummaryRequest) ([]byte, error) {
	minerClient, err := newMinerGRPCClient(url)
	if err != nil {
		return nil, err
	}

	getLatestFinalizedBlockSummaryResponse, err := minerClient.GetLatestFinalizedBlockSummary(context.Background(), req)
	if err != nil {
		return nil, err
	}

	summary := chain.BlockSummaryGRPCToBlockSummary(getLatestFinalizedBlockSummaryResponse.BlockSummary)

	return json.Marshal(summary)
}
