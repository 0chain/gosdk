package sdk

import (
	"encoding/json"
	"fmt"
	"github.com/0chain/gosdk/core/common"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/0chain/gosdk/core/version"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/sdk/mocks"
	"github.com/stretchr/testify/assert"
)

func init() {
	Logger.SetLevel(0)
}

func TestGetVersion(t *testing.T) {
	assert.Equal(t, version.VERSIONSTR, GetVersion())
}

func TestSetLogLevel(t *testing.T) {
	// statement cover
	SetLogLevel(0)
}

func TestSetLogFile(t *testing.T) {
	var logFile = "test.log"
	defer func() { _ = os.Remove(logFile) }()

	SetLogFile(logFile, true)

	f, err := os.Stat(logFile)
	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
	assert.False(t, f.IsDir())
}

func TestGetNetwork(t *testing.T) {
	curMiners := blockchain.GetMiners()
	defer blockchain.SetMiners(curMiners)
	curSharders := blockchain.GetSharders()
	defer blockchain.SetSharders(curSharders)

	miners := []string{"https://miner_0"}
	sharders := []string{"https://sharder_0"}
	blockchain.SetMiners(miners)
	blockchain.SetSharders(sharders)

	network := GetNetwork()

	assert.Equal(t, miners, network.Miners)
	assert.Equal(t, sharders, network.Sharders)
}

func TestSetMaxTxnQuery(t *testing.T) {
	curTxnQuery := blockchain.GetMaxTxnQuery()
	defer blockchain.SetMaxTxnQuery(curTxnQuery)

	txnQuery := rand.Int()

	SetMaxTxnQuery(txnQuery)

	assert.Equal(t, txnQuery, blockchain.GetMaxTxnQuery())
}

func TestSetQuerySleepTime(t *testing.T) {
	curQuerySleepTime := blockchain.GetQuerySleepTime()
	defer blockchain.SetQuerySleepTime(curQuerySleepTime)

	txnQuerySleepTime := rand.Int()

	SetQuerySleepTime(txnQuerySleepTime)

	assert.Equal(t, txnQuerySleepTime, blockchain.GetQuerySleepTime())
}

func TestSetMinSubmit(t *testing.T) {
	curMinSubmit := blockchain.GetMinSubmit()
	defer blockchain.SetMinSubmit(curMinSubmit)

	txnMinSubmit := rand.Int()

	SetMinSubmit(txnMinSubmit)

	assert.Equal(t, txnMinSubmit, blockchain.GetMinSubmit())
}

func TestSetMinConfirmation(t *testing.T) {
	curMinConfirmation := blockchain.GetMinConfirmation()
	defer blockchain.SetMinConfirmation(curMinConfirmation)

	txnMinConfirmation := rand.Int()

	SetMinConfirmation(txnMinConfirmation)

	assert.Equal(t, txnMinConfirmation, blockchain.GetMinConfirmation())
}

func TestSetNetwork(t *testing.T) {
	curMiners := blockchain.GetMiners()
	defer blockchain.SetMiners(curMiners)
	curSharders := blockchain.GetSharders()
	defer blockchain.SetSharders(curSharders)

	miners := []string{"https://miner_0"}
	sharders := []string{"https://sharder_0"}

	SetNetwork(miners, sharders)

	assert.Equal(t, miners, blockchain.GetMiners())
	assert.Equal(t, sharders, blockchain.GetSharders())
}

func TestCreateReadPool(t *testing.T) {
	curQuerySleepTime := blockchain.GetQuerySleepTime()
	defer blockchain.SetQuerySleepTime(curQuerySleepTime)
	curMaxTxnQuery := blockchain.GetMaxTxnQuery()
	defer blockchain.SetMaxTxnQuery(curMaxTxnQuery)

	_, _, _, cncl := setupMockInitStorageSDK(t, configDir, 0)
	defer cncl()
	mocks.SetMinerHandler(t, "/v1/transaction/put", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	mocks.SetSharderHandler(t, "/v1/transaction/get/confirmation", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"txn":{"hash":"1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23","version":"1.0","client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","public_key":"eeb0c33325cbee0fb58bc09962f69a44d0b22ac2824a063eb1002273347e601a4612e6fea7e1a1ae62e0e3b7f1301c4de8a855bae86ebfa6e9dbbb41c3e39c24","transaction_data":"{\"OpType\":\"Move\",\"PreValue\":\"/1.txt\",\"CurrValue\":\"/d/1.txt\"}","transaction_value":0,"signature":"98427e25b635b8d88881ddc9a1f84db0951f145ffa90462c0290ed84563bdc92","creation_date":1617159987,"transaction_type":10,"transaction_fee":0,"txn_output_hash":""},"block_hash":"4dd9de1f3724a688686a5b54879cf424a9f8e6cb56ab77bfd19586dfcc804ba8"}`))
	})
	blockchain.SetQuerySleepTime(1)
	blockchain.SetMaxTxnQuery(3)

	err := CreateReadPool()

	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
}

func TestAllocationPoolStats_AllocFilter(t *testing.T) {
	allocationPoolStats0 := []*AllocationPoolStat{{AllocationID: "allocation_0", ID: "00"}, {AllocationID: "allocation_1", ID: "10"}, {AllocationID: "allocation_0", ID: "01"}}
	allocationPoolStats1 := []*AllocationPoolStat{{AllocationID: "allocation_0", ID: "00"}, {AllocationID: "allocation_0", ID: "01"}}
	aps := &AllocationPoolStats{
		Pools: allocationPoolStats0,
		Back:  nil,
	}

	t.Run("Test_Empty_AllocID_Argument", func(t *testing.T) {
		aps.AllocFilter("")
		assert.Equal(t, allocationPoolStats0, aps.Pools)
	})

	t.Run("Test_With_Non_Empty_AllocID_Argument", func(t *testing.T) {
		aps.AllocFilter("allocation_0")
		assert.Equal(t, allocationPoolStats1, aps.Pools)
	})
}

func TestGetReadPoolInfo(t *testing.T) {
	infoStr := `{"pools":[{"id":"1bf4729c2ae3c950646bb5b79eebe290231906bdc4285471219046b38eb780df","balance":5000000000,"expire_at":1620643080,"allocation_id":"a324ee3369adfeedbd767ae9932970c67d52bf2018cd7b8ebcfe550dea6c6961","blobbers":[{"blobber_id":"63230aeda3360ae6540e8604db40333638d963edac8d92fb9b576915155d5dc6","balance":1250000000},{"blobber_id":"749154fee3f95528d96e4651413a56f0be8ceb2e9daec8d11fad93efa0082b77","balance":1250000000},{"blobber_id":"833bc63ba483a9c4e557f609f2afb66bc0983d4067999e7d93843b4111ed5507","balance":1250000000},{"blobber_id":"a983087814ebd0bf267449dc179fe1790c18eb305ab888807196741aa9adbc97","balance":1250000000}],"locked":true},{"id":"677f02013f2963803017b31cb8045cba55d5d46624f4ed8defe9d28ae69a7037","balance":5000000000,"expire_at":1620642882,"allocation_id":"b78e4a211213a0296ddf9cd06133c8f5dca9dd256b5694e3fa824b7d33a355ef","blobbers":[{"blobber_id":"4e729c0ef0b8177df1c130dde2988d4cb7456afa183bf93cbf3bd74120eeebaf","balance":833333333},{"blobber_id":"63230aeda3360ae6540e8604db40333638d963edac8d92fb9b576915155d5dc6","balance":833333333},{"blobber_id":"68d17529d461a1257cbca5058fd92a28424bf783660bace37f362e5cb02520d5","balance":833333333},{"blobber_id":"749154fee3f95528d96e4651413a56f0be8ceb2e9daec8d11fad93efa0082b77","balance":833333333},{"blobber_id":"833bc63ba483a9c4e557f609f2afb66bc0983d4067999e7d93843b4111ed5507","balance":833333333},{"blobber_id":"a983087814ebd0bf267449dc179fe1790c18eb305ab888807196741aa9adbc97","balance":833333333}],"locked":true},{"id":"366669cdd10bc977895f27ae94fb8b1fc9c1e81026c8378a508cc8f0167df9e7","balance":4999023444,"expire_at":1620643139,"allocation_id":"f5bb9b7449acc7628eb6a1fbcb35cc8ca49b2f4c412c16586ec8eb9098fa8d11","blobbers":[{"blobber_id":"4e729c0ef0b8177df1c130dde2988d4cb7456afa183bf93cbf3bd74120eeebaf","balance":1249755860},{"blobber_id":"63230aeda3360ae6540e8604db40333638d963edac8d92fb9b576915155d5dc6","balance":1249755860},{"blobber_id":"749154fee3f95528d96e4651413a56f0be8ceb2e9daec8d11fad93efa0082b77","balance":1249755860},{"blobber_id":"a983087814ebd0bf267449dc179fe1790c18eb305ab888807196741aa9adbc97","balance":1249755860}],"locked":true}]}`
	var expectedInfo *AllocationPoolStats
	_ = json.Unmarshal([]byte(infoStr), &expectedInfo)

	_, _, _, cncl := setupMockInitStorageSDK(t, configDir, 0)
	defer cncl()
	mocks.SetSharderHandler(t, fmt.Sprintf("%v%v%v", "/v1/screst/", STORAGE_SCADDRESS, "/getReadPoolStat"), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(infoStr))
	})

	info, err := GetReadPoolInfo("")

	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
	assert.Equal(t, expectedInfo, info)
}

func TestReadPoolLock(t *testing.T) {
	curQuerySleepTime := blockchain.GetQuerySleepTime()
	defer blockchain.SetQuerySleepTime(curQuerySleepTime)
	curMaxTxnQuery := blockchain.GetMaxTxnQuery()
	defer blockchain.SetMaxTxnQuery(curMaxTxnQuery)

	_, _, _, cncl := setupMockInitStorageSDK(t, configDir, 0)
	defer cncl()
	mocks.SetMinerHandler(t, "/v1/transaction/put", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	mocks.SetSharderHandler(t, "/v1/transaction/get/confirmation", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"txn":{"hash":"1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23","version":"1.0","client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","public_key":"eeb0c33325cbee0fb58bc09962f69a44d0b22ac2824a063eb1002273347e601a4612e6fea7e1a1ae62e0e3b7f1301c4de8a855bae86ebfa6e9dbbb41c3e39c24","transaction_data":"{\"OpType\":\"Move\",\"PreValue\":\"/1.txt\",\"CurrValue\":\"/d/1.txt\"}","transaction_value":0,"signature":"98427e25b635b8d88881ddc9a1f84db0951f145ffa90462c0290ed84563bdc92","creation_date":1617159987,"transaction_type":10,"transaction_fee":0,"txn_output_hash":""},"block_hash":"4dd9de1f3724a688686a5b54879cf424a9f8e6cb56ab77bfd19586dfcc804ba8"}`))
	})
	blockchain.SetQuerySleepTime(1)
	blockchain.SetMaxTxnQuery(3)

	err := ReadPoolLock(time.Hour*720, "a324ee3369adfeedbd767ae9932970c67d52bf2018cd7b8ebcfe550dea6c6961", "", 500000000, 0)

	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
}

