package sample

import (
	"fmt"
	"sync"
	"testing"

	"github.com/0chain/gosdk/zcnbridge/wallet"
	"github.com/0chain/gosdk/zcncore"
)

type BlobberAggregate struct {
	BlobberID string `json:"blobber_id" gorm:"index:idx_blobber_aggregate,unique"`
	Round     int64  `json:"round" gorm:"index:idx_blobber_aggregate,unique"`

	WritePrice  uint64 `json:"write_price"`
	Capacity    int64  `json:"capacity"`  // total blobber capacity
	Allocated   int64  `json:"allocated"` // allocated capacity
	SavedData   int64  `json:"saved_data"`
	ReadData    int64  `json:"read_data"`
	OffersTotal uint64 `json:"offers_total"`
	TotalStake  uint64 `json:"total_stake"`

	TotalServiceCharge  uint64  `json:"total_service_charge"`
	ChallengesPassed    uint64  `json:"challenges_passed"`
	ChallengesCompleted uint64  `json:"challenges_completed"`
	OpenChallenges      uint64  `json:"open_challenges"`
	InactiveRounds      int64   `json:"InactiveRounds"`
	RankMetric          float64 `json:"rank_metric" gorm:"index:idx_ba_rankmetric"`
}

const ChainConfig = `
	 {
        "chain_id":"0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe",
		"signature_scheme" : "bls0chain",
		"block_worker" : "http://dev.zus.network/dns",
		"min_submit" : 50,
		"min_confirmation" : 50,
		"confirmation_chain_length" : 3,
		"num_keys" : 1,
		"eth_node" : "https://ropsten.infura.io/v3/xxxxxxxxxxxxxxx"
	 }
`

func TestGetAggregates(t *testing.T) {
	t.Skip("learning test")
	err := zcncore.Init(ChainConfig)
	if err != nil {
		fmt.Println("Init failed")
		return
	}

	var w = `
	{
		"client_id":"0bc96a0980170045863d826f9eb579d8144013210602e88426408e9f83c236f6",
	"client_key":"a4e58c66b072d27288b650db9a476fe66a1a4f69e0f8fb11499f9ec3a579e21e5dc0298b8c5ae5baa205730d06bc04b07a31943ab3bd620e8427c15d5c413b9e",
	"keys":[
		{
			"public_key":"a4e58c66b072d27288b650db9a476fe66a1a4f69e0f8fb11499f9ec3a579e21e5dc0298b8c5ae5baa205730d06bc04b07a31943ab3bd620e8427c15d5c413b9e",
			"private_key":"c0f3a3100241888ea9c2cc5c7300e3e510a8e7190c2c20b03f80e3937a91530d"
		}],
	"mnemonics":"snake mixed bird cream cotton trouble small fee finger catalog measure spoon private second canal pact unable close predict dream mask delay path inflict",
	"version":"1.0",
	"date_created":"2019-06-19 13:37:50.466889 -0700 PDT m=+0.023873276"
	}`

	err = zcncore.SetWalletInfo(w, false)
	if err != nil {
		fmt.Println("set wallet info failed: ", err)
		return
	}

	wg := &sync.WaitGroup{}
	var snaps []BlobberAggregate
	statusBar := wallet.NewZCNStatus(&snaps)
	statusBar.Wg = wg
	wg.Add(5)
	err = zcncore.GetBlobberSnapshots(int64(587), 20, 0, statusBar)
	if err != nil {
		t.Error(err)
		return
	}

	err = zcncore.GetSharderSnapshots(int64(587), 20, 0, statusBar)
	if err != nil {
		t.Error(err)
		return
	}

	err = zcncore.GetMinerSnapshots(int64(587), 20, 0, statusBar)
	if err != nil {
		t.Error(err)
		return
	}

	err = zcncore.GetAuthorizerSnapshots(int64(587), 20, 0, statusBar)
	if err != nil {
		t.Error(err)
		return
	}

	err = zcncore.GetValidatorSnapshots(int64(587), 20, 0, statusBar)
	if err != nil {
		t.Error(err)
		return
	}

	err = zcncore.GetUserSnapshots(int64(587), 20, 0, statusBar)
	if err != nil {
		t.Error(err)
		return
	}

	wg.Wait()
	if !statusBar.Success {
		t.Error(statusBar.Err)
	}
}
