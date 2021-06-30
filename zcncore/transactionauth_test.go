package zcncore

import (
	"bytes"
	"encoding/json"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/zcncrypto"
	zcncryptomock "github.com/0chain/gosdk/core/zcncrypto/mocks"
	"github.com/0chain/gosdk/zcnmocks"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

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

func TestSign(t *testing.T) {
	t.Run("Test Sign", func(t *testing.T) {
		_config.chain.SignatureScheme = "bls0chain"
		_config.authUrl = "TestSign"
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}
		var mockClient = zcnmocks.HttpClient{}
		util.Client = &mockClient
		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return strings.HasPrefix(req.URL.Path, "TestSign")
		})).Return(&http.Response{
			Body: func() io.ReadCloser {
				jsonFR, err := json.Marshal(&transaction.Transaction{
					// Hash: mockHash,
					// Signature: "ed25519",
				})
				require.NoError(t, err)
				return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
			}(),
			StatusCode: http.StatusOK,
		}, nil)
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
		mockTxn := &transaction.Transaction{
			PublicKey: mockPublicKey,
		}
		mockTxn.ComputeHashData()
		_config.chain.SignatureScheme = "bls0chain"

		sig := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
		sig.SetPrivateKey(_config.wallet.Keys[0].PrivateKey)
		var err error

		mockTxn.Signature, err = sig.Sign(mockTxn.Hash)

		_config.authUrl = "TestSend"
		var mockClient = zcnmocks.HttpClient{}
		util.Client = &mockClient
		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return strings.HasPrefix(req.URL.Path, "TestSend")
		})).Return(&http.Response{
			Body: func() io.ReadCloser {
				jsonFR, err := json.Marshal(mockTxn)
				require.NoError(t, err)
				return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
			}(),
			StatusCode: http.StatusOK,
		}, nil)
		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return strings.HasPrefix(req.URL.Path, "TestSend")
		})).Return(&http.Response{
			Body: func() io.ReadCloser {
				jsonFR, err := json.Marshal(mockTxn)
				require.NoError(t, err)
				return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
			}(),
			StatusCode: http.StatusOK,
		}, nil)
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{},
			},
		}

		err = ta.Send(clientID, 1, "desc")
		require.NoError(t, err)
	})
}

func TestStoreData(t *testing.T) {
	var (
		mockPublicKey    = "mock public key"
		mockPrivateKey   = "mock private key"
		mockSignature    = "mock signature"
		mockClientID     = "mock client id"
		mockTxnData      = "mock txn data"
		mockCreationDate = int64(1625030157)
		mockValue        = int64(1)
		mockTxn          = &transaction.Transaction{
			PublicKey:       mockPublicKey,
			ClientID:        mockClientID,
			TransactionData: mockTxnData,
			CreationDate:    mockCreationDate,
			ToClientID:      mockClientID,
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
	_config.authUrl = "TestStoreData"

	t.Run("Test Store Data", func(t *testing.T) {
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
				require.EqualValues(t, "application/json", strings.Split(req.Header.Get("Content-Type"), ";")[0])
				defer req.Body.Close()
				body, err := ioutil.ReadAll(req.Body)
				require.NoError(t, err, "ioutil.ReadAll(req.Body)")
				var reqTxn *transaction.Transaction
				err = json.Unmarshal(body, &reqTxn)
				require.NoError(t, err, "json.Unmarshal(body, &reqTxn)")
				require.EqualValues(t, mockTxn, reqTxn)
				return true
			}
			return false
		})).Return(&http.Response{
			Body: func() io.ReadCloser {
				jsonFR, err := json.Marshal(mockTxn)
				require.NoError(t, err, "json.Marshal(mockTxn)")
				return ioutil.NopCloser(bytes.NewReader(jsonFR))
			}(),
			StatusCode: http.StatusOK,
		}, nil)
		//mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		//	return strings.HasPrefix(req.URL.Path, "TestStoreData")
		//})).Return(&http.Response{
		//	Body: func() io.ReadCloser {
		//		jsonFR, err := json.Marshal(mockTxn)
		//		require.NoError(t, err)
		//		return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
		//	}(),
		//	StatusCode: http.StatusOK,
		//}, nil)
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

		err := ta.StoreData(mockTxnData)
		require.NoError(t, err)
	})
}