func TestReadPoolUnlock(t *testing.T) {
	curQuerySleepTime := blockchain.GetQuerySleepTime()
	defer blockchain.SetQuerySleepTime(curQuerySleepTime)
	curMaxTxnQuery := blockchain.GetMaxTxnQuery()
	defer blockchain.SetMaxTxnQuery(curMaxTxnQuery)

	_, _, _, cncl := setupMockInitStorageSDK(t, configDir, 0)
	defer cncl()
	mocks.SetMinerHandler(t, "/v1/transaction/put", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	mocks.SetSharderHandler(t, "/v1/transaction/get/confirmation", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"txn":{"hash":"1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23","version":"1.0","client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","public_key":"eeb0c33325cbee0fb58bc09962f69a44d0b22ac2824a063eb1002273347e601a4612e6fea7e1a1ae62e0e3b7f1301c4de8a855bae86ebfa6e9dbbb41c3e39c24","transaction_data":"{\"OpType\":\"Move\",\"PreValue\":\"/1.txt\",\"CurrValue\":\"/d/1.txt\"}","transaction_value":0,"signature":"98427e25b635b8d88881ddc9a1f84db0951f145ffa90462c0290ed84563bdc92","creation_date":1617159987,"transaction_type":10,"transaction_fee":0,"txn_output_hash":""},"block_hash":"4dd9de1f3724a688686a5b54879cf424a9f8e6cb56ab77bfd19586dfcc804ba8"}`))
	})
	blockchain.SetQuerySleepTime(1)
	blockchain.SetMaxTxnQuery(3)

	err := ReadPoolUnlock("366669cdd10bc977895f27ae94fb8b1fc9c1e81026c8378a508cc8f0167df9e7", 0)

	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
}

func TestGetStakePoolInfo(t *testing.T) {
	infoStr := `{"pool_id":"2b2ed3816062f47dce1b92c30ee060e833816a1e378cffca6dd18fc413baf7b6","balance":3000000000,"unstake":0,"free":50573240320,"capacity":3543348019200,"write_price":63694267,"offers":[],"offers_total":0,"delegate":[{"id":"03fa3602ca5e4be845e08adb6f35a7708d835e4a7c4f772e8e5c66ea9ae455b9","balance":3000000000,"delegate_id":"d477d12134c2d7ba5ab71ac8ad37f244224695ef3215be990c3215d531c5a329","rewards":0,"interests":0,"penalty":0,"pending_interests":0,"unstake":0}],"interests":0,"penalty":0,"rewards":{"charge":0,"blobber":0,"validator":0},"settings":{"delegate_wallet":"e73d88ba1133b8bc8e1683b8c11f30b9c172b4a067463f5657f82c0f2b77561a","min_stake":10000000000,"max_stake":1000000000000,"num_delegates":50,"service_charge":0.3}}`
	var expectedInfo *StakePoolInfo
	_ = json.Unmarshal([]byte(infoStr), &expectedInfo)

	_, _, _, cncl := setupMockInitStorageSDK(t, configDir, 0)
	defer cncl()
	mocks.SetSharderHandler(t, fmt.Sprintf("%v%v%v", "/v1/screst/", STORAGE_SCADDRESS, "/getStakePoolStat"), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(infoStr))
	})

	info, err := GetStakePoolInfo("")

	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
	assert.Equal(t, expectedInfo, info)
}

func TestGetStakePoolUserInfo(t *testing.T) {
	infoStr := `{"pools": {"0de4fd069f1304f11ce5794ae4216f92326044392e74320932f5cba67bdeff09":[{"id":"9a61e70c8f04332c35b79a69fc7e9e85b9ca32a48827f119679efd205d4d5d5f","balance":5000000000,"delegate_id":"d477d12134c2d7ba5ab71ac8ad37f244224695ef3215be990c3215d531c5a329","rewards":0,"interests":167000,"penalty":0,"pending_interests":0,"unstake":0},{"id":"a3b4317d0ca73b8a303769cb086bb769cce2300cbb99eeb08ad63799bd61c48d","balance":3000000000,"delegate_id":"d477d12134c2d7ba5ab71ac8ad37f244224695ef3215be990c3215d531c5a329","rewards":0,"interests":0,"penalty":0,"pending_interests":0,"unstake":0}],"2b2ed3816062f47dce1b92c30ee060e833816a1e378cffca6dd18fc413baf7b6":[{"id":"03fa3602ca5e4be845e08adb6f35a7708d835e4a7c4f772e8e5c66ea9ae455b9","balance":3000000000,"delegate_id":"d477d12134c2d7ba5ab71ac8ad37f244224695ef3215be990c3215d531c5a329","rewards":0,"interests":0,"penalty":0,"pending_interests":0,"unstake":0}]}}`
	var expectedInfo *StakePoolUserInfo
	_ = json.Unmarshal([]byte(infoStr), &expectedInfo)

	_, _, _, cncl := setupMockInitStorageSDK(t, configDir, 0)
	defer cncl()
	mocks.SetSharderHandler(t, fmt.Sprintf("%v%v%v", "/v1/screst/", STORAGE_SCADDRESS, "/getUserStakePoolStat"), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(infoStr))
	})

	info, err := GetStakePoolUserInfo("")

	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
	assert.Equal(t, expectedInfo, info)
}

func TestStakePoolLock(t *testing.T) {
	curQuerySleepTime := blockchain.GetQuerySleepTime()
	defer blockchain.SetQuerySleepTime(curQuerySleepTime)
	curMaxTxnQuery := blockchain.GetMaxTxnQuery()
	defer blockchain.SetMaxTxnQuery(curMaxTxnQuery)

	_, _, _, cncl := setupMockInitStorageSDK(t, configDir, 0)
	defer cncl()
	mocks.SetMinerHandler(t, "/v1/transaction/put", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	mocks.SetSharderHandler(t, "/v1/transaction/get/confirmation", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"txn":{"hash":"1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23","version":"1.0","client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","public_key":"eeb0c33325cbee0fb58bc09962f69a44d0b22ac2824a063eb1002273347e601a4612e6fea7e1a1ae62e0e3b7f1301c4de8a855bae86ebfa6e9dbbb41c3e39c24","transaction_data":"{\"OpType\":\"Move\",\"PreValue\":\"/1.txt\",\"CurrValue\":\"/d/1.txt\"}","transaction_value":0,"signature":"98427e25b635b8d88881ddc9a1f84db0951f145ffa90462c0290ed84563bdc92","creation_date":1617159987,"transaction_type":10,"transaction_fee":0,"txn_output_hash":""},"block_hash":"4dd9de1f3724a688686a5b54879cf424a9f8e6cb56ab77bfd19586dfcc804ba8"}`))
	})
	blockchain.SetQuerySleepTime(1)
	blockchain.SetMaxTxnQuery(3)

	poolID, err := StakePoolLock("", 3000000000, 0)

	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
	assert.Equal(t, poolID, "1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23")
}

