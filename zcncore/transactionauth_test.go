package zcncore

import (
	"testing"

	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// func TestNewTransactionWithAuth(t *testing.T) {
// 	t.Run("Test New Transaction With Auth Success", func(t *testing.T) {
// 		mockWalletCallback := mocks.TransactionCallbackImpl{}
// 		mockWalletCallback.On("OnTransactionComplete", &Transaction{}, 0).Return()
// 		resp, err := newTransactionWithAuth(mockWalletCallback, 1)
// 		require.NotEmpty(t, resp)
// 		// expectedErrorMsg := "magic block info not found"
// 		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)\
// 		require.NoError(t, err)
// 	})
// }

// func TestTransactionAuthSetTransactionCallback(t *testing.T) {
// 	t.Run("Test New Transaction With Auth transaction already exists", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{},
// 		}
// 		mockWalletCallback := TransactionCallbackImpl{}
// 		mockWalletCallback.On("OnTransactionComplete", &Transaction{}, 0).Return()
// 		err := ta.SetTransactionCallback(mockWalletCallback)
// 		expectedErrorMsg := "transaction already exists. cannot set transaction hash."
// 		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
// 		// require.NoError(t, err)
// 	})
// 	t.Run("Test New Transaction With Auth success", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txnStatus: -1,
// 			},
// 		}
// 		mockWalletCallback := TransactionCallbackImpl{}
// 		mockWalletCallback.On("OnTransactionComplete", &Transaction{}, 0).Return()
// 		err := ta.SetTransactionCallback(mockWalletCallback)

// 		require.NoError(t, err)
// 	})
// }
// func TestTransactionAuthSetTransactionFee(t *testing.T) {
// 	t.Run("Test Transaction Auth Set Transaction Fee", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{},
// 		}

// 		err := ta.SetTransactionFee(1)
// 		expectedErrorMsg := "transaction already exists. cannot set transaction fee."
// 		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
// 		// require.NoError(t, err)
// 	})
// 	t.Run("Test Transaction Auth Set Transaction Fee", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txnStatus: -1,
// 				txn:       &transaction.Transaction{},
// 			},
// 		}

// 		err := ta.SetTransactionFee(1)

// 		require.NoError(t, err)
// 	})
// }

// func TestVerifyFn(t *testing.T) {
// 	t.Run("Test Verify Fn", func(t *testing.T) {
// 		resp, err := verifyFn(mnemonic, hash, public_key)
// 		// expectedErrorMsg := "transaction already exists. cannot set transaction fee."
// 		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
// 		require.NoError(t, err)
// 		require.NotEmpty(t, resp)
// 	})

// }
func TestSign(t *testing.T) {
	t.Run("Test Sign", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}
		_config.wallet = zcncrypto.Wallet{
			Keys: []zcncrypto.KeyPair{
				zcncrypto.KeyPair{
					PublicKey:  public_key,
					PrivateKey: private_key,
				},
			},
		}

		err := ta.sign("ortherSig")
		expectedErrorMsg := "odd length"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}

func TestSend(t *testing.T) {
	t.Run("Test Send", func(t *testing.T) {
		_config.wallet = zcncrypto.Wallet{
			Keys: []zcncrypto.KeyPair{
				zcncrypto.KeyPair{
					PublicKey:  public_key,
					PrivateKey: private_key,
				},
			},
		}
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		err := ta.Send(clientID, 1, "desc")
		require.NoError(t, err)
	})
}
func TestStoreData(t *testing.T) {
	t.Run("Test Store Data", func(t *testing.T) {
		_config.wallet = zcncrypto.Wallet{
			Keys: []zcncrypto.KeyPair{
				zcncrypto.KeyPair{
					PublicKey:  public_key,
					PrivateKey: private_key,
				},
			},
		}
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		err := ta.StoreData("data")
		require.NoError(t, err)
	})
}
func TestExecuteFaucetSCWallet(t *testing.T) {
	t.Run("Test Execute Faucet SC Wallet", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{},
		}

		err := ta.ExecuteFaucetSCWallet("data", "", []byte("test"))
		expectedErrorMsg := "invalid character 'd' looking for beginning of value"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestExecuteSmartContract(t *testing.T) {
	t.Run("Test Execute Smart Contract", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		err := ta.ExecuteSmartContract("address", "GET", "{}", 1)
		require.NoError(t, err)
	})
}
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
func TestTransactionAuthVestingTrigger(t *testing.T) {
	t.Run("Test Transaction Auth Vesting Trigger", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.VestingTrigger("poolID")
		require.NoError(t, resp)
	})
}
func TestTransactionAuthVestingStop(t *testing.T) {
	t.Run("Test Transaction Auth Vesting Stop", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.VestingStop(&VestingStopRequest{})
		require.NoError(t, resp)
	})
}