// func TestExecuteFaucetSCWallet(t *testing.T) {
// 	t.Run("Test Execute Faucet SC Wallet", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		err := ta.ExecuteFaucetSCWallet(walletString, "get", []byte("test"))
// 		require.NoError(t, err)
// 	})
// }
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

// func TestTransactionAuthGetTransactionHash(t *testing.T) {
// 	t.Run("Test Get Transaction Hash", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.GetTransactionHash()
// 		require.NotNil(t, resp)
// 	})
// }
// func TestTransactionAuthVerify(t *testing.T) {
// 	t.Run("Test Transaction Auth Verify", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		err := ta.Verify()
// 		expectedErrorMsg := "invalid transaction. cannot be verified."
// 		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
// 	})
// }
// func TestTransactionAuthGetVerifyOutput(t *testing.T) {
// 	t.Run("Test Transaction Auth Get Verify Output", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.GetVerifyOutput()
// 		require.NotNil(t, resp)
// 	})
// }
// func TestTransactionAuthGetTransactionError(t *testing.T) {
// 	t.Run("Test Transaction Auth Get Transaction Error", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.GetTransactionError()
// 		require.NotNil(t, resp)
// 	})
// }

// func TestTransactionAuthGetVerifyError(t *testing.T) {
// 	t.Run("Test Transaction Auth Get Verify Error", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.GetVerifyError()
// 		require.NotNil(t, resp)
// 	})
// }
// func TestTransactionAuthOutput(t *testing.T) {
// 	t.Run("Test Transaction Auth Output", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.Output()
// 		require.NotNil(t, resp)
// 	})
// }
// func TestTransactionAuthVestingTrigger(t *testing.T) {
// 	t.Run("Test Transaction Auth Vesting Trigger", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.VestingTrigger("poolID")
// 		require.NoError(t, resp)
// 	})
// }
// func TestTransactionAuthVestingStop(t *testing.T) {
// 	t.Run("Test Transaction Auth Vesting Stop", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.VestingStop(&VestingStopRequest{})
// 		require.NoError(t, resp)
// 	})
// }

// func TestTransactionAuthVestingUnlock(t *testing.T) {
// 	t.Run("Test Transaction Auth Vesting Unlock", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.VestingUnlock("poolID")
// 		require.NoError(t, resp)
// 	})
// }

// func TestTransactionAuthVestingAdd(t *testing.T) {
// 	t.Run("Test Transaction Auth Vesting Add", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.VestingAdd(&VestingAddRequest{}, 1)
// 		require.NoError(t, resp)
// 	})
// }
// func TestTransactionAuthVestingDelete(t *testing.T) {
// 	t.Run("Test Transaction Auth Vesting Delete", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.VestingDelete("poolID")
// 		require.NoError(t, resp)
// 	})
// }
// func TestTransactionAuthVestingUpdateConfig(t *testing.T) {
// 	t.Run("Test Transaction Auth Vesting Update Config", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.VestingUpdateConfig(&VestingSCConfig{})
// 		require.NoError(t, resp)
// 	})
// }

