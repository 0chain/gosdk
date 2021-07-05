package zcncore

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/zcncrypto"
	zcncryptomock "github.com/0chain/gosdk/core/zcncrypto/mocks"
	"github.com/0chain/gosdk/zcnmocks"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	mockClientID   = "mock client id"
	mockPrivateKey = "62fc118369fb9dd1fa6065d4f8f765c52ac68ad5aced17a1e5c4f8b4301a9469b987071c14695caf340ea11560f5a3cb76ad1e709803a8b339826ab3964e470a"
	mockPublicKey  = "b987071c14695caf340ea11560f5a3cb76ad1e709803a8b339826ab3964e470a"
)

var verifyPublickey = `e8a6cfa7b3076ae7e04764ffdfe341632a136b52953dfafa6926361dd9a466196faecca6f696774bbd64b938ff765dbc837e8766a5e2d8996745b2b94e1beb9e`
var signPrivatekey = `5e1fc9c03d53a8b9a63030acc2864f0c33dffddb3c276bf2b3c8d739269cc018`

//RUNOK
func TestNewTransactionWithAuth(t *testing.T) {
	t.Run("Test New Transaction With Auth Success", func(t *testing.T) {
		mockWalletCallback := MockTransactionCallback{}
		mockWalletCallback.On("OnTransactionComplete", &Transaction{}, 0).Return()
		resp, err := newTransactionWithAuth(mockWalletCallback, 1)
		require.NotEmpty(t, resp)
		// expectedErrorMsg := "magic block info not found"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)\
		require.NoError(t, err)
	})
}

//RUNOK
func TestTransactionAuthSetTransactionCallback(t *testing.T) {
	t.Run("Test New Transaction With Auth transaction already exists", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{},
		}
		mockWalletCallback := MockTransactionCallback{}
		mockWalletCallback.On("OnTransactionComplete", &Transaction{}, 0).Return()
		err := ta.SetTransactionCallback(mockWalletCallback)
		expectedErrorMsg := "transaction already exists. cannot set transaction hash."
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
		// require.NoError(t, err)
	})
	t.Run("Test New Transaction With Auth success", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txnStatus: -1,
			},
		}
		mockWalletCallback := MockTransactionCallback{}
		mockWalletCallback.On("OnTransactionComplete", &Transaction{}, 0).Return()
		err := ta.SetTransactionCallback(mockWalletCallback)

		require.NoError(t, err)
	})
}

//RUNOK
func TestTransactionAuthSetTransactionFee(t *testing.T) {
	t.Run("Test Transaction Auth Set Transaction Fee", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{},
		}

		err := ta.SetTransactionFee(1)
		expectedErrorMsg := "transaction already exists. cannot set transaction fee."
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
		// require.NoError(t, err)
	})
	t.Run("Test Transaction Auth Set Transaction Fee", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txnStatus: -1,
				txn:       &transaction.Transaction{},
			},
		}

		err := ta.SetTransactionFee(1)

		require.NoError(t, err)
	})
}

//RUNOK
func TestVerifyFn(t *testing.T) {
	t.Run("Test Verify Fn", func(t *testing.T) {
		resp, err := verifyFn(mnemonic, hash, public_key)
		// expectedErrorMsg := "signature_mismatch"
		// require.Equal(t,expectedErrorMsg,err)

		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
		require.NotNil(t, err)
		require.Equal(t, false, resp)
	})
}

//RUNOK
func TestSend(t *testing.T) {
	var (
		mockClientID     = "mock client id"
		mockTxnData      = "mock txn data"
		mockValue        = int64(1)
	)

	t.Run("TestSend", func(t *testing.T) {
		ta := setupTestZcncore(mockClientID,mockTxnData,mockValue,"TestSend")

		err := ta.Send(mockClientID, mockValue, mockTxnData)
		require.NoError(t, err)
	})
}

//RUNOK
func TestStoreData(t *testing.T) {
	var (
		mockTxnData      = "mock txn data"
		mockValue        = int64(1)
	)

	t.Run("TestStoreData", func(t *testing.T) {
		ta := setupTestZcncore(mockClientID,mockTxnData,mockValue,"TestStoreData")

		err := ta.StoreData(mockTxnData)
		require.NoError(t, err)
	})
}