func TestTransactionAuthVestingUnlock(t *testing.T) {
	t.Run("Test Transaction Auth Vesting Unlock", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.VestingUnlock("poolID")
		require.NoError(t, resp)
	})
}

func TestTransactionAuthVestingAdd(t *testing.T) {
	t.Run("Test Transaction Auth Vesting Add", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.VestingAdd(&VestingAddRequest{}, 1)
		require.NoError(t, resp)
	})
}
func TestTransactionAuthVestingDelete(t *testing.T) {
	t.Run("Test Transaction Auth Vesting Delete", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.VestingDelete("poolID")
		require.NoError(t, resp)
	})
}
func TestTransactionAuthVestingUpdateConfig(t *testing.T) {
	t.Run("Test Transaction Auth Vesting Update Config", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.VestingUpdateConfig(&VestingSCConfig{})
		require.NoError(t, resp)
	})
}

func TestTransactionAuthMinerSCSettings(t *testing.T) {
	t.Run("Test Transaction Auth Miner SC Settings", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.MinerSCSettings(&MinerSCMinerInfo{})
		require.NoError(t, resp)
	})
}
func TestTransactionAuthMinerSCLock(t *testing.T) {
	t.Run("Test Transaction Auth Miner SC Lock", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.MinerSCLock("minerID", 1)
		require.NoError(t, resp)
	})
}
func TestTransactionAuthMienrSCUnlock(t *testing.T) {
	t.Run("Test Transaction Auth Miner SC Lock", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.MienrSCUnlock("nodeID", "poolID")
		require.NoError(t, resp)
	})
}
func TestTransactionAuthLockTokens(t *testing.T) {
	t.Run("Test Transaction Auth Lock Tokens", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.LockTokens(1, 1, 1)
		require.NoError(t, resp)
	})
}
func TestTransactionAuthUnlockTokens(t *testing.T) {
	t.Run("Test Transaction Auth Unlock Tokens", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.UnlockTokens("poolID")
		require.NoError(t, resp)
	})
}
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
func TestTransactionAuthFinalizeAllocation(t *testing.T) {
	t.Run("Test Transaction Auth Finalize Allocation", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.FinalizeAllocation("poolID", 1)
		require.NoError(t, resp)
	})
}
func TestTransactionAuthCancelAllocation(t *testing.T) {
	t.Run("Test Transaction Auth Cancel Allocation", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.CancelAllocation("alloc string", 1)
		require.NoError(t, resp)
	})
}
func TestTransactionAuthCreateAllocation(t *testing.T) {
	t.Run("Test Transaction Auth Cancel Allocation", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.CreateAllocation(&CreateAllocationRequest{}, 1, 1)
		require.NoError(t, resp)
	})
}
func TestTransactionAuthCreateReadPool(t *testing.T) {
	t.Run("Test Transaction Auth Create ReadPool", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.CreateReadPool(1)
		require.NoError(t, resp)
	})
}
func TestTransactionAuthReadPoolLock(t *testing.T) {
	t.Run("Test Transaction Auth Create ReadPool", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.ReadPoolLock("allocID", "blobberID", 1, 1, 1)
		require.NoError(t, resp)
	})
}
func TestTransactionAuthReadPoolUnlock(t *testing.T) {
	t.Run("Test Transaction Auth Read Pool Unlock", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.ReadPoolUnlock("poolID", 1)
		require.NoError(t, resp)
	})
}
func TestTransactionAuthStakePoolLock(t *testing.T) {
	t.Run("Test Transaction Auth Stake Pool Lock", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.StakePoolLock("blobberID", 1, 1)
		require.NoError(t, resp)
	})
}
func TestTransactionAuthStakePoolUnlock(t *testing.T) {
	t.Run("Test Transaction Auth Stake Pool Lock", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.StakePoolUnlock("blobberID", "poolID", 1)
		require.NoError(t, resp)
	})
}
func TestTransactionAuthStakePoolPayInterests(t *testing.T) {
	t.Run("Test Transaction Auth Stake Pool Lock", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.StakePoolPayInterests("blobberID", 1)
		require.NoError(t, resp)
	})
}
func TestTransactionAuthUpdateBlobberSettings(t *testing.T) {
	t.Run("Test Transaction Auth Stake Pool Lock", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.UpdateBlobberSettings(&Blobber{}, 1)
		require.NoError(t, resp)
	})
}
func TestTransactionAuthUpdateAllocation(t *testing.T) {
	t.Run("Test Transaction Auth Update Allocation", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.UpdateAllocation("allocID", 1, 1, 1, 1)
		require.NoError(t, resp)
	})
}
func TestTransactionAuthWritePoolLock(t *testing.T) {
	t.Run("Test Transaction Auth Write Pool Lock", func(t *testing.T) {
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		resp := ta.WritePoolLock("allocID", "blobberID", 1, 1, 1)
		require.NoError(t, resp)
	})
}

