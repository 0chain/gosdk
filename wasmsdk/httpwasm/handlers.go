package httpwasm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zcncore"
)

var signPrivatekey = `18c09c2639d7c8b3f26b273cdbfddf330c4f86c2ac3030a6b9a8533dc0c91f5e`

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// Miners API
func smartContractTxnValue(w http.ResponseWriter, req *http.Request) {
	txn := &Transaction{}
	respondWithJSON(w, http.StatusOK, txn)
}

func createWallet(w http.ResponseWriter, req *http.Request) {
	var walletMockConfig struct {
		ClientID  string `json:"id"`
		ClientKey string `json:"public_key"`
	}

	err := json.NewDecoder(req.Body).Decode(&walletMockConfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := &util.PostResponse{
		Url:        "",
		StatusCode: http.StatusOK,
		Status:     "1",
		Body:       "",
	}

	respondWithJSON(w, http.StatusOK, result)
}

func getClientDetail(w http.ResponseWriter, req *http.Request) {
	v := req.URL.Query()
	clientID := v.Get("id")

	response := zcncore.GetClientResponse{
		ID:           "9a566aa4f8e8c342fed97c8928040a21f21b8f574e5782c28568635ba9c75a85",
		Version:      "1.0",
		CreationDate: 1621618322,
		PublicKey:    "041eeb1b4eb9b2456799d8e2a566877e83bc5d76ff38b964bd4b7796f6a6ccae6f1966a4d91d362669fafa3d95526b132a6341e3dfff6447e0e76a07b3a7cfa6e8034574266b382b8e5174477ab8a32a49a57eda74895578031cd2d41fd0aef446046d6e633f5eb68a93013dfac1420bf7a1e1bf7a87476024478e97a1cc115de9",
	}

	if clientID == "9a566aa4f8e8c342fed97c8928040a21f21b8f574e5782c28568635ba9c75a85" {
		respondWithJSON(w, http.StatusOK, response)
	}
}

// Sharders API
func verifyTransaction(w http.ResponseWriter, req *http.Request) {
	v := req.URL.Query()
	hash := v.Get("hash")

	signScheme := zcncrypto.NewSignatureScheme("bls0chain")
	signScheme.SetPrivateKey(signPrivatekey)
	scheme, _ := signScheme.Sign(hash)

	spuu := sdk.StakePoolUnlockUnstake{
		Unstake: common.Timestamp(1641016719),
	}

	txnOutput, _ := json.Marshal(spuu)

	txn := &Transaction{}
	txn.Hash = hash
	txn.ChainID = blockchain.GetChainID()
	txn.ClientID = client.GetClientID()
	txn.Signature = scheme
	txn.TransactionOutput = string(txnOutput)
	txn.Value = 500

	var objmap = map[string]Transaction{
		"txn": *txn,
	}

	respondWithJSON(w, http.StatusOK, objmap)
}

func getPoolStat(w http.ResponseWriter, req *http.Request) {
	v := req.URL.Query()
	client_id := v.Get("client_id")

	poolStat := sdk.AllocationPoolStats{
		Pools: []*sdk.AllocationPoolStat{
			{
				ID:           client_id,
				Balance:      1000,
				AllocationID: common.Key(GetMockAllocationId(10)),
			},
		},
		Back: &sdk.BackPool{
			ID:      client_id,
			Balance: 150,
		},
	}

	respondWithJSON(w, http.StatusOK, poolStat)
}

func getStakePoolStat(w http.ResponseWriter, req *http.Request) {
	v := req.URL.Query()
	blobber_id := v.Get("blobber_id")

	poolStat := sdk.StakePoolInfo{
		ID: common.Key(blobber_id),
	}

	respondWithJSON(w, http.StatusOK, poolStat)
}

func getUserStakePoolStat(w http.ResponseWriter, req *http.Request) {
	v := req.URL.Query()
	client_id := v.Get("client_id")

	poolStat := sdk.StakePoolUserInfo{
		Pools: map[common.Key][]*sdk.StakePoolDelegatePoolInfo{
			common.Key(client_id): {
				{
					ID: common.Key(client_id),
				},
			},
		},
	}

	respondWithJSON(w, http.StatusOK, poolStat)
}

func getChallengePoolStat(w http.ResponseWriter, req *http.Request) {
	v := req.URL.Query()
	allocation_id := v.Get("allocation_id")

	poolStat := sdk.ChallengePoolInfo{
		ID:         allocation_id,
		Balance:    common.Balance(2000),
		StartTime:  common.Timestamp(1633878133),
		Expiration: common.Timestamp(1641016719),
		Finalized:  true,
	}

	respondWithJSON(w, http.StatusOK, poolStat)
}

//
// storage SC configurations and blobbers
//

type StorageStakePoolConfig struct {
	MinLock          common.Balance `json:"min_lock"`
	InterestRate     float64        `json:"interest_rate"`
	InterestInterval time.Duration  `json:"interest_interval"`
}

// read pool configs

type StorageReadPoolConfig struct {
	MinLock       common.Balance `json:"min_lock"`
	MinLockPeriod time.Duration  `json:"min_lock_period"`
	MaxLockPeriod time.Duration  `json:"max_lock_period"`
}

// write pool configurations

type StorageWritePoolConfig struct {
	MinLock       common.Balance `json:"min_lock"`
	MinLockPeriod time.Duration  `json:"min_lock_period"`
	MaxLockPeriod time.Duration  `json:"max_lock_period"`
}

func getConfig(w http.ResponseWriter, req *http.Request) {

	readPool := StorageReadPoolConfig{
		MinLock:       common.Balance(500),
		MinLockPeriod: time.Duration(time.Hour),
		MaxLockPeriod: time.Duration(time.Hour * 12),
	}
	writePool := StorageWritePoolConfig{
		MinLock:       common.Balance(1500),
		MinLockPeriod: time.Duration(time.Hour),
		MaxLockPeriod: time.Duration(time.Hour * 12),
	}
	stakePool := StorageStakePoolConfig{
		MinLock:          common.Balance(2000),
		InterestRate:     float64(0.25),
		InterestInterval: time.Duration(time.Hour * 720),
	}

	conf := sdk.InputMap{
		Fields: map[string]interface{}{
			"readpool":  readPool,
			"writepool": writePool,
			"stakepool": stakePool,
		},
	}

	respondWithJSON(w, http.StatusOK, conf)
}

func getBlobbers(w http.ResponseWriter, req *http.Request) {
	type nodes struct {
		Nodes []*sdk.Blobber
	}

	objmap := &nodes{
		Nodes: []*sdk.Blobber{
			{
				ID:      "0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe",
				BaseURL: "http://pedro.alphanet-0chain.net:5051",
			},
			{
				ID:      "d374fc5b55a496d26e9d642ed0708746fd64a24bd59139dacf50b4a4ec4c9b51",
				BaseURL: "http://pedro.alphanet-0chain.net:5052",
			},
		},
	}

	respondWithJSON(w, http.StatusOK, objmap)
}

func getBlobber(w http.ResponseWriter, req *http.Request) {
	v := req.URL.Query()
	blobber_id := v.Get("blobber_id")
	objmap := &sdk.Blobber{
		ID:              common.Key(blobber_id),
		Capacity:        common.Size(1000000),
		Used:            common.Size(500000),
		LastHealthCheck: common.Timestamp(1633878133),
	}

	respondWithJSON(w, http.StatusOK, objmap)
}

func allocation(w http.ResponseWriter, req *http.Request) {
	v := req.URL.Query()
	allocation_id := v.Get("allocation")
	objmap := &sdk.Allocation{
		ID: allocation_id,
	}

	respondWithJSON(w, http.StatusOK, objmap)
}

func allocations(w http.ResponseWriter, req *http.Request) {
	v := req.URL.Query()
	allocation_id := v.Get("client")
	objmap := []*sdk.Allocation{
		{ID: allocation_id},
	}

	respondWithJSON(w, http.StatusOK, objmap)
}

func allocationMinLock(w http.ResponseWriter, req *http.Request) {
	v := req.URL.Query()
	allocation_data := v.Get("allocation_data")

	var mockResponse struct {
		DataShards        int            `json:"data_shards"`
		ParityShards      int            `json:"parity_shards"`
		Size              int64          `json:"size"`
		OwnerID           string         `json:"owner_id"`
		PublicKey         string         `json:"owner_public_key"`
		Expiry            int64          `json:"expiration_date"`
		PreferredBlobbers []string       `json:"preferred_blobbers"`
		ReadPrice         sdk.PriceRange `json:"read_price_range"`
		WritePrice        sdk.PriceRange `json:"write_price_range"`
		Mcct              time.Duration  `json:"max_challenge_completion_time"`
	}
	err := json.Unmarshal([]byte(allocation_data), &mockResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var response = make(map[string]int64)
	response["min_lock_demand"] = mockResponse.Expiry

	respondWithJSON(w, http.StatusOK, response)
}

func sharderGetBalance(w http.ResponseWriter, req *http.Request) {
	v := req.URL.Query()
	client_id := v.Get("client_id")

	var response = make(map[string]int64)

	fmt.Println("client_id:" + client_id)
	response["balance"] = 1000

	respondWithJSON(w, http.StatusOK, response)
}

func vestingPoolInfo(w http.ResponseWriter, req *http.Request) {
	v := req.URL.Query()
	pool_id := v.Get("pool_id")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pool_id))
}