//RUNOK
func TestTransactionAuthExecuteFaucetSCWallet(t *testing.T) {
	var (
		mockWalletString = `{"client_id":"679b06b89fc418cfe7f8fc908137795de8b7777e9324901432acce4781031c93","client_key":"2c2aaca87c9d80108c4d5dc27fc8eefc57be98af55d26a548ebf92a86cd90615d19d715a9ed6d009798877189babf405384a2980e102ce72f824890b20f8ce1e","keys":[{"public_key":"2c2aaca87c9d80108c4d5dc27fc8eefc57be98af55d26a548ebf92a86cd90615d19d715a9ed6d009798877189babf405384a2980e102ce72f824890b20f8ce1e","private_key":"mock private key"}],"mnemonics":"bamboo list citizen release bronze duck woman moment cart crucial extra hip witness mixture flash into priority length pattern deposit title exhaust flush addict","version":"1.0","date_created":"2021-06-15 11:11:40.306922176 +0700 +07 m=+1.187131283"}`

		mockToClientID   = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d3"
		mockTxnData      = `{"name":"GET","input":"dGVzdA=="}`
		mockValue        = int64(0)
	)

	t.Run("TestTransactionAuthExecuteFaucetSCWallet", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthExecuteFaucetSCWallet")

		err := ta.ExecuteFaucetSCWallet(mockWalletString, "GET", []byte("test"))
		require.NoError(t, err)
	})
}

//RUNOK
func TestExecuteSmartContract(t *testing.T) {
	var (
		mockToClientID   = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d3"
		mockTxnData      = `{"name":"GET","input":{}}`
		mockValue        = int64(1)
	)

	t.Run("TestExecuteSmartContract", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestExecuteSmartContract")

		scData := make(map[string]interface{})
		out, err := json.Marshal(scData)
		require.NoError(t, err)
		err = ta.ExecuteSmartContract(mockToClientID, "GET", string(out), 1)
		require.NoError(t, err)
	})
}

//RUNOK
func TestTransactionAuthSetTransactionHash(t *testing.T) {
	t.Run("Test Set Transaction Hash", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		err := ta.SetTransactionHash(hash)
		expectedErrorMsg := "transaction already exists. cannot set transaction hash."
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}

//RUNOK
func TestTransactionAuthGetTransactionHash(t *testing.T) {
	t.Run("Test Get Transaction Hash", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.GetTransactionHash()
		require.NotNil(t, resp)
	})
}

//RUNOK
func TestTransactionAuthVerify(t *testing.T) {
	t.Run("Test Transaction Auth Verify", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		err := ta.Verify()
		expectedErrorMsg := "invalid transaction. cannot be verified."
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}

//RUNOK
func TestTransactionAuthGetVerifyOutput(t *testing.T) {
	t.Run("Test Transaction Auth Get Verify Output", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.GetVerifyOutput()
		require.NotNil(t, resp)
	})
}

//RUNOK
func TestTransactionAuthGetTransactionError(t *testing.T) {
	t.Run("Test Transaction Auth Get Transaction Error", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.GetTransactionError()
		require.NotNil(t, resp)
	})
}

//RUNOK
func TestTransactionAuthGetVerifyError(t *testing.T) {
	t.Run("Test Transaction Auth Get Verify Error", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.GetVerifyError()
		require.NotNil(t, resp)
	})
}

//RUNOK
func TestTransactionAuthOutput(t *testing.T) {
	t.Run("Test Transaction Auth Output", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.Output()
		require.NotNil(t, resp)
	})
}

//RUNOK
func TestTransactionAuthRegisterMultiSig(t *testing.T) {
	t.Run("Test Transaction Auth Register MultiSig", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.RegisterMultiSig(walletString, msw)
		expectedErrorMsg := "not implemented"
		assert.EqualErrorf(t, resp, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, resp)
	})
}