func TestStakePoolUnlock(t *testing.T) {
	curQuerySleepTime := blockchain.GetQuerySleepTime()
	defer blockchain.SetQuerySleepTime(curQuerySleepTime)
	curMaxTxnQuery := blockchain.GetMaxTxnQuery()
	defer blockchain.SetMaxTxnQuery(curMaxTxnQuery)

	_, _, _, cncl := setupMockInitStorageSDK(t, configDir, 0)
	defer cncl()
	mocks.SetMinerHandler(t, "/v1/transaction/put", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	mocks.SetSharderHandler(t, "/v1/transaction/get/confirmation", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"txn":{"hash":"1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23","version":"1.0","client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","public_key":"eeb0c33325cbee0fb58bc09962f69a44d0b22ac2824a063eb1002273347e601a4612e6fea7e1a1ae62e0e3b7f1301c4de8a855bae86ebfa6e9dbbb41c3e39c24","transaction_data":"{\"OpType\":\"Move\",\"PreValue\":\"/1.txt\",\"CurrValue\":\"/d/1.txt\"}","transaction_value":0,"signature":"98427e25b635b8d88881ddc9a1f84db0951f145ffa90462c0290ed84563bdc92","creation_date":1617159987,"transaction_type":10,"transaction_fee":0,"txn_output_hash":"e73d88ba1133b8bc8e1683b8c11f30b9c172b4a067463f5657f82c0f2b77561a","transaction_output":"{\"unstake\": 1618160720}"},"block_hash":"4dd9de1f3724a688686a5b54879cf424a9f8e6cb56ab77bfd19586dfcc804ba8"}`))
	})
	blockchain.SetQuerySleepTime(1)
	blockchain.SetMaxTxnQuery(3)

	unstake, err := StakePoolUnlock("", "03fa3602ca5e4be845e08adb6f35a7708d835e4a7c4f772e8e5c66ea9ae455b9", 0)
	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
	assert.Equal(t, unstake, common.Timestamp(1618160720))
}

func TestStakePoolPayInterests(t *testing.T) {
	curQuerySleepTime := blockchain.GetQuerySleepTime()
	defer blockchain.SetQuerySleepTime(curQuerySleepTime)
	curMaxTxnQuery := blockchain.GetMaxTxnQuery()
	defer blockchain.SetMaxTxnQuery(curMaxTxnQuery)

	_, _, _, cncl := setupMockInitStorageSDK(t, configDir, 0)
	defer cncl()
	mocks.SetMinerHandler(t, "/v1/transaction/put", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	mocks.SetSharderHandler(t, "/v1/transaction/get/confirmation", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"txn":{"hash":"1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23","version":"1.0","client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","public_key":"eeb0c33325cbee0fb58bc09962f69a44d0b22ac2824a063eb1002273347e601a4612e6fea7e1a1ae62e0e3b7f1301c4de8a855bae86ebfa6e9dbbb41c3e39c24","transaction_data":"{\"OpType\":\"Move\",\"PreValue\":\"/1.txt\",\"CurrValue\":\"/d/1.txt\"}","transaction_value":0,"signature":"98427e25b635b8d88881ddc9a1f84db0951f145ffa90462c0290ed84563bdc92","creation_date":1617159987,"transaction_type":10,"transaction_fee":0,"txn_output_hash":"e73d88ba1133b8bc8e1683b8c11f30b9c172b4a067463f5657f82c0f2b77561a","transaction_output":"{\"unstake\": 1618160720}"},"block_hash":"4dd9de1f3724a688686a5b54879cf424a9f8e6cb56ab77bfd19586dfcc804ba8"}`))
	})
	blockchain.SetQuerySleepTime(1)
	blockchain.SetMaxTxnQuery(3)

	err := StakePoolPayInterests("")
	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
}

func TestGetWritePoolInfo(t *testing.T) {
	infoStr := `{"pools":[{"id":"1bf4729c2ae3c950646bb5b79eebe290231906bdc4285471219046b38eb780df","balance":5000000000,"expire_at":1620643080,"allocation_id":"a324ee3369adfeedbd767ae9932970c67d52bf2018cd7b8ebcfe550dea6c6961","blobbers":[{"blobber_id":"63230aeda3360ae6540e8604db40333638d963edac8d92fb9b576915155d5dc6","balance":1250000000},{"blobber_id":"749154fee3f95528d96e4651413a56f0be8ceb2e9daec8d11fad93efa0082b77","balance":1250000000},{"blobber_id":"833bc63ba483a9c4e557f609f2afb66bc0983d4067999e7d93843b4111ed5507","balance":1250000000},{"blobber_id":"a983087814ebd0bf267449dc179fe1790c18eb305ab888807196741aa9adbc97","balance":1250000000}],"locked":true},{"id":"677f02013f2963803017b31cb8045cba55d5d46624f4ed8defe9d28ae69a7037","balance":5000000000,"expire_at":1620642882,"allocation_id":"b78e4a211213a0296ddf9cd06133c8f5dca9dd256b5694e3fa824b7d33a355ef","blobbers":[{"blobber_id":"4e729c0ef0b8177df1c130dde2988d4cb7456afa183bf93cbf3bd74120eeebaf","balance":833333333},{"blobber_id":"63230aeda3360ae6540e8604db40333638d963edac8d92fb9b576915155d5dc6","balance":833333333},{"blobber_id":"68d17529d461a1257cbca5058fd92a28424bf783660bace37f362e5cb02520d5","balance":833333333},{"blobber_id":"749154fee3f95528d96e4651413a56f0be8ceb2e9daec8d11fad93efa0082b77","balance":833333333},{"blobber_id":"833bc63ba483a9c4e557f609f2afb66bc0983d4067999e7d93843b4111ed5507","balance":833333333},{"blobber_id":"a983087814ebd0bf267449dc179fe1790c18eb305ab888807196741aa9adbc97","balance":833333333}],"locked":true},{"id":"366669cdd10bc977895f27ae94fb8b1fc9c1e81026c8378a508cc8f0167df9e7","balance":4999023444,"expire_at":1620643139,"allocation_id":"f5bb9b7449acc7628eb6a1fbcb35cc8ca49b2f4c412c16586ec8eb9098fa8d11","blobbers":[{"blobber_id":"4e729c0ef0b8177df1c130dde2988d4cb7456afa183bf93cbf3bd74120eeebaf","balance":1249755860},{"blobber_id":"63230aeda3360ae6540e8604db40333638d963edac8d92fb9b576915155d5dc6","balance":1249755860},{"blobber_id":"749154fee3f95528d96e4651413a56f0be8ceb2e9daec8d11fad93efa0082b77","balance":1249755860},{"blobber_id":"a983087814ebd0bf267449dc179fe1790c18eb305ab888807196741aa9adbc97","balance":1249755860}],"locked":true}]}`
	var expectedInfo *AllocationPoolStats
	_ = json.Unmarshal([]byte(infoStr), &expectedInfo)

	_, _, _, cncl := setupMockInitStorageSDK(t, configDir, 0)
	defer cncl()
	mocks.SetSharderHandler(t, fmt.Sprintf("%v%v%v", "/v1/screst/", STORAGE_SCADDRESS, "/getWritePoolStat"), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(infoStr))
	})

	info, err := GetWritePoolInfo("")

	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
	assert.Equal(t, expectedInfo, info)
}