// func TestTransactionAuthWritePoolUnlock(t *testing.T) {
// 	t.Run("Test Transaction Auth Write Pool Lock", func(t *testing.T) {
// 		var mockClient = mocks.HttpClient{}
// 		util.Client = &mockClient
// 		_config.wallet = zcncrypto.Wallet{
// 			Keys: []zcncrypto.KeyPair{
// 				zcncrypto.KeyPair{
// 					PublicKey:  public_key,
// 					PrivateKey: private_key,
// 				},
// 			},
// 		}
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}
// 		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
// 			return strings.HasPrefix(req.URL.Path, "/transaction")
// 		})).Return(&http.Response{
// 			Body: func() io.ReadCloser {
// 				jsonFR, err := json.Marshal(&transaction.Transaction{
// 					Hash:      mockHash,
// 					Signature: "bls0chain",
// 				})
// 				require.NoError(t, err)
// 				return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
// 			}(),
// 			StatusCode: http.StatusOK,
// 		}, nil)
// 		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
// 			return strings.HasPrefix(req.URL.Path, "/transaction")
// 		})).Return(&http.Response{
// 			Body: func() io.ReadCloser {
// 				jsonFR, err := json.Marshal(&transaction.Transaction{
// 					Hash:      mockHash,
// 					Signature: "bls0chain",
// 				})
// 				require.NoError(t, err)
// 				return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
// 			}(),
// 			StatusCode: http.StatusOK,
// 		}, nil)
// 		resp := ta.WritePoolUnlock("poolID", 1)
// 		require.NoError(t, resp)
// 	})
// }

// func TestTransactionAuthGetAuthorize(t *testing.T) {
// 	t.Run("Test Transaction Auth get Authorize", func(t *testing.T) {
// 		_config.wallet = zcncrypto.Wallet{
// 			Keys: []zcncrypto.KeyPair{
// 				zcncrypto.KeyPair{
// 					PublicKey:  public_key,
// 					PrivateKey: private_key,
// 				},
// 			},
// 		}
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		transaction, err := ta.getAuthorize()
// 		require.Nil(t, transaction)
// 		expectedErrorMsg := "network error. host not reachable"
// 		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
// 	})
// }