// func TestTransactionAuthMinerSCSettings(t *testing.T) {
// 	t.Run("Test Transaction Auth Miner SC Settings", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.MinerSCSettings(&MinerSCMinerInfo{})
// 		require.NoError(t, resp)
// 	})
// }
// func TestTransactionAuthMinerSCLock(t *testing.T) {
// 	t.Run("Test Transaction Auth Miner SC Lock", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.MinerSCLock("minerID", 1)
// 		require.NoError(t, resp)
// 	})
// }
// func TestTransactionAuthMienrSCUnlock(t *testing.T) {
// 	t.Run("Test Transaction Auth Miner SC Lock", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.MienrSCUnlock("nodeID", "poolID")
// 		require.NoError(t, resp)
// 	})
// }
// func TestTransactionAuthLockTokens(t *testing.T) {
// 	t.Run("Test Transaction Auth Lock Tokens", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.LockTokens(1, 1, 1)
// 		require.NoError(t, resp)
// 	})
// }
// func TestTransactionAuthUnlockTokens(t *testing.T) {
// 	t.Run("Test Transaction Auth Unlock Tokens", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.UnlockTokens("poolID")
// 		require.NoError(t, resp)
// 	})
// }
// func TestTransactionAuthRegisterMultiSig(t *testing.T) {
// 	t.Run("Test Transaction Auth Register MultiSig", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.RegisterMultiSig(walletString, msw)
// 		expectedErrorMsg := "not implemented"
// 		assert.EqualErrorf(t, resp, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, resp)
// 	})
// }
// func TestTransactionAuthFinalizeAllocation(t *testing.T) {
// 	t.Run("Test Transaction Auth Finalize Allocation", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.FinalizeAllocation("poolID", 1)
// 		require.NoError(t, resp)
// 	})
// }
// func TestTransactionAuthCancelAllocation(t *testing.T) {
// 	t.Run("Test Transaction Auth Cancel Allocation", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.CancelAllocation("alloc string", 1)
// 		require.NoError(t, resp)
// 	})
// }
// func TestTransactionAuthCreateAllocation(t *testing.T) {
// 	t.Run("Test Transaction Auth Cancel Allocation", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.CreateAllocation(&CreateAllocationRequest{}, 1, 1)
// 		require.NoError(t, resp)
// 	})
// }
// func TestTransactionAuthCreateReadPool(t *testing.T) {
// 	t.Run("Test Transaction Auth Create ReadPool", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.CreateReadPool(1)
// 		require.NoError(t, resp)
// 	})
// }
// func TestTransactionAuthReadPoolLock(t *testing.T) {
// 	t.Run("Test Transaction Auth Create ReadPool", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.ReadPoolLock("allocID", "blobberID", 1, 1, 1)
// 		require.NoError(t, resp)
// 	})
// }
// func TestTransactionAuthReadPoolUnlock(t *testing.T) {
// 	t.Run("Test Transaction Auth Read Pool Unlock", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.ReadPoolUnlock("poolID", 1)
// 		require.NoError(t, resp)
// 	})
// }
// func TestTransactionAuthStakePoolLock(t *testing.T) {
// 	t.Run("Test Transaction Auth Stake Pool Lock", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.StakePoolLock("blobberID", 1, 1)
// 		require.NoError(t, resp)
// 	})
// }
// func TestTransactionAuthStakePoolUnlock(t *testing.T) {
// 	t.Run("Test Transaction Auth Stake Pool Lock", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.StakePoolUnlock("blobberID", "poolID", 1)
// 		require.NoError(t, resp)
// 	})
// }
// func TestTransactionAuthStakePoolPayInterests(t *testing.T) {
// 	t.Run("Test Transaction Auth Stake Pool Lock", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.StakePoolPayInterests("blobberID", 1)
// 		require.NoError(t, resp)
// 	})
// }
// func TestTransactionAuthUpdateBlobberSettings(t *testing.T) {
// 	t.Run("Test Transaction Auth Stake Pool Lock", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.UpdateBlobberSettings(&Blobber{}, 1)
// 		require.NoError(t, resp)
// 	})
// }
// func TestTransactionAuthUpdateAllocation(t *testing.T) {
// 	t.Run("Test Transaction Auth Update Allocation", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.UpdateAllocation("allocID", 1, 1, 1, 1)
// 		require.NoError(t, resp)
// 	})
// }
// func TestTransactionAuthWritePoolLock(t *testing.T) {
// 	t.Run("Test Transaction Auth Write Pool Lock", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		resp := ta.WritePoolLock("allocID", "blobberID", 1, 1, 1)
// 		require.NoError(t, resp)
// 	})
// }

// func TestTransactionAuthWritePoolUnlock(t *testing.T) {
// 	t.Run("Test Transaction Auth Write Pool Lock", func(t *testing.T) {
// 		var mockClient = zcnmocks.HttpClient{}
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
// 		_config.authUrl = "TestTransactionAuthWritePoolUnlock"

// 		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
// 			return strings.HasPrefix(req.URL.Path, "/TestTransactionAuthWritePoolUnlock")
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
// 			return strings.HasPrefix(req.URL.Path, "/TestTransactionAuthWritePoolUnlock")
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
func setupSignatureSchemeMock(ss *zcncryptomock.SignatureScheme) {
	zcncrypto.NewSignatureScheme = func(sigScheme string) zcncrypto.SignatureScheme {
		return ss
	}
}