func TestWritePoolLock(t *testing.T) {
	curQuerySleepTime := blockchain.GetQuerySleepTime()
	defer blockchain.SetQuerySleepTime(curQuerySleepTime)
	curMaxTxnQuery := blockchain.GetMaxTxnQuery()
	defer blockchain.SetMaxTxnQuery(curMaxTxnQuery)

	_, _, _, cncl := setupMockInitStorageSDK(t, configDir, 0)
	defer cncl()
	mocks.SetMinerHandler(t, "/v1/transaction/put", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	mocks.SetSharderHandler(t, "/v1/transaction/get/confirmation", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"txn":{"hash":"1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23","version":"1.0","client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","public_key":"eeb0c33325cbee0fb58bc09962f69a44d0b22ac2824a063eb1002273347e601a4612e6fea7e1a1ae62e0e3b7f1301c4de8a855bae86ebfa6e9dbbb41c3e39c24","transaction_data":"{\"OpType\":\"Move\",\"PreValue\":\"/1.txt\",\"CurrValue\":\"/d/1.txt\"}","transaction_value":0,"signature":"98427e25b635b8d88881ddc9a1f84db0951f145ffa90462c0290ed84563bdc92","creation_date":1617159987,"transaction_type":10,"transaction_fee":0,"txn_output_hash":""},"block_hash":"4dd9de1f3724a688686a5b54879cf424a9f8e6cb56ab77bfd19586dfcc804ba8"}`))
	})
	blockchain.SetQuerySleepTime(1)
	blockchain.SetMaxTxnQuery(3)

	err := WritePoolLock(time.Hour*720, "a324ee3369adfeedbd767ae9932970c67d52bf2018cd7b8ebcfe550dea6c6961", "", 500000000, 0)

	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
}

func TestWritePoolLock1(t *testing.T) {
	curQuerySleepTime := blockchain.GetQuerySleepTime()
	defer blockchain.SetQuerySleepTime(curQuerySleepTime)
	curMaxTxnQuery := blockchain.GetMaxTxnQuery()
	defer blockchain.SetMaxTxnQuery(curMaxTxnQuery)

	_, _, _, cncl := setupMockInitStorageSDK(t, configDir, 0)
	defer cncl()
	mocks.SetMinerHandler(t, "/v1/transaction/put", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	mocks.SetSharderHandler(t, "/v1/transaction/get/confirmation", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"txn":{"hash":"1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23","version":"1.0","client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","public_key":"eeb0c33325cbee0fb58bc09962f69a44d0b22ac2824a063eb1002273347e601a4612e6fea7e1a1ae62e0e3b7f1301c4de8a855bae86ebfa6e9dbbb41c3e39c24","transaction_data":"{\"OpType\":\"Move\",\"PreValue\":\"/1.txt\",\"CurrValue\":\"/d/1.txt\"}","transaction_value":0,"signature":"98427e25b635b8d88881ddc9a1f84db0951f145ffa90462c0290ed84563bdc92","creation_date":1617159987,"transaction_type":10,"transaction_fee":0,"txn_output_hash":""},"block_hash":"4dd9de1f3724a688686a5b54879cf424a9f8e6cb56ab77bfd19586dfcc804ba8"}`))
	})
	blockchain.SetQuerySleepTime(1)
	blockchain.SetMaxTxnQuery(3)

	err := WritePoolUnlock("366669cdd10bc977895f27ae94fb8b1fc9c1e81026c8378a508cc8f0167df9e7", 0)

	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
}

func TestGetChallengePoolInfo(t *testing.T) {
	infoStr := `{"id":"366669cdd10bc977895f27ae94fb8b1fc9c1e81026c8378a508cc8f0167df9e7","balance":90000000000,"start_time":1618160720,"expiration":1620753691,"finalized":false}`
	var expectedInfo *ChallengePoolInfo
	_ = json.Unmarshal([]byte(infoStr), &expectedInfo)

	_, _, _, cncl := setupMockInitStorageSDK(t, configDir, 0)
	defer cncl()
	mocks.SetSharderHandler(t, fmt.Sprintf("%v%v%v", "/v1/screst/", STORAGE_SCADDRESS, "/getChallengePoolStat"), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(infoStr))
	})

	info, err := GetChallengePoolInfo("f5bb9b7449acc7628eb6a1fbcb35cc8ca49b2f4c412c16586ec8eb9098fa8d11")

	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
	assert.Equal(t, expectedInfo, info)
}

func TestGetStorageSCConfig(t *testing.T) {
	infoStr := `{"min_alloc_size":1024,"min_alloc_duration":300000000000,"max_challenge_completion_time":1800000000000,"min_offer_duration":36000000000000,"min_blobber_capacity":1024,"readpool":{"min_lock":1000000000,"min_lock_period":60000000000,"max_lock_period":31536000000000000},"writepool":{"min_lock":1000000000,"min_lock_period":120000000000,"max_lock_period":31536000000000000},"stakepool":{"min_lock":1000000000,"interest_rate":0.0000334,"interest_interval":60000000000},"validator_reward":0.025,"blobber_slash":0.1,"max_read_price":1000000000000,"max_write_price":1000000000000,"failed_challenges_to_cancel":20,"failed_challenges_to_revoke_min_lock":10,"challenge_enabled":true,"max_challenges_per_generation":100,"challenge_rate_per_mb_min":1,"max_delegates":200,"max_charge":0.5,"time_unit":2592000000000000}`
	var expectedConf *StorageSCConfig
	_ = json.Unmarshal([]byte(infoStr), &expectedConf)

	_, _, _, cncl := setupMockInitStorageSDK(t, configDir, 0)
	defer cncl()
	mocks.SetSharderHandler(t, fmt.Sprintf("%v%v%v", "/v1/screst/", STORAGE_SCADDRESS, "/getConfig"), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(infoStr))
	})

	conf, err := GetStorageSCConfig()

	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
	assert.Equal(t, expectedConf, conf)
}

