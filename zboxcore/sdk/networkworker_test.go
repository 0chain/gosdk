package sdk

import (
	"context"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/sdk/mocks"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestUpdateNetworkDetailsWorker(t *testing.T) {
	t.Run("Test_Cover_Context_Canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		go UpdateNetworkDetailsWorker(ctx)
		cancel()
	})
}

func TestUpdateNetworkDetails(t *testing.T) {
	curMiner := blockchain.GetMiners()
	curSharder := blockchain.GetMiners()
	curBlockWorker := blockchain.GetBlockWorker()
	blockWorkerMock, c := mocks.NewBlockWorkerHTTPServer(t, []string{"https://miner-0"}, []string{"https://sharder-0"})
	defer func() {
		blockchain.SetMiners(curMiner)
		blockchain.SetSharders(curSharder)
		blockchain.SetBlockWorker(curBlockWorker)
		c()
	}()
	tests := []struct {
		name        string
		blockWorker string
		wantMiners  []string
		wantSharder []string
		wantError   bool
	}{
		{
			"Test_Error_Get_Block_Worker_Failed",
			"some-wrong-block-worker" + string([]byte{0x7f, 0, 0}),
			curMiner,
			curSharder,
			true,
		},
		{
			"Test_Update_Required_Success",
			blockWorkerMock,
			[]string{"https://miner-0"},
			[]string{"https://sharder-0"},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := assert.New(t)
			blockchain.SetBlockWorker(tt.blockWorker)
			err := UpdateNetworkDetails()
			if tt.wantError {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error: %v", err)
			miners, sharders := blockchain.GetMiners(), blockchain.GetSharders()
			assertion.EqualValues(tt.wantMiners, miners)
			assertion.EqualValues(tt.wantSharder, sharders)
		})
	}
}

func TestUpdateRequired(t *testing.T) {
	curMiner := blockchain.GetMiners()
	curSharder := blockchain.GetMiners()
	type args struct {
		networkDetails *Network
	}
	tests := []struct {
		name           string
		args           args
		additionalMock func() (teardown func())
		want           bool
	}{
		{
			"Test_Required_Update_With_Empty_Loaded_Miners_Or_Sharders",
			args{&Network{Miners: []string{"some_miner"}, Sharders: []string{"some_sharder"}}},
			func() (teardown func()) { return func() {} },
			true,
		},
		{
			"Test_Required_Update",
			args{&Network{Miners: []string{"some_miner"}, Sharders: []string{"some_sharder"}}},
			func() (teardown func()) {
				blockchain.SetMiners([]string{"some_miner1"})
				blockchain.SetSharders([]string{"some_sharder1"})
				return func() {
					blockchain.SetMiners(curMiner)
					blockchain.SetSharders(curSharder)
				}
			},
			true,
		},
		{
			"Test_Non_Required_Update",
			args{&Network{Miners: []string{"some_miner"}, Sharders: []string{"some_sharder"}}},
			func() (teardown func()) {
				blockchain.SetMiners([]string{"some_miner"})
				blockchain.SetSharders([]string{"some_sharder"})
				return func() {
					blockchain.SetMiners(curMiner)
					blockchain.SetSharders(curSharder)
				}
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.additionalMock != nil {
				if teardown := tt.additionalMock(); teardown != nil {
					defer teardown()
				}
			}
			assert.Equal(t, tt.want, UpdateRequired(tt.args.networkDetails))
		})
	}
}

func TestGetNetworkDetails(t *testing.T) {
	var curBlockWorker = blockchain.GetBlockWorker()
	tests := []struct {
		name           string
		additionalMock func(t *testing.T) (teardown func())
		want           *Network
		wantErr        bool
	}{
		{
			"Test_Error_New_HTTP_Request_Failed",
			func(t *testing.T) (teardown func()) {
				blockchain.SetBlockWorker(string([]byte{0x7f, 0, 0}))
				return func() {
					blockchain.SetBlockWorker(curBlockWorker)
				}
			},
			nil,
			true,
		},
		{
			"Test_Error_JSON_Unmarshal_Reaponse_Body_Failed",
			func(t *testing.T) (teardown func()) {
				url, cl, _ := mocks.NewHTTPServer(t, map[string]http.Handler{
					"/network": http.HandlerFunc(func(w http.ResponseWriter, t *http.Request) {
						w.WriteHeader(200)
						w.Write([]byte("this is not json format"))
					}),
				})
				blockchain.SetBlockWorker(url)
				return func() {
					blockchain.SetBlockWorker(curBlockWorker)
					cl()
				}
			},
			nil,
			true,
		},
		{
			"Test_Error_Response_Status_Failed",
			func(t *testing.T) (teardown func()) {
				url, cl, _ := mocks.NewHTTPServer(t, map[string]http.Handler{
					"/network": http.HandlerFunc(func(w http.ResponseWriter, t *http.Request) {
						w.WriteHeader(400)
					}),
				})
				blockchain.SetBlockWorker(url)
				return func() {
					blockchain.SetBlockWorker(curBlockWorker)
					cl()
				}
			},
			nil,
			true,
		},
		{
			"Test_Error_Response_Status_Failed",
			func(t *testing.T) (teardown func()) {
				url, cl := mocks.NewBlockWorkerHTTPServer(t, []string{"https://miner_0"}, []string{"https://sharder_0"})
				blockchain.SetBlockWorker(url)
				return func() {
					blockchain.SetBlockWorker(curBlockWorker)
					cl()
				}
			},
			&Network{
				Miners:   []string{"https://miner_0"},
				Sharders: []string{"https://sharder_0"},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := assert.New(t)
			if tt.additionalMock != nil {
				if teardown := tt.additionalMock(t); teardown != nil {
					defer teardown()
				}
			}
			got, err := GetNetworkDetails()
			if tt.wantErr {
				assertion.Error(err, "expected error != nil")
				return
			}
			assertion.NoErrorf(err, "unexpected error: %v", err)
			assertion.EqualValues(tt.want, got)
		})
	}
}