//RUNOK
func TestTransactionAuthFinalizeAllocation(t *testing.T) {
	var (
		mockToClientID   = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7"
		mockTxnData      = `{"name":"finalize_allocation","input":{"allocation_id":"mock pool id"}}`
		mockValue        = int64(0)
	)

	t.Run("TestTransactionAuthFinalizeAllocation", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthFinalizeAllocation")

		err := ta.FinalizeAllocation("mock pool id", 1)
		require.NoError(t, err)
	})
}

//RUNOK
func TestTransactionAuthCancelAllocation(t *testing.T) {
	var (
		mockToClientID   = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7"
		mockTxnData      = `{"name":"cancel_allocation","input":{"allocation_id":"mock allocation id"}}`
		mockValue        = int64(0)
	)

	t.Run("TestTransactionAuthCancelAllocation", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthCancelAllocation")

		resp := ta.CancelAllocation("mock allocation id", 1)
		require.NoError(t, resp)
	})
}

//RUNOK
func TestTransactionAuthVestingTrigger(t *testing.T) {
	var (
		mockToClientID   = "2bba5b05949ea59c80aed3ac3474d7379d3be737e8eb5a968c52295e48333ead"
		mockTxnData      = `{"name":"trigger","input":{"pool_id":"mock pool id"}}`
		mockValue        = int64(0)
	)

	t.Run("TestTransactionAuthVestingStop", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthVestingStop")

		resp := ta.VestingTrigger("mock pool id")
		require.NoError(t, resp)
	})
}

//RUNOK
func TestTransactionAuthVestingStop(t *testing.T) {
	var (
		mockToClientID   = "2bba5b05949ea59c80aed3ac3474d7379d3be737e8eb5a968c52295e48333ead"
		mockTxnData      = `{"name":"stop","input":{"pool_id":"","destination":""}}`
		mockValue        = int64(0)
	)

	t.Run("TestTransactionAuthVestingStop", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthVestingStop")

		resp := ta.VestingStop(&VestingStopRequest{})
		require.NoError(t, resp)
	})
}

//RUNOK
func TestTransactionAuthVestingUnlock(t *testing.T) {
	var (
		mockToClientID   = "2bba5b05949ea59c80aed3ac3474d7379d3be737e8eb5a968c52295e48333ead"
		mockTxnData      = `{"name":"unlock","input":{"pool_id":"mock pool id"}}`
		mockValue        = int64(0)
	)

	t.Run("TestTransactionAuthVestingUnlock", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthVestingUnlock")

		resp := ta.VestingUnlock("mock pool id")
		require.NoError(t, resp)
	})
}

//RUNOK
func TestTransactionAuthVestingAdd(t *testing.T) {
	var (
		mockToClientID   = "2bba5b05949ea59c80aed3ac3474d7379d3be737e8eb5a968c52295e48333ead"
		mockTxnData      = `{"name":"add","input":{"description":"","start_time":0,"duration":0,"destinations":null}}`
		mockValue        = int64(1)
	)

	t.Run("TestTransactionAuthVestingAdd", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthVestingAdd")

		resp := ta.VestingAdd(&VestingAddRequest{}, 1)
		require.NoError(t, resp)
	})
}

//RUNOK
func TestTransactionAuthVestingDelete(t *testing.T) {
	var (
		mockToClientID   = "2bba5b05949ea59c80aed3ac3474d7379d3be737e8eb5a968c52295e48333ead"
		mockTxnData      = `{"name":"delete","input":{"pool_id":"mock pool id"}}`
		mockValue        = int64(0)
	)

	t.Run("TestTransactionAuthVestingDelete", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthVestingDelete")

		resp := ta.VestingDelete("mock pool id")
		require.NoError(t, resp)
	})
}

//RUNOK
func TestTransactionAuthVestingUpdateConfig(t *testing.T) {
	var (
		mockToClientID   = "2bba5b05949ea59c80aed3ac3474d7379d3be737e8eb5a968c52295e48333ead"
		mockTxnData      = `{"name":"update_config","input":{"min_lock":0,"min_duration":0,"max_duration":0,"max_destinations":0,"max_description_length":0}}`
		mockValue        = int64(0)
	)

	t.Run("TestTransactionAuthVestingUpdateConfig", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthVestingUpdateConfig")

		resp := ta.VestingUpdateConfig(&VestingSCConfig{})
		require.NoError(t, resp)
	})
}