func TestGetBlobbers(t *testing.T) {
	blobbersStr := `{"Nodes": [{"id":"0de4fd069f1304f11ce5794ae4216f92326044392e74320932f5cba67bdeff09","url":"http://prod-node-201.fra.zcn.zeroservices.eu:5056","terms":{"read_price":127388535,"write_price":63694267,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"capacity":3543348019200,"used":0,"last_health_check":1617955541,"stake_pool_settings":{"delegate_wallet":"e73d88ba1133b8bc8e1683b8c11f30b9c172b4a067463f5657f82c0f2b77561a","min_stake":10000000000,"max_stake":1000000000000,"num_delegates":50,"service_charge":0.3}},{"id":"2b2ed3816062f47dce1b92c30ee060e833816a1e378cffca6dd18fc413baf7b6","url":"http://prod-node-201.fra.zcn.zeroservices.eu:5051","terms":{"read_price":127388535,"write_price":63694267,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"capacity":3543348019200,"used":0,"last_health_check":1617954789,"stake_pool_settings":{"delegate_wallet":"e73d88ba1133b8bc8e1683b8c11f30b9c172b4a067463f5657f82c0f2b77561a","min_stake":10000000000,"max_stake":1000000000000,"num_delegates":50,"service_charge":0.3}},{"id":"4e729c0ef0b8177df1c130dde2988d4cb7456afa183bf93cbf3bd74120eeebaf","url":"http://one.devnet-0chain.net:31306","terms":{"read_price":100000000,"write_price":100000000,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"capacity":107374182400,"used":39191576611,"last_health_check":1618161842,"stake_pool_settings":{"delegate_wallet":"63873aca9102193fc8c6aedb79d2f57f7468ab768bbc270ee2f0fc97b21345c6","min_stake":10000000000,"max_stake":1000000000000,"num_delegates":50,"service_charge":0.3}},{"id":"63230aeda3360ae6540e8604db40333638d963edac8d92fb9b576915155d5dc6","url":"http://one.devnet-0chain.net:31302","terms":{"read_price":100000000,"write_price":100000000,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"capacity":107374182400,"used":40265318435,"last_health_check":1618161842,"stake_pool_settings":{"delegate_wallet":"63873aca9102193fc8c6aedb79d2f57f7468ab768bbc270ee2f0fc97b21345c6","min_stake":10000000000,"max_stake":1000000000000,"num_delegates":50,"service_charge":0.3}}]}`
	var nodes struct{ Nodes []*Blobber }
	_ = json.Unmarshal([]byte(blobbersStr), &nodes)
	expectedBlobbers := nodes.Nodes

	_, _, _, cncl := setupMockInitStorageSDK(t, configDir, 0)
	defer cncl()
	mocks.SetSharderHandler(t, fmt.Sprintf("%v%v%v", "/v1/screst/", STORAGE_SCADDRESS, "/getblobbers"), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(blobbersStr))
	})

	blobbers, err := GetBlobbers()

	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
	assert.Equal(t, expectedBlobbers, blobbers)
}

func TestGetBlobber(t *testing.T) {
	blobberStr := `{"id":"68d17529d461a1257cbca5058fd92a28424bf783660bace37f362e5cb02520d5","url":"http://one.devnet-0chain.net:31303","terms":{"read_price":100000000,"write_price":100000000,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"capacity":107374182400,"used":39728447523,"last_health_check":1618162347,"stake_pool_settings":{"delegate_wallet":"63873aca9102193fc8c6aedb79d2f57f7468ab768bbc270ee2f0fc97b21345c6","min_stake":10000000000,"max_stake":1000000000000,"num_delegates":50,"service_charge":0.3}}`
	var expectedBlobber *Blobber
	_ = json.Unmarshal([]byte(blobberStr), &expectedBlobber)

	_, _, _, cncl := setupMockInitStorageSDK(t, configDir, 0)
	defer cncl()
	mocks.SetSharderHandler(t, fmt.Sprintf("%v%v%v", "/v1/screst/", STORAGE_SCADDRESS, "/getBlobber"), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(blobberStr))
	})

	blobber, err := GetBlobber("2b2ed3816062f47dce1b92c30ee060e833816a1e378cffca6dd18fc413baf7b6")

	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
	assert.Equal(t, expectedBlobber, blobber)
}

func TestGetClientEncryptedPublicKey(t *testing.T) {
	_, _, _, cncl := setupMockInitStorageSDK(t, configDir, 0)
	defer cncl()

	pubKey, err := GetClientEncryptedPublicKey()
	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
	assert.NotEmpty(t, pubKey)
}

func TestGetAllocationFromAuthTicket(t *testing.T) {
	allocStr := `{"id":"69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f","tx":"69fe503551eea5559c92712dffc932d8cfecd8ae641b2f242db29887e9ce618f","data_shards":2,"parity_shards":2,"size":2147483648,"expiration_date":1620643139,"owner_id":"d477d12134c2d7ba5ab71ac8ad37f244224695ef3215be990c3215d531c5a329","owner_public_key":"a3ed59da959b4a88d3612e558a3c78a3c2cf73184246df0e785c9d21f44d6c21b29403544352214c5bc6039abc72acd61eb8d9a75a0b6666d2043a9a3f46930c","payer_id":"","blobbers":[{"id":"4e729c0ef0b8177df1c130dde2988d4cb7456afa183bf93cbf3bd74120eeebaf","url":"http://one.devnet-0chain.net:31306"},{"id":"63230aeda3360ae6540e8604db40333638d963edac8d92fb9b576915155d5dc6","url":"http://one.devnet-0chain.net:31302"},{"id":"749154fee3f95528d96e4651413a56f0be8ceb2e9daec8d11fad93efa0082b77","url":"http://one.devnet-0chain.net:31305"},{"id":"a983087814ebd0bf267449dc179fe1790c18eb305ab888807196741aa9adbc97","url":"http://one.devnet-0chain.net:31301"}],"stats":{"used_size":8,"num_of_writes":12,"num_of_reads":0,"total_challenges":18,"num_open_challenges":18,"num_success_challenges":0,"num_failed_challenges":0,"latest_closed_challenge":""},"time_unit":2592000000000000,"blobber_details":[{"blobber_id":"63230aeda3360ae6540e8604db40333638d963edac8d92fb9b576915155d5dc6","size":536870912,"terms":{"read_price":100000000,"write_price":100000000,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"min_lock_demand":5000000,"spent":244140,"penalty":0,"read_reward":244140,"returned":0,"challenge_reward":0,"final_reward":0},{"blobber_id":"a983087814ebd0bf267449dc179fe1790c18eb305ab888807196741aa9adbc97","size":536870912,"terms":{"read_price":100000000,"write_price":100000000,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"min_lock_demand":5000000,"spent":244140,"penalty":0,"read_reward":244140,"returned":0,"challenge_reward":0,"final_reward":0},{"blobber_id":"4e729c0ef0b8177df1c130dde2988d4cb7456afa183bf93cbf3bd74120eeebaf","size":536870912,"terms":{"read_price":100000000,"write_price":100000000,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"min_lock_demand":5000000,"spent":244140,"penalty":0,"read_reward":244140,"returned":0,"challenge_reward":0,"final_reward":0},{"blobber_id":"749154fee3f95528d96e4651413a56f0be8ceb2e9daec8d11fad93efa0082b77","size":536870912,"terms":{"read_price":100000000,"write_price":100000000,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"min_lock_demand":5000000,"spent":244140,"penalty":0,"read_reward":244140,"returned":0,"challenge_reward":0,"final_reward":0}],"read_price_range":{"min":100000000,"max":200000000},"write_price_range":{"min":100000000,"max":200000000},"challenge_completion_time":120000000000,"start_time":1618051139}`
	var expectedAlloc *Allocation
	_ = json.Unmarshal([]byte(allocStr), &expectedAlloc)

	_, _, _, cncl := setupMockInitStorageSDK(t, configDir, 0)
	defer cncl()
	mocks.SetSharderHandler(t, fmt.Sprintf("%v%v%v", "/v1/screst/", STORAGE_SCADDRESS, "/allocation"), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(allocStr))
	})
	authTicketStr := "eyJjbGllbnRfaWQiOiI5YmY0MzBkNmYwODZmMWJkYzJkMjZhZDdhNzA4YTBlNzk1OGFhOWFlMjBlZmJjNjc3ODQ1MDczOWZiMWNhNDY4Iiwib3duZXJfaWQiOiI5YmY0MzBkNmYwODZmMWJkYzJkMjZhZDdhNzA4YTBlNzk1OGFhOWFlMjBlZmJjNjc3ODQ1MDczOWZiMWNhNDY4IiwiYWxsb2NhdGlvbl9pZCI6IjY5ZmU1MDM1NTFlZWE1NTU5YzkyNzEyZGZmYzkzMmQ4Y2ZlY2Q4YWU2NDFiMmYyNDJkYjI5ODg3ZTljZTYxOGYiLCJmaWxlX3BhdGhfaGFzaCI6ImM4ODRhYmIzMmFhMDM1N2UyNTQxYjY4M2Y2ZTUyYmZhYjkxNDNkMzNiOTY4OTc3Y2Y2YmEzMWI0M2U4MzI2OTciLCJmaWxlX25hbWUiOiIxLnR4dCIsInJlZmVyZW5jZV90eXBlIjoiZiIsImV4cGlyYXRpb24iOjE2MjQ5OTE3NDcsInRpbWVzdGFtcCI6MTYxNzIxNTc0NywicmVfZW5jcnlwdGlvbl9rZXkiOiIiLCJzaWduYXR1cmUiOiI1Mjk3Y2UyYzVlNzU1NTFhMmJmNWEzMmQ3YmU2MzM4N2U5NzIxZTM2N2QzMDc5ZTI1ZmViZDFkMmIxMWE2NzIwIn0="

	alloc, err := GetAllocationFromAuthTicket(authTicketStr)

	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
	assert.Equal(t, expectedAlloc.ID, alloc.ID)
	assert.Equal(t, expectedAlloc.Tx, alloc.Tx)
	assert.Equal(t, expectedAlloc.DataShards, alloc.DataShards)
	assert.Equal(t, expectedAlloc.ParityShards, alloc.ParityShards)
	assert.Equal(t, expectedAlloc.Size, alloc.Size)
	assert.Equal(t, expectedAlloc.Expiration, alloc.Expiration)
	assert.Equal(t, expectedAlloc.Owner, alloc.Owner)
	assert.Equal(t, expectedAlloc.OwnerPublicKey, alloc.OwnerPublicKey)
	assert.EqualValues(t, expectedAlloc.Payer, alloc.Payer)
	assert.EqualValues(t, expectedAlloc.Stats, alloc.Stats)
}