func getClientPools(w http.ResponseWriter, req *http.Request) {
	v := req.URL.Query()
	client_id := v.Get("client_id")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(client_id))
}

func getVestingSCConfig(w http.ResponseWriter, req *http.Request) {
	scconfig := zcncore.VestingSCConfig{
		MinLock:              common.Balance(2000),
		MinDuration:          time.Duration(time.Hour),
		MaxDuration:          time.Duration(time.Hour * 48),
		MaxDestinations:      20,
		MaxDescriptionLength: 100,
	}

	respondWithJSON(w, http.StatusOK, scconfig)
}

func getMinerList(w http.ResponseWriter, req *http.Request) {
	minerArray := []string{"127.0.0.1:1/miner01", "127.0.0.1:1/miner02"}
	respondWithJSON(w, http.StatusOK, minerArray)
}

func getSharderList(w http.ResponseWriter, req *http.Request) {
	minerArray := []string{"127.0.0.1:1/sharder01", "127.0.0.1:1/sharder02"}
	respondWithJSON(w, http.StatusOK, minerArray)
}

func getNodeStat(w http.ResponseWriter, req *http.Request) {
	v := req.URL.Query()
	id := v.Get("id")

	var scn zcncore.MinerSCNodes
	scn.Nodes = append(scn.Nodes, zcncore.Node{Miner: zcncore.Miner{ID: GetMockId(1)}})
	scn.Nodes = append(scn.Nodes, zcncore.Node{Miner: zcncore.Miner{ID: id}})
	scn.Nodes = append(scn.Nodes, zcncore.Node{Miner: zcncore.Miner{ID: GetMockId(1000)}})

	respondWithJSON(w, http.StatusOK, scn)
}

// Server API
func getNetwork(w http.ResponseWriter, req *http.Request) {
	n := zcncore.Network{Miners: blockchain.GetMiners(), Sharders: blockchain.GetSharders()}
	respondWithJSON(w, http.StatusOK, n)
}

func commitToFabric(w http.ResponseWriter, req *http.Request) {
	var fabricMockConfig struct {
		Channel          string   `json:"channel"`
		ChaincodeName    string   `json:"chaincode_name"`
		ChaincodeVersion string   `json:"chaincode_version"`
		Method           string   `json:"method"`
		Args             []string `json:"args"`
	}

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(req.Body).Decode(&fabricMockConfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, pass, _ := req.BasicAuth()
	if user == "TEST" && pass == "TEST" {
		respondWithJSON(w, http.StatusOK, fabricMockConfig)
	}
}

func setupAuthHost(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ID,200,OK,9a566aa4f8e8c342fed97c8928040a21f21b8f574e5782c28568635ba9c75a85"))
}