//RUNOK
func TestTransactionAuthMinerSCSettings(t *testing.T) {
	var (
		mockToClientID   = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d9"
		mockTxnData      = `{"name":"update_settings","input":{"simple_miner":null,"pending":null,"active":null,"deleting":null}}`
		mockValue        = int64(0)
	)

	t.Run("TestTransactionAuthMinerSCSettings", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthMinerSCSettings")

		resp := ta.MinerSCSettings(&MinerSCMinerInfo{})
		require.NoError(t, resp)
	})
}

//RUNOK
func TestTransactionAuthMinerSCLock(t *testing.T) {
	var (
		mockToClientID   = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d9"
		mockTxnData      = `{"name":"addToDelegatePool","input":{"id":"mock miner id"}}`
		mockValue        = int64(1)
	)

	t.Run("TestTransactionAuthMinerSCLock", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthMinerSCLock")

		resp := ta.MinerSCLock("mock miner id", 1)
		require.NoError(t, resp)
	})
}

//RUNOK
func TestTransactionAuthMienrSCUnlock(t *testing.T) {
	var (
		mockToClientID   = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d9"
		mockTxnData      = `{"name":"deleteFromDelegatePool","input":{"id":"mock node id","pool_id":"mock pool id"}}`
		mockValue        = int64(0)
	)

	t.Run("TestTransactionAuthMienrSCUnlock", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthMienrSCUnlock")

		resp := ta.MienrSCUnlock("mock node id", "mock pool id")
		require.NoError(t, resp)
	})
}

//RUNOK
func TestTransactionAuthLockTokens(t *testing.T) {
	var (
		mockToClientID   = "cf8d0df9bd8cc637a4ff4e792ffe3686da6220c45f0e1103baa609f3f1751ef4"
		mockTxnData      = `{"name":"lock","input":{"duration":"1h1m"}}`
		mockValue        = int64(1)
	)

	t.Run("TestTransactionAuthLockTokens", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthLockTokens")

		resp := ta.LockTokens(1, 1, 1)
		require.NoError(t, resp)
	})
}

//RUNOK
func TestTransactionAuthUnlockTokens(t *testing.T) {
	var (
		mockToClientID   = "cf8d0df9bd8cc637a4ff4e792ffe3686da6220c45f0e1103baa609f3f1751ef4"
		mockTxnData      = `{"name":"unlock","input":{"pool_id":"mock pool id"}}`
		mockValue        = int64(0)
	)

	t.Run("TestTransactionAuthUnlockTokens", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthUnlockTokens")

		resp := ta.UnlockTokens("mock pool id")
		require.NoError(t, resp)
	})
}

//RUNOK
func TestTransactionAuthCreateAllocation(t *testing.T) {
	var (
		mockToClientID   = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7"
		mockTxnData      = `{"name":"new_allocation_request","input":{"data_shards":0,"parity_shards":0,"size":0,"expiration_date":0,"owner_id":"","owner_public_key":"","preferred_blobbers":null,"read_price_range":{"min":0,"max":0},"write_price_range":{"min":0,"max":0},"max_challenge_completion_time":0}}`
		mockValue        = int64(1)
	)

	t.Run("TestTransactionAuthCreateAllocation", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthCreateAllocation")

		resp := ta.CreateAllocation(&CreateAllocationRequest{}, 1, 1)
		require.NoError(t, resp)
	})
}

//RUNOK
func TestTransactionAuthCreateReadPool(t *testing.T) {
	var (
		mockToClientID   = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7"
		mockTxnData      = `{"name":"new_read_pool","input":null}`
		mockValue        = int64(0)
	)

	t.Run("TestTransactionAuthCreateReadPool", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthCreateReadPool")

		resp := ta.CreateReadPool(1)
		require.NoError(t, resp)
	})
}

//RUNOK
func TestTransactionAuthReadPoolLock(t *testing.T) {
	var (
		mockToClientID   = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7"
		mockTxnData      = `{"name":"read_pool_lock","input":{"duration":1,"allocation_id":"mock allocation id","blobber_id":"mock blobber id"}}`
		mockValue        = int64(1)
	)

	t.Run("TestTransactionAuthReadPoolLock", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthReadPoolLock")

		resp := ta.ReadPoolLock("mock allocation id", "mock blobber id", 1, 1, 1)
		require.NoError(t, resp)
	})
}