func TestSetNumBlockDownloads(t *testing.T) {
	t.Run("Test_Set_Num_Block_Download_0", func(t *testing.T) {
		nbd := numBlockDownloads
		defer func() { numBlockDownloads = nbd }()

		SetNumBlockDownloads(0)
		assert.Equal(t, nbd, numBlockDownloads)
	})
	t.Run("Test_Set_Num_Block_Download_100", func(t *testing.T) {
		nbd := numBlockDownloads
		defer func() { numBlockDownloads = nbd }()

		SetNumBlockDownloads(100)
		assert.Equal(t, 100, numBlockDownloads)
	})
	t.Run("Test_Set_Num_Block_Download_101", func(t *testing.T) {
		nbd := numBlockDownloads
		defer func() { numBlockDownloads = nbd }()

		SetNumBlockDownloads(101)
		assert.Equal(t, nbd, numBlockDownloads)
	})
	t.Run("Test_Set_Num_Block_Download_50", func(t *testing.T) {
		nbd := numBlockDownloads
		defer func() { numBlockDownloads = nbd }()

		SetNumBlockDownloads(50)
		assert.Equal(t, 50, numBlockDownloads)
	})
}

func TestGetAllocations(t *testing.T) {
	allocationsStr := `[{"id":"a324ee3369adfeedbd767ae9932970c67d52bf2018cd7b8ebcfe550dea6c6961","tx":"a324ee3369adfeedbd767ae9932970c67d52bf2018cd7b8ebcfe550dea6c6961","data_shards":2,"parity_shards":2,"size":2147483648,"expiration_date":1620643080,"owner_id":"d477d12134c2d7ba5ab71ac8ad37f244224695ef3215be990c3215d531c5a329","owner_public_key":"a3ed59da959b4a88d3612e558a3c78a3c2cf73184246df0e785c9d21f44d6c21b29403544352214c5bc6039abc72acd61eb8d9a75a0b6666d2043a9a3f46930c","payer_id":"","blobbers":[{"id":"63230aeda3360ae6540e8604db40333638d963edac8d92fb9b576915155d5dc6","url":"http://one.devnet-0chain.net:31302"},{"id":"749154fee3f95528d96e4651413a56f0be8ceb2e9daec8d11fad93efa0082b77","url":"http://one.devnet-0chain.net:31305"},{"id":"833bc63ba483a9c4e557f609f2afb66bc0983d4067999e7d93843b4111ed5507","url":"http://one.devnet-0chain.net:31304"},{"id":"a983087814ebd0bf267449dc179fe1790c18eb305ab888807196741aa9adbc97","url":"http://one.devnet-0chain.net:31301"}],"stats":{"used_size":0,"num_of_writes":0,"num_of_reads":0,"total_challenges":0,"num_open_challenges":0,"num_success_challenges":0,"num_failed_challenges":0,"latest_closed_challenge":""},"time_unit":2592000000000000,"blobber_details":[{"blobber_id":"749154fee3f95528d96e4651413a56f0be8ceb2e9daec8d11fad93efa0082b77","size":536870912,"terms":{"read_price":100000000,"write_price":100000000,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"min_lock_demand":5000000,"spent":0,"penalty":0,"read_reward":0,"returned":0,"challenge_reward":0,"final_reward":0},{"blobber_id":"63230aeda3360ae6540e8604db40333638d963edac8d92fb9b576915155d5dc6","size":536870912,"terms":{"read_price":100000000,"write_price":100000000,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"min_lock_demand":5000000,"spent":0,"penalty":0,"read_reward":0,"returned":0,"challenge_reward":0,"final_reward":0},{"blobber_id":"a983087814ebd0bf267449dc179fe1790c18eb305ab888807196741aa9adbc97","size":536870912,"terms":{"read_price":100000000,"write_price":100000000,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"min_lock_demand":5000000,"spent":0,"penalty":0,"read_reward":0,"returned":0,"challenge_reward":0,"final_reward":0},{"blobber_id":"833bc63ba483a9c4e557f609f2afb66bc0983d4067999e7d93843b4111ed5507","size":536870912,"terms":{"read_price":100000000,"write_price":100000000,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"min_lock_demand":5000000,"spent":0,"penalty":0,"read_reward":0,"returned":0,"challenge_reward":0,"final_reward":0}],"read_price_range":{"min":100000000,"max":200000000},"write_price_range":{"min":100000000,"max":200000000},"challenge_completion_time":120000000000,"start_time":1618051080},{"id":"b78e4a211213a0296ddf9cd06133c8f5dca9dd256b5694e3fa824b7d33a355ef","tx":"b78e4a211213a0296ddf9cd06133c8f5dca9dd256b5694e3fa824b7d33a355ef","data_shards":4,"parity_shards":2,"size":2147483648,"expiration_date":1620642882,"owner_id":"d477d12134c2d7ba5ab71ac8ad37f244224695ef3215be990c3215d531c5a329","owner_public_key":"a3ed59da959b4a88d3612e558a3c78a3c2cf73184246df0e785c9d21f44d6c21b29403544352214c5bc6039abc72acd61eb8d9a75a0b6666d2043a9a3f46930c","payer_id":"","blobbers":[{"id":"4e729c0ef0b8177df1c130dde2988d4cb7456afa183bf93cbf3bd74120eeebaf","url":"http://one.devnet-0chain.net:31306"},{"id":"63230aeda3360ae6540e8604db40333638d963edac8d92fb9b576915155d5dc6","url":"http://one.devnet-0chain.net:31302"},{"id":"68d17529d461a1257cbca5058fd92a28424bf783660bace37f362e5cb02520d5","url":"http://one.devnet-0chain.net:31303"},{"id":"749154fee3f95528d96e4651413a56f0be8ceb2e9daec8d11fad93efa0082b77","url":"http://one.devnet-0chain.net:31305"},{"id":"833bc63ba483a9c4e557f609f2afb66bc0983d4067999e7d93843b4111ed5507","url":"http://one.devnet-0chain.net:31304"},{"id":"a983087814ebd0bf267449dc179fe1790c18eb305ab888807196741aa9adbc97","url":"http://one.devnet-0chain.net:31301"}],"stats":{"used_size":0,"num_of_writes":0,"num_of_reads":0,"total_challenges":0,"num_open_challenges":0,"num_success_challenges":0,"num_failed_challenges":0,"latest_closed_challenge":""},"time_unit":2592000000000000,"blobber_details":[{"blobber_id":"a983087814ebd0bf267449dc179fe1790c18eb305ab888807196741aa9adbc97","size":357913942,"terms":{"read_price":100000000,"write_price":100000000,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"min_lock_demand":3333333,"spent":0,"penalty":0,"read_reward":0,"returned":0,"challenge_reward":0,"final_reward":0},{"blobber_id":"63230aeda3360ae6540e8604db40333638d963edac8d92fb9b576915155d5dc6","size":357913942,"terms":{"read_price":100000000,"write_price":100000000,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"min_lock_demand":3333333,"spent":0,"penalty":0,"read_reward":0,"returned":0,"challenge_reward":0,"final_reward":0},{"blobber_id":"68d17529d461a1257cbca5058fd92a28424bf783660bace37f362e5cb02520d5","size":357913942,"terms":{"read_price":100000000,"write_price":100000000,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"min_lock_demand":3333333,"spent":0,"penalty":0,"read_reward":0,"returned":0,"challenge_reward":0,"final_reward":0},{"blobber_id":"749154fee3f95528d96e4651413a56f0be8ceb2e9daec8d11fad93efa0082b77","size":357913942,"terms":{"read_price":100000000,"write_price":100000000,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"min_lock_demand":3333333,"spent":0,"penalty":0,"read_reward":0,"returned":0,"challenge_reward":0,"final_reward":0},{"blobber_id":"833bc63ba483a9c4e557f609f2afb66bc0983d4067999e7d93843b4111ed5507","size":357913942,"terms":{"read_price":100000000,"write_price":100000000,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"min_lock_demand":3333333,"spent":0,"penalty":0,"read_reward":0,"returned":0,"challenge_reward":0,"final_reward":0},{"blobber_id":"4e729c0ef0b8177df1c130dde2988d4cb7456afa183bf93cbf3bd74120eeebaf","size":357913942,"terms":{"read_price":100000000,"write_price":100000000,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"min_lock_demand":3333333,"spent":0,"penalty":0,"read_reward":0,"returned":0,"challenge_reward":0,"final_reward":0}],"read_price_range":{"min":0,"max":10000000000},"write_price_range":{"min":0,"max":10000000000},"challenge_completion_time":120000000000,"start_time":1618050882},{"id":"f5bb9b7449acc7628eb6a1fbcb35cc8ca49b2f4c412c16586ec8eb9098fa8d11","tx":"f5bb9b7449acc7628eb6a1fbcb35cc8ca49b2f4c412c16586ec8eb9098fa8d11","data_shards":2,"parity_shards":2,"size":2147483648,"expiration_date":1620643139,"owner_id":"d477d12134c2d7ba5ab71ac8ad37f244224695ef3215be990c3215d531c5a329","owner_public_key":"a3ed59da959b4a88d3612e558a3c78a3c2cf73184246df0e785c9d21f44d6c21b29403544352214c5bc6039abc72acd61eb8d9a75a0b6666d2043a9a3f46930c","payer_id":"","blobbers":[{"id":"4e729c0ef0b8177df1c130dde2988d4cb7456afa183bf93cbf3bd74120eeebaf","url":"http://one.devnet-0chain.net:31306"},{"id":"63230aeda3360ae6540e8604db40333638d963edac8d92fb9b576915155d5dc6","url":"http://one.devnet-0chain.net:31302"},{"id":"749154fee3f95528d96e4651413a56f0be8ceb2e9daec8d11fad93efa0082b77","url":"http://one.devnet-0chain.net:31305"},{"id":"a983087814ebd0bf267449dc179fe1790c18eb305ab888807196741aa9adbc97","url":"http://one.devnet-0chain.net:31301"}],"stats":{"used_size":8,"num_of_writes":12,"num_of_reads":0,"total_challenges":18,"num_open_challenges":18,"num_success_challenges":0,"num_failed_challenges":0,"latest_closed_challenge":""},"time_unit":2592000000000000,"blobber_details":[{"blobber_id":"63230aeda3360ae6540e8604db40333638d963edac8d92fb9b576915155d5dc6","size":536870912,"terms":{"read_price":100000000,"write_price":100000000,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"min_lock_demand":5000000,"spent":244140,"penalty":0,"read_reward":244140,"returned":0,"challenge_reward":0,"final_reward":0},{"blobber_id":"a983087814ebd0bf267449dc179fe1790c18eb305ab888807196741aa9adbc97","size":536870912,"terms":{"read_price":100000000,"write_price":100000000,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"min_lock_demand":5000000,"spent":244140,"penalty":0,"read_reward":244140,"returned":0,"challenge_reward":0,"final_reward":0},{"blobber_id":"4e729c0ef0b8177df1c130dde2988d4cb7456afa183bf93cbf3bd74120eeebaf","size":536870912,"terms":{"read_price":100000000,"write_price":100000000,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"min_lock_demand":5000000,"spent":244140,"penalty":0,"read_reward":244140,"returned":0,"challenge_reward":0,"final_reward":0},{"blobber_id":"749154fee3f95528d96e4651413a56f0be8ceb2e9daec8d11fad93efa0082b77","size":536870912,"terms":{"read_price":100000000,"write_price":100000000,"min_lock_demand":0.1,"max_offer_duration":2678400000000000,"challenge_completion_time":120000000000},"min_lock_demand":5000000,"spent":244140,"penalty":0,"read_reward":244140,"returned":0,"challenge_reward":0,"final_reward":0}],"read_price_range":{"min":100000000,"max":200000000},"write_price_range":{"min":100000000,"max":200000000},"challenge_completion_time":120000000000,"start_time":1618051139}]`
	var expectedAllocs []*Allocation
	_ = json.Unmarshal([]byte(allocationsStr), &expectedAllocs)

	_, _, _, cncl := setupMockInitStorageSDK(t, configDir, 0)
	defer cncl()
	mocks.SetSharderHandler(t, fmt.Sprintf("%v%v%v", "/v1/screst/", STORAGE_SCADDRESS, "/allocations"), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(allocationsStr))
	})

	allocs, err := GetAllocations()

	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
	assert.Equal(t, len(expectedAllocs), len(allocs))
	for idx, expectedAlloc := range expectedAllocs {
		assert.NotNil(t, allocs[idx])
		assert.Equal(t, expectedAlloc.ID, allocs[idx].ID)
		assert.Equal(t, expectedAlloc.Tx, allocs[idx].Tx)
		assert.Equal(t, expectedAlloc.DataShards, allocs[idx].DataShards)
		assert.Equal(t, expectedAlloc.ParityShards, allocs[idx].ParityShards)
		assert.Equal(t, expectedAlloc.Size, allocs[idx].Size)
		assert.Equal(t, expectedAlloc.Expiration, allocs[idx].Expiration)
		assert.Equal(t, expectedAlloc.Owner, allocs[idx].Owner)
		assert.Equal(t, expectedAlloc.OwnerPublicKey, allocs[idx].OwnerPublicKey)
		assert.EqualValues(t, expectedAlloc.Payer, allocs[idx].Payer)
		assert.EqualValues(t, expectedAlloc.Stats, allocs[idx].Stats)
	}
}