//RUNOK
func TestTransactionAuthReadPoolUnlock(t *testing.T) {
	var (
		mockToClientID   = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7"
		mockTxnData      = `{"name":"read_pool_unlock","input":{"pool_id":"mock pool id"}}`
		mockValue        = int64(0)
	)

	t.Run("TestTransactionAuthReadPoolUnlock", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthReadPoolUnlock")

		resp := ta.ReadPoolUnlock("mock pool id", 1)
		require.NoError(t, resp)
	})
}

//RUNOK
func TestTransactionAuthStakePoolLock(t *testing.T) {
	var (
		mockToClientID   = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7"
		mockTxnData      = `{"name":"stake_pool_lock","input":{"blobber_id":"mock blobber id"}}`
		mockValue        = int64(1)
	)

	t.Run("TestTransactionAuthStakePoolLock", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthStakePoolLock")

		resp := ta.StakePoolLock("mock blobber id", 1, 1)
		require.NoError(t, resp)
	})
}

//RUNOK
func TestTransactionAuthStakePoolUnlock(t *testing.T) {
	var (
		mockToClientID   = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7"
		mockTxnData      = `{"name":"stake_pool_unlock","input":{"blobber_id":"mock blobber id","pool_id":"mock pool id"}}`
		mockValue        = int64(0)
	)

	t.Run("TestTransactionAuthStakePoolUnlock", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthStakePoolUnlock")

		resp := ta.StakePoolUnlock("mock blobber id", "mock pool id", 1)
		require.NoError(t, resp)
	})
}

//RUNOK
func TestTransactionAuthStakePoolPayInterests(t *testing.T) {
	var (
		mockToClientID   = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7"
		mockTxnData      = `{"name":"stake_pool_pay_interests","input":{"blobber_id":"mock blobber id"}}`
		mockValue        = int64(0)
	)

	t.Run("TestTransactionAuthStakePoolPayInterests", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthStakePoolPayInterests")

		resp := ta.StakePoolPayInterests("mock blobber id", 1)
		require.NoError(t, resp)
	})
}

//RUNOK
func TestTransactionAuthUpdateBlobberSettings(t *testing.T) {
	var (
		mockToClientID   = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7"
		mockTxnData      = `{"name":"update_blobber_settings","input":{"id":"","url":"","terms":{"read_price":0,"write_price":0,"min_lock_demand":0,"max_offer_duration":0,"challenge_completion_time":0},"capacity":0,"used":0,"last_health_check":0,"stake_pool_settings":{"delegate_wallet":"","min_stake":0,"max_stake":0,"num_delegates":0}}}`
		mockValue        = int64(0)
	)

	t.Run("TestTransactionAuthUpdateBlobberSettings", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthWritePoolLock")

		resp := ta.UpdateBlobberSettings(&Blobber{}, 1)
		require.NoError(t, resp)
	})
}

//RUNOK
func TestTransactionAuthWritePoolLock(t *testing.T) {
	var (
		mockToClientID   = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7"
		mockTxnData      = `{"name":"write_pool_lock","input":{"duration":1,"allocation_id":"mock allocation id","blobber_id":"mock blobber id"}}`
		mockValue        = int64(1)
	)
	
	t.Run("TestTransactionAuthWritePoolLock", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthWritePoolLock")

		resp := ta.WritePoolLock("mock allocation id", "mock blobber id", 1, 1, 1)
		require.NoError(t, resp)
	})
}

//RUNOK
func TestTransactionAuthWritePoolUnlock(t *testing.T) {
	var (
		mockToClientID   = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7"
		mockTxnData      = `{"name":"write_pool_unlock","input":{"pool_id":"mock pool id"}}`
		mockValue        = int64(0)	
	)
	
	t.Run("TestTransactionAuthWritePoolUnlock", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthWritePoolUnlock")
		resp := ta.WritePoolUnlock("mock pool id", 1)
		require.NoError(t, resp)
	})
	
}

//RUNOK
func TestTransactionAuthUpdateAllocation(t *testing.T) {
	var (
		mockToClientID   = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7"
		mockTxnData      = `{"name":"update_allocation_request","input":{"id":"mock pool id","size":1,"expiration_date":1}}`
		mockValue        = int64(1)	
	)
	
	t.Run("Test Transaction Auth Update Allocation", func(t *testing.T) {
		ta := setupTestZcncore(mockToClientID,mockTxnData,mockValue,"TestTransactionAuthUpdateAllocation")
		resp := ta.UpdateAllocation("mock pool id", 1, 1, 1, 1)
		require.NoError(t, resp)
	})
}
func setupSignatureSchemeMock(ss *zcncryptomock.SignatureScheme) {
	zcncrypto.NewSignatureScheme = func(sigScheme string) zcncrypto.SignatureScheme {
		return ss
	}
}
func setupTestZcncore(toClientID string, txnData string, value int64, testFuncName string) *TransactionWithAuth {
	var (
		mockPublicKey    = "mock public key"
		mockPrivateKey   = "mock private key"
		mockSignature    = "mock signature"
		mockClientID     = "mock client id"
		mockToClientID   = toClientID
		mockTxnData      = txnData
		mockCreationDate = int64(1625030157)
		mockValue        = value
		mockTxn          = &transaction.Transaction{
			PublicKey:       mockPublicKey,
			ClientID:        mockClientID,
			TransactionData: mockTxnData,
			CreationDate:    mockCreationDate,
			ToClientID:      mockToClientID,
			Value:           mockValue,
			Signature:       mockSignature,
			TransactionType: transaction.TxnTypeData,
		}
	)
	mockTxn.ComputeHashData()
	_config.wallet = zcncrypto.Wallet{
		ClientID: mockClientID,
		Keys: []zcncrypto.KeyPair{
			{
				PublicKey:  mockPublicKey,
				PrivateKey: mockPrivateKey,
			},
		},
	}
	_config.chain.SignatureScheme = "bls0chain"
	_config.authUrl = testFuncName
	var mockClient = zcnmocks.HttpClient{}
	util.Client = &mockClient

	mockSignatureScheme := &zcncryptomock.SignatureScheme{}
	mockSignatureScheme.On("SetPrivateKey", mockPrivateKey).Return(nil)
	mockSignatureScheme.On("SetPublicKey", mockPublicKey).Return(nil)
	mockSignatureScheme.On("Sign", mockTxn.Hash).Return(mockSignature, nil)
	mockSignatureScheme.On("Verify", mockSignature, mockTxn.Hash).Return(true, nil)
	mockSignatureScheme.On("Add", mockTxn.Signature, mockTxn.Hash).Return(mockSignature, nil)
	setupSignatureSchemeMock(mockSignatureScheme)

	mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		if req.Method == "POST" && req.URL.Path == _config.authUrl+"/transaction" {
			return true
		}
		return false
	})).Return(&http.Response{
		Body: func() io.ReadCloser {
			jsonFR, _ := json.Marshal(mockTxn)
			return ioutil.NopCloser(bytes.NewReader(jsonFR))
		}(),
		StatusCode: http.StatusOK,
	}, nil)

	_config.chain.Miners = []string{testFuncName, ""}

	mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		if strings.HasPrefix(req.URL.Path, "/dns") || strings.HasPrefix(req.URL.Path,testFuncName) || strings.HasPrefix(req.URL.Path, "/v1/") {
			return true
		}
		return false
	})).Return(&http.Response{
		Body: func() io.ReadCloser {
			jsonFR, _ := json.Marshal(mockTxn)
			return ioutil.NopCloser(bytes.NewReader(jsonFR))
		}(),
		StatusCode: http.StatusOK,
	}, nil)

	ta := &TransactionWithAuth{
		t: &Transaction{
			txn: &transaction.Transaction{
				Hash:         mockTxn.Hash,
				ClientID:     mockClientID,
				PublicKey:    mockPublicKey,
				ToClientID:   mockClientID,
				CreationDate: mockCreationDate,
				Value:        mockValue,
			},
		},
	}
	return ta
}