func TestCreateAllocation(t *testing.T) {
	curQuerySleepTime := blockchain.GetQuerySleepTime()
	defer blockchain.SetQuerySleepTime(curQuerySleepTime)
	curMaxTxnQuery := blockchain.GetMaxTxnQuery()
	defer blockchain.SetMaxTxnQuery(curMaxTxnQuery)

	_, _, _, cncl := setupMockInitStorageSDK(t, configDir, 0)
	defer cncl()
	mocks.SetMinerHandler(t, "/v1/transaction/put", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	mocks.SetSharderHandler(t, "/v1/transaction/get/confirmation", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"txn":{"hash":"1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23","version":"1.0","client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","public_key":"eeb0c33325cbee0fb58bc09962f69a44d0b22ac2824a063eb1002273347e601a4612e6fea7e1a1ae62e0e3b7f1301c4de8a855bae86ebfa6e9dbbb41c3e39c24","transaction_data":"{\"OpType\":\"Move\",\"PreValue\":\"/1.txt\",\"CurrValue\":\"/d/1.txt\"}","transaction_value":0,"signature":"98427e25b635b8d88881ddc9a1f84db0951f145ffa90462c0290ed84563bdc92","creation_date":1617159987,"transaction_type":10,"transaction_fee":0,"txn_output_hash":""},"block_hash":"4dd9de1f3724a688686a5b54879cf424a9f8e6cb56ab77bfd19586dfcc804ba8"}`))
	})
	blockchain.SetQuerySleepTime(1)
	blockchain.SetMaxTxnQuery(3)

	hash, err := CreateAllocation(2, 3, 2*1024*1024*1024, time.Now().Add(720*time.Hour).Unix(), PriceRange{Min: 100000000, Max: 3000000000}, PriceRange{Min: 100000000, Max: 3000000000}, 7*24*time.Hour, 10000000000)

	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
	assert.Equal(t, "1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23", hash)
}

func TestUpdateAllocation(t *testing.T) {
	curQuerySleepTime := blockchain.GetQuerySleepTime()
	defer blockchain.SetQuerySleepTime(curQuerySleepTime)
	curMaxTxnQuery := blockchain.GetMaxTxnQuery()
	defer blockchain.SetMaxTxnQuery(curMaxTxnQuery)

	_, _, _, cncl := setupMockInitStorageSDK(t, configDir, 0)
	defer cncl()
	mocks.SetMinerHandler(t, "/v1/transaction/put", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	mocks.SetSharderHandler(t, "/v1/transaction/get/confirmation", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"txn":{"hash":"1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23","version":"1.0","client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","public_key":"eeb0c33325cbee0fb58bc09962f69a44d0b22ac2824a063eb1002273347e601a4612e6fea7e1a1ae62e0e3b7f1301c4de8a855bae86ebfa6e9dbbb41c3e39c24","transaction_data":"{\"OpType\":\"Move\",\"PreValue\":\"/1.txt\",\"CurrValue\":\"/d/1.txt\"}","transaction_value":0,"signature":"98427e25b635b8d88881ddc9a1f84db0951f145ffa90462c0290ed84563bdc92","creation_date":1617159987,"transaction_type":10,"transaction_fee":0,"txn_output_hash":""},"block_hash":"4dd9de1f3724a688686a5b54879cf424a9f8e6cb56ab77bfd19586dfcc804ba8"}`))
	})
	blockchain.SetQuerySleepTime(1)
	blockchain.SetMaxTxnQuery(3)

	hash, err := UpdateAllocation(2*1024*1024*1024, time.Now().Add(720*time.Hour).Unix(), "a324ee3369adfeedbd767ae9932970c67d52bf2018cd7b8ebcfe550dea6c6961", 10000000000)

	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
	assert.Equal(t, "1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23", hash)
}

func TestFinalizeAllocation(t *testing.T) {
	curQuerySleepTime := blockchain.GetQuerySleepTime()
	defer blockchain.SetQuerySleepTime(curQuerySleepTime)
	curMaxTxnQuery := blockchain.GetMaxTxnQuery()
	defer blockchain.SetMaxTxnQuery(curMaxTxnQuery)

	_, _, _, cncl := setupMockInitStorageSDK(t, configDir, 0)
	defer cncl()
	mocks.SetMinerHandler(t, "/v1/transaction/put", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	mocks.SetSharderHandler(t, "/v1/transaction/get/confirmation", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"txn":{"hash":"1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23","version":"1.0","client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","public_key":"eeb0c33325cbee0fb58bc09962f69a44d0b22ac2824a063eb1002273347e601a4612e6fea7e1a1ae62e0e3b7f1301c4de8a855bae86ebfa6e9dbbb41c3e39c24","transaction_data":"{\"OpType\":\"Move\",\"PreValue\":\"/1.txt\",\"CurrValue\":\"/d/1.txt\"}","transaction_value":0,"signature":"98427e25b635b8d88881ddc9a1f84db0951f145ffa90462c0290ed84563bdc92","creation_date":1617159987,"transaction_type":10,"transaction_fee":0,"txn_output_hash":""},"block_hash":"4dd9de1f3724a688686a5b54879cf424a9f8e6cb56ab77bfd19586dfcc804ba8"}`))
	})
	blockchain.SetQuerySleepTime(1)
	blockchain.SetMaxTxnQuery(3)

	hash, err := FinalizeAllocation("a324ee3369adfeedbd767ae9932970c67d52bf2018cd7b8ebcfe550dea6c6961")

	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
	assert.Equal(t, "1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23", hash)
}

func TestCancelAllocation(t *testing.T) {
	curQuerySleepTime := blockchain.GetQuerySleepTime()
	defer blockchain.SetQuerySleepTime(curQuerySleepTime)
	curMaxTxnQuery := blockchain.GetMaxTxnQuery()
	defer blockchain.SetMaxTxnQuery(curMaxTxnQuery)

	_, _, _, cncl := setupMockInitStorageSDK(t, configDir, 0)
	defer cncl()
	mocks.SetMinerHandler(t, "/v1/transaction/put", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	mocks.SetSharderHandler(t, "/v1/transaction/get/confirmation", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"txn":{"hash":"1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23","version":"1.0","client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","public_key":"eeb0c33325cbee0fb58bc09962f69a44d0b22ac2824a063eb1002273347e601a4612e6fea7e1a1ae62e0e3b7f1301c4de8a855bae86ebfa6e9dbbb41c3e39c24","transaction_data":"{\"OpType\":\"Move\",\"PreValue\":\"/1.txt\",\"CurrValue\":\"/d/1.txt\"}","transaction_value":0,"signature":"98427e25b635b8d88881ddc9a1f84db0951f145ffa90462c0290ed84563bdc92","creation_date":1617159987,"transaction_type":10,"transaction_fee":0,"txn_output_hash":""},"block_hash":"4dd9de1f3724a688686a5b54879cf424a9f8e6cb56ab77bfd19586dfcc804ba8"}`))
	})
	blockchain.SetQuerySleepTime(1)
	blockchain.SetMaxTxnQuery(3)

	hash, err := CancelAllocation("a324ee3369adfeedbd767ae9932970c67d52bf2018cd7b8ebcfe550dea6c6961")

	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
	assert.Equal(t, "1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23", hash)
}

func TestUpdateBlobberSettings(t *testing.T) {
	curQuerySleepTime := blockchain.GetQuerySleepTime()
	defer blockchain.SetQuerySleepTime(curQuerySleepTime)
	curMaxTxnQuery := blockchain.GetMaxTxnQuery()
	defer blockchain.SetMaxTxnQuery(curMaxTxnQuery)

	_, _, _, cncl := setupMockInitStorageSDK(t, configDir, 0)
	defer cncl()
	mocks.SetMinerHandler(t, "/v1/transaction/put", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	mocks.SetSharderHandler(t, "/v1/transaction/get/confirmation", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"txn":{"hash":"1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23","version":"1.0","client_id":"9bf430d6f086f1bdc2d26ad7a708a0e7958aa9ae20efbc6778450739fb1ca468","public_key":"eeb0c33325cbee0fb58bc09962f69a44d0b22ac2824a063eb1002273347e601a4612e6fea7e1a1ae62e0e3b7f1301c4de8a855bae86ebfa6e9dbbb41c3e39c24","transaction_data":"{\"OpType\":\"Move\",\"PreValue\":\"/1.txt\",\"CurrValue\":\"/d/1.txt\"}","transaction_value":0,"signature":"98427e25b635b8d88881ddc9a1f84db0951f145ffa90462c0290ed84563bdc92","creation_date":1617159987,"transaction_type":10,"transaction_fee":0,"txn_output_hash":""},"block_hash":"4dd9de1f3724a688686a5b54879cf424a9f8e6cb56ab77bfd19586dfcc804ba8"}`))
	})
	blockchain.SetQuerySleepTime(1)
	blockchain.SetMaxTxnQuery(3)

	resp, err := UpdateBlobberSettings(&Blobber{})

	assert.NoErrorf(t, err, "unexpected error but got: %v", err)
	assert.Equal(t, "1309ee2ab8d21b213e959ab0e26201d734bd2752945d9897cb9d98a3c11a6a23", resp)
}

func TestCommitToFabric(t *testing.T) {
	type args struct {
		metaTxnData      string
		fabricConfigJSON string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CommitToFabric(tt.args.metaTxnData, tt.args.fabricConfigJSON)
			if (err != nil) != tt.wantErr {
				t.Errorf("CommitToFabric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CommitToFabric() got = %v, want %v", got, tt.want)
			}
		})
	}
}