package zcncore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	mockHash = "15743192c9f78cf56824f83d92d54e48f50ca53c305a316ad7070b9ba4fac486"
)

func TestVestingTrigger(t *testing.T) {
	t.Run("Test Vesting Trigger Success", func(t *testing.T) {
		var mockClient = mocks.HttpClient{}
		util.Client = &mockClient
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown
		_config.wallet = zcncrypto.Wallet{
			Keys: []zcncrypto.KeyPair{
				zcncrypto.KeyPair{
					PublicKey:  public_key,
					PrivateKey: private_key,
				},
			},
		}
		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			fmt.Println("----------", req.URL.Path)
			return strings.HasPrefix(req.URL.Path, "")
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
		err := trans.VestingTrigger("poolID")

		require.Nil(t, err)
	})
}

// func TestGetBlockInfoByRound(t *testing.T) {
// 	t.Run("Test Vesting Trigger Success", func(t *testing.T) {
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
// 		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
// 			return true
// 		})).Return(&http.Response{
// 			Body: func() io.ReadCloser {
// 				jsonFR, err := json.Marshal(&blockHeader{
// 					// Hash: mockHash,
// 					// Signature: "ed25519",
// 				})
// 				require.NoError(t, err)
// 				return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
// 			}(),
// 			StatusCode: http.StatusOK,
// 		}, nil)
// 		resp, err := getBlockInfoByRound(1, 1, "test")
// 		require.Nil(t, resp)
// 		require.NotEmpty(t, err)
// 	})
// }

// func TestValidateChain(t *testing.T) {
// 	t.Run("Test validate Chain Success", func(t *testing.T) {
// 		_config.wallet = zcncrypto.Wallet{
// 			Keys: []zcncrypto.KeyPair{
// 				zcncrypto.KeyPair{
// 					PublicKey:  public_key,
// 					PrivateKey: private_key,
// 				},
// 			},
// 		}
// 		resp := validateChain(&blockHeader{
// 			Version : "",
// 			CreationDate : 1,
// 			Hash : "",
// 			MinerId : "",
// 			Round : 1,
// 			RoundRandomSeed : 1,
// 			MerkleTreeRoot : "",
// 			StateHash : "",
// 			ReceiptMerkleTreeRoot : "",
// 			NumTxns : 1,
// 		})
// 		require.Nil(t, resp)
// 		// require.NotEmpty(t, err)
// 	})
// }
func TestIsTransactionExpired(t *testing.T) {
	t.Run("Test Is Transaction Expired False", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown
		_config.wallet = zcncrypto.Wallet{
			Keys: []zcncrypto.KeyPair{
				zcncrypto.KeyPair{
					PublicKey:  public_key,
					PrivateKey: private_key,
				},
			},
		}
		resp := trans.isTransactionExpired(1, 1)

		require.False(t, resp)
	})
	t.Run("Test Is Transaction Expired False", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown
		_config.wallet = zcncrypto.Wallet{
			Keys: []zcncrypto.KeyPair{
				zcncrypto.KeyPair{
					PublicKey:  public_key,
					PrivateKey: private_key,
				},
			},
		}
		resp := trans.isTransactionExpired(0, 1)

		require.False(t, resp)
	})
	t.Run("Test Is Transaction Expired False", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown
		_config.wallet = zcncrypto.Wallet{
			Keys: []zcncrypto.KeyPair{
				zcncrypto.KeyPair{
					PublicKey:  public_key,
					PrivateKey: private_key,
				},
			},
		}
		resp := trans.isTransactionExpired(100, 100)

		require.True(t, resp)
	})
}
func TestSignFn(t *testing.T) {
	t.Run("SignFn Success", func(t *testing.T) {
		_config.chain.SignatureScheme = "bls0chain"
		_config.wallet = zcncrypto.Wallet{
			Keys: []zcncrypto.KeyPair{
				zcncrypto.KeyPair{
					PublicKey:  public_key,
					PrivateKey: private_key,
				},
			},
		}
		resp, err := signFn("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
		require.NoError(t, err)
		require.NotNil(t, resp)
	})
}

// func TestSignWithWallet(t *testing.T) {
// 	t.Run("Sign With Wallet Success", func(t *testing.T) {
// 		_config.chain.SignatureScheme = "bls0chain"
// 		_config.wallet = zcncrypto.Wallet{
// 			Keys: []zcncrypto.KeyPair{
// 				zcncrypto.KeyPair{
// 					PublicKey : public_key,
// 					PrivateKey : private_key,
// 				},

// 			},
// 		}
// 		resp ,err := signWithWallet("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",_config.wallet)
// 		require.NoError(t, err)
// 		require.NotNil(t,resp)
// 	})
// }
func TestTxnTypeString(t *testing.T) {
	t.Run("Txn Type String send", func(t *testing.T) {

		resp := txnTypeString(transaction.TxnTypeSend)
		require.Equal(t, "send", resp)
	})
	t.Run("Txn Type String lock-in", func(t *testing.T) {

		resp := txnTypeString(transaction.TxnTypeLockIn)
		require.Equal(t, "lock-in", resp)
	})
	t.Run("Txn Type String data", func(t *testing.T) {

		resp := txnTypeString(transaction.TxnTypeData)
		require.Equal(t, "data", resp)
	})
	t.Run("Txn Type String smart contract", func(t *testing.T) {

		resp := txnTypeString(transaction.TxnTypeSmartContract)
		require.Equal(t, "smart contract", resp)
	})
	t.Run("Txn Type String unknown", func(t *testing.T) {

		resp := txnTypeString(123)
		require.Equal(t, "unknown", resp)
	})
}
func TestOutput(t *testing.T) {
	t.Run("Output Success", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
		}
		resp := trans.Output()
		// require.NoError(t, err)
		require.Equal(t, []byte(trans.txnOut), resp)
	})
}
func TestCompleteTxn(t *testing.T) {
	t.Run("Complete Txn Success", func(t *testing.T) {

		trans := &Transaction{}
		trans.completeTxn(1, "abc", nil)
		expected := &Transaction{
			txnStatus: 1,
			txnOut:    "abc",
			txnError:  nil,
		}

		require.EqualValues(t, trans, expected)
	})
}
func TestCompleteVerify(t *testing.T) {
	t.Run("Complete Verify Success", func(t *testing.T) {

		trans := &Transaction{}
		trans.completeVerify(1, "abc", nil)
		expected := &Transaction{
			verifyStatus: 1,
			verifyOut:    "abc",
			verifyError:  nil,
		}

		require.EqualValues(t, trans, expected)
	})
}

// func TestNewTransaction(t *testing.T) {
// 	t.Run("New Transaction Success", func(t *testing.T) {

// 		mockWalletCallback := TransactionCallbackImpl{}
// 		mockWalletCallback.On("OnTransactionComplete", &Transaction{}, 0).Return()
// 		trans, err := newTransaction(mockWalletCallback, 100)
// 		expected := &Transaction{
// 			txn: &transaction.Transaction{},
// 		}
// 		expected.txn.TransactionFee = 100
// 		require.EqualValues(t, trans.txn.TransactionFee, expected.txn.TransactionFee)
// 		require.NoError(t, err)
// 	})
// }
// func TestNewTransactionFunction(t *testing.T) {
// 	t.Run("New Transaction wallet info not found", func(t *testing.T) {

// 		_config.chain.Sharders = []string{"3", "4"}
// 		_config.isValidWallet = false
// 		_config.wallet.ClientID = "test"

// 		mockWalletCallback := TransactionCallbackImpl{}
// 		mockWalletCallback.On("OnTransactionComplete", &Transaction{}, 0).Return()

// 		resp, err := NewTransaction(mockWalletCallback, 100)

// 		expectedErrorMsg := "wallet info not found. set wallet info."
// 		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
// 		require.Nil(t, resp)
// 	})
// 	t.Run("New Transaction check config Success", func(t *testing.T) {

// 		_config.chain.Sharders = []string{"3", "4"}
// 		_config.isValidWallet = true
// 		_config.wallet.ClientID = "test"
// 		_config.isConfigured = true
// 		_config.chain.Miners = []string{"0", "1"}
// 		_config.chain.Sharders = []string{"0", "1"}

// 		mockWalletCallback := TransactionCallbackImpl{}
// 		mockWalletCallback.On("OnTransactionComplete", &Transaction{}, 0).Return()

// 		resp, err := NewTransaction(mockWalletCallback, 100)

// 		// expectedErrorMsg := "SDK not initialized"
// 		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
// 		require.NotNil(t, resp)
// 		require.NoError(t, err)
// 	})
// 	t.Run("New Transaction check config Success", func(t *testing.T) {

// 		_config.chain.Sharders = []string{"3", "4"}
// 		_config.isValidWallet = true
// 		_config.wallet.ClientID = "test"
// 		_config.isConfigured = true
// 		_config.chain.Miners = []string{"0", "1"}
// 		_config.chain.Sharders = []string{"0", "1"}
// 		_config.isSplitWallet = true
// 		_config.authUrl = "auth url"
// 		mockWalletCallback := TransactionCallbackImpl{}

// 		mockWalletCallback.On("OnTransactionComplete", &Transaction{}, 0).Return()

// 		resp, err := NewTransaction(mockWalletCallback, 100)

// 		// expectedErrorMsg := "SDK not initialized"
// 		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
// 		require.NotEmpty(t, resp)
// 		require.NoError(t, err)
// 	})
// 	t.Run("New Transaction check config fails", func(t *testing.T) {

// 		_config.chain.Sharders = []string{"3", "4"}
// 		_config.isValidWallet = true
// 		_config.wallet.ClientID = "test"
// 		_config.isConfigured = true
// 		_config.chain.Miners = []string{"0", "1"}
// 		_config.chain.Sharders = []string{"0", "1"}
// 		_config.isSplitWallet = true
// 		// _config.authUrl = "auth url"
// 		mockWalletCallback := TransactionCallbackImpl{}

// 		mockWalletCallback.On("OnTransactionComplete", &Transaction{}, 0).Return()

// 		resp, err := NewTransaction(mockWalletCallback, 100)

// 		// expectedErrorMsg := "SDK not initialized"
// 		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
// 		require.NotEmpty(t, resp)
// 		require.NoError(t, err)
// 	})
// }
// func TestSetTransactionCallback(t *testing.T) {
// 	t.Run("SetTransaction Callback Success", func(t *testing.T) {

// 		trans := &Transaction{
// 			txnOut: "test",
// 		}
// 		trans.txnStatus = StatusUnknown
// 		mockWalletCallback := TransactionCallbackImpl{}
// 		mockWalletCallback.On("OnTransactionComplete", &Transaction{}, 0).Return()
// 		err := trans.SetTransactionCallback(mockWalletCallback)
// 		// require.NoError(t, err)
// 		require.NoError(t, err)
// 	})
// 	t.Run("SetTransaction Callback transaction already exists", func(t *testing.T) {

// 		trans := &Transaction{
// 			txnOut: "test",
// 		}
// 		mockWalletCallback := TransactionCallbackImpl{}
// 		mockWalletCallback.On("OnTransactionComplete", &Transaction{}, 0).Return()
// 		err := trans.SetTransactionCallback(mockWalletCallback)
// 		// require.NoError(t, err)
// 		expectedErrorMsg := "transaction already exists. cannot set transaction hash."
// 		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
// 	})
// }
func TestSetTransactionFee(t *testing.T) {
	t.Run("SetTransaction Callback Success", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.SetTransactionFee(100)
		// require.NoError(t, err)
		require.NoError(t, err)
	})
	t.Run("SetTransaction Callback transaction already exists", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
		}

		err := trans.SetTransactionFee(100)
		// require.NoError(t, err)
		expectedErrorMsg := "transaction already exists. cannot set transaction fee."
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}

// func TestSend(t *testing.T) {
// 	t.Run("Send Success", func(t *testing.T) {
// 		_config.chain.SignatureScheme = "bls0chain"
// 		_config.chain.Miners = []string{"1", "2"}
// 		_config.wallet = zcncrypto.Wallet{
// 			Keys: []zcncrypto.KeyPair{
// 				zcncrypto.KeyPair{
// 					PublicKey:  public_key,
// 					PrivateKey: private_key,
// 				},
// 			},
// 		}

// 		transsubmit transaction erro
// 		trans.txnStatus = StatusUnknown

// 		err := trans.Send(clientID, 100, "desc")

// 		expected := &Transaction{
// 			txn: &transaction.Transaction{
// 				TransactionType: transaction.TxnTypeSend,
// 				ToClientID:      clientID,
// 				Value:           100,
// 				TransactionData: "desc",
// 			},
// 		}
// 		require.Equal(t, expected.txn.Value, trans.txn.TransactionData)
// 		require.NoError(t, err)
// 	})

// }

// func TestSendWithSignatureHash(t *testing.T) {
// 	t.Run("Complete Verify Success", func(t *testing.T) {
// 		_config.chain.Miners = []string{"1", "2"}
// 		_config.chain.MinSubmit = 9
// 		trans := &Transaction{
// 			txn: &transaction.Transaction{
// 				TransactionType: 0,
// 				ToClientID:      "",
// 				Value:           0,
// 				TransactionData: "",
// 				Signature:       "",
// 				CreationDate:    0,
// 			},
// 		}
// 		err := trans.SendWithSignatureHash(clientID, 100, "desc", "sig", 1, "hash")
// 		expected := &Transaction{
// 			verifyStatus: 1,
// 			verifyOut:    "abc",
// 			verifyError:  nil,
// 			txn: &transaction.Transaction{
// 				Value: 100,
// 			},
// 		}
// 		require.NoError(t, err)
// 		require.EqualValues(t, expected.txn.Value, trans.txn.Value)
// 	})
// }
// func TestStoreData(t *testing.T) {
// 	t.Run("New Transaction Success", func(t *testing.T) {
// 		_config.chain.SignatureScheme = "bls0chain"

// 		trans := &Transaction{
// 			txnOut: "test",
// 			txn: &transaction.Transaction{
// 				Signature: "signature",
// 			},
// 		}
// 		_config.wallet = zcncrypto.Wallet{
// 			Keys: []zcncrypto.KeyPair{
// 				zcncrypto.KeyPair{
// 					PublicKey:  public_key,
// 					PrivateKey: private_key,
// 				},
// 			},
// 		}
// 		s := &sync.WaitGroup{}
// 		s.Add(5)
// 		err := trans.StoreData("a")
// 		expected := &Transaction{
// 			txn: &transaction.Transaction{
// 				TransactionData: "a",
// 			},
// 		}
// 		expected.txn.TransactionFee = 100
// 		require.EqualValues(t, expected.txn.TransactionData, trans.txn.TransactionData)
// 		require.NoError(t, err)
// 	})
// }t.Run("New Transaction check config fails", func(t *testing.T) {

// 	_config.chain.Sharders = []string{"3", "4"}
// 	_config.isValidWallet = true
// 	_config.wallet.ClientID = "test"

// 	mockWalletCallback := TransactionCallbackImpl{}
// 	mockWalletCallback.On("OnTransactionComplete", &Transaction{}, 0).Return()

// 	resp, err := NewTransaction(mockWalletCallback, 100)

// 	expectedErrorMsg := "SDK not initialized"
// 	assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
// 	require.Nil(t, resp)
// })
// t.Run("New Transaction Success", func(t *testing.T) {
// 	_config.authUrl = "test url"
// 	_config.isConfigured = true
// 	_config.chain.Miners = []string{"1", "2"}
// 	_config.chain.Sharders = []string{"3", "4"}
// 	_config.isValidWallet = true
// 	_config.wallet.ClientID = "test"

// 	mockWalletCallback := TransactionCallbackImpl{}
// 	mockWalletCallback.On("OnTransactionComplete", &Transaction{}, 0).Return()
// 	resp, err := NewTransaction(mockWalletCallback, 100)
// 	expected := &Transaction{
// 		txn: &transaction.Transaction{},
// 	}
// 	expected.txn.TransactionFee = 100
// 	require.NotEmpty(t, resp)
// 	require.NoError(t, err)
// })

func TestCreateSmartContractTxn(t *testing.T) {
	t.Run("Create Smart Contract Txn Success", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.createSmartContractTxn("clientID", "get", "", 100)
		expected := &Transaction{
			txn: &transaction.Transaction{
				TransactionType: 1000,
				Value:           100,
				ToClientID:      "clientID",
			},
		}
		require.EqualValues(t, expected.txn.Value, trans.txn.Value)
		require.EqualValues(t, expected.txn.ToClientID, trans.txn.ToClientID)
		require.EqualValues(t, expected.txn.TransactionType, trans.txn.TransactionType)
		require.NoError(t, err)
	})
}
func TestCreateFaucetSCWallet(t *testing.T) {
	t.Run("Create Smart Contract Txn Success", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		wallet, err := trans.createFaucetSCWallet(walletString, "get", []byte("input"))

		require.NotEmpty(t, wallet)
		require.NoError(t, err)
	})
	t.Run("Create Smart Contract Txn Fails", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		wallet, err := trans.createFaucetSCWallet("walletString", "get", []byte("input"))
		expectedErrorMsg := "invalid character 'w' looking for beginning of value"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
		require.Nil(t, wallet)
	})
}
func TestSetTransactionHash(t *testing.T) {
	t.Run("Set Transaction Hash Success", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.SetTransactionHash(hash)

		require.NoError(t, err)
	})
	t.Run("Set Transaction Hash Fails", func(t *testing.T) {

		trans := &Transaction{
			txnStatus: 0,
			txnOut:    "test",
			txn:       &transaction.Transaction{},
		}

		err := trans.SetTransactionHash(hash)
		expectedErrorMsg := "transaction already exists. cannot set transaction hash."
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestGetTransactionHash(t *testing.T) {
	t.Run("Get Transaction Hash error parsing", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}

		resp := trans.GetTransactionHash()
		require.Empty(t, resp)
	})
	t.Run("Get Transaction Hash Success", func(t *testing.T) {

		trans := &Transaction{
			txnOut:    "test",
			txn:       &transaction.Transaction{},
			txnHash:   hash,
			txnStatus: 0,
		}
		trans.txnStatus = StatusUnknown

		resp := trans.GetTransactionHash()
		require.NotEmpty(t, resp)
	})
	t.Run("Get Transaction Hash empty", func(t *testing.T) {

		trans := &Transaction{
			txnOut:    "test",
			txn:       &transaction.Transaction{},
			txnStatus: 1,
		}

		resp := trans.GetTransactionHash()
		require.Empty(t, resp)
	})

}

// func TestGetLatestFinalized(t *testing.T) {
// 	t.Run("Get Transaction Hash block info not found", func(t *testing.T) {
// 		_config.chain.Sharders = []string{"1", "2", "3"}
// 		req := httptest.NewRequest(http.MethodGet, "/upper?word=abc", nil)
// 		ctx := context.Background()
// 		// var result = make(chan *util.GetResponse, 3)
// 		// for _, v := range result {

// 		// }
// 		resp, err := GetLatestFinalized(ctx, 1)
// 		require.Nil(t, resp)
// 		expectedErrorMsg := ""
// 		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
// 	})
// }
func TestGetMagicBlockByNumber(t *testing.T) {
	t.Run("Test Get Verify Output magic block info not found", func(t *testing.T) {
		ctx := context.Background()
		resp, err := GetMagicBlockByNumber(ctx, 1, 1)
		require.Empty(t, resp)
		expectedErrorMsg := "magic block info not found"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	//handle success case
}
func TestVerify(t *testing.T) {
	t.Run("Test Get Verify invalid transaction", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.verifyStatus = 1

		err := trans.Verify()

		expectedErrorMsg := "invalid transaction. cannot be verified."
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Verify invalid transaction", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
			txn: &transaction.Transaction{
				CreationDate: 1,
			},
			txnHash: hash,
		}
		trans.verifyStatus = 1

		err := trans.Verify()

		require.NoError(t, err)
	})
	// t.Run("Test Get Verify invalid transaction", func(t *testing.T) {
	// 	trans := &Transaction{
	// 		// txnOut: "test",
	// 		txn:    &transaction.Transaction{},
	// 		txnHash: hash,
	// 		txnStatus: 1,
	// 		txnOut: ` {
	// 			"ARN": "arn:aws:secretsmanager:us-east-2:xxxx:secret:team_dev-Xhzkt6",
	// 			CreatedDate: 2018-07-05 06:50:07 +0000 UTC,
	// 			Name: "team_dev",
	// 			SecretString: "{\"password\":\"test\"}",
	// 			VersionId: "6b65bfe4-7908-474b-9ae6-xxxx",
	// 			entity: "{
	// 				ARN: "arn:aws:secretsmanager:us-east-2:xxxx:secret:team_dev-Xhzkt6",
	// 				CreatedDate: 2018-07-05 06:50:07 +0000 UTC,
	// 				Name: "team_dev",
	// 				SecretString: "{\"password\":\"test\"}",
	// 				VersionId: "6b65bfe4-7908-474b-9ae6-xxxx",
	// 				VersionStages: ["AWSCURRENT"]
	// 			  }"
	// 		  }`,
	// 	}
	// 	trans.verifyStatus = 1

	// 	err := trans.Verify()

	// 	require.NoError(t,err)
	// })
	// t.Run("Test Get Verify Output Success", func(t *testing.T) {
	// 	trans := &Transaction{
	// 		txnOut: "test",
	// 		txn:    &transaction.Transaction{},
	// 		verifyOut: "test",
	// 	}
	// 	trans.verifyStatus = 0
	// 	err := trans.Verify()

	// 	require.Equal(t,"test", resp)
	// })
}
func TestGetVerifyOutput(t *testing.T) {
	t.Run("Test Get Verify Output Error", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.verifyStatus = 1

		resp := trans.GetVerifyOutput()

		require.Empty(t, resp)
	})
	t.Run("Test Get Verify Output Success", func(t *testing.T) {
		trans := &Transaction{
			txnOut:    "test",
			txn:       &transaction.Transaction{},
			verifyOut: "test",
		}
		trans.verifyStatus = 0
		resp := trans.GetVerifyOutput()

		require.Equal(t, "test", resp)
	})
}
func TestGetTransactionError(t *testing.T) {
	t.Run("Test Get Transaction Error", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = 1

		trans.txnError = errors.New("")
		resp := trans.GetTransactionError()

		require.Empty(t, resp)
	})
	t.Run("Test Get Transaction Error Success", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown
		trans.txnError = errors.New("")

		resp := trans.GetTransactionError()

		require.Empty(t, resp)
	})

}
func TestGetVerifyError(t *testing.T) {
	t.Run("Test Get Verify Error Success", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown
		resp := trans.GetVerifyError()

		require.Empty(t, resp)
	})
	t.Run("Test Get Verify Error", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown
		trans.verifyStatus = 1
		trans.verifyError = errors.New("")
		resp := trans.GetVerifyError()

		require.Empty(t, resp)
	})
}

func TestVestingStop(t *testing.T) {
	t.Run("Test Vesting Stop Success", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown
		err := trans.VestingStop(&VestingStopRequest{})

		require.NoError(t, err)
	})
}
func TestVestingUnlock(t *testing.T) {
	t.Run("Test Vesting Unlock Success", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown
		err := trans.VestingUnlock("poolID")

		require.NoError(t, err)
	})
}
func TestVestingAdd(t *testing.T) {
	t.Run("Test Vesting Add Success", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown
		err := trans.VestingAdd(&VestingAddRequest{}, 1)

		require.NoError(t, err)
	})
}
func TestVestingDelete(t *testing.T) {
	t.Run("Test Vesting Update Config Success", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown
		err := trans.VestingDelete("poolID")

		require.NoError(t, err)
	})
}
func TestVestingUpdateConfig(t *testing.T) {
	t.Run("Test Vesting Update Config Success", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown
		err := trans.VestingUpdateConfig(&VestingSCConfig{})

		require.NoError(t, err)
	})
}
func TestMinerSCSettings(t *testing.T) {
	t.Run("Test Miner SCSettings Success", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown
		err := trans.MinerSCSettings(&MinerSCMinerInfo{})

		require.NoError(t, err)
	})
}
func TestMinerSCLock(t *testing.T) {
	t.Run("Test Miner SCLock Success", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown
		err := trans.MinerSCLock("nodeID", 1)

		require.NoError(t, err)
	})
}
func TestMienrSCUnlock(t *testing.T) {
	t.Run("Test Mienr SCUnlock Success", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown
		err := trans.MienrSCUnlock("nodeID", "poolID")

		require.NoError(t, err)
	})
}
func TestLockTokens(t *testing.T) {
	t.Run("Test Unlock Tokens Success", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown
		err := trans.UnlockTokens("poolID")

		require.NoError(t, err)
	})
}
func TestUnlockTokens(t *testing.T) {
	t.Run("Test Unlock Tokens Success", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown
		err := trans.UnlockTokens("poolID")

		require.NoError(t, err)
	})
}
func TestRegisterMultiSig(t *testing.T) {
	t.Run("Test Register MultiSig Success", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown
		err := trans.RegisterMultiSig(walletString, msw)

		require.NoError(t, err)
	})

}

// func TestNewMSTransaction(t *testing.T) {
// 	t.Run("Test New MSTransaction Success", func(t *testing.T) {
// 		mockWalletCallback := mocks.TransactionCallbackImpl{}
// 		mockWalletCallback.On("OnTransactionComplete", &Transaction{}, 0).Return()
// 		resp, err := NewMSTransaction(walletString, mockWalletCallback)

// 		require.NotNil(t, resp)
// 		require.NoError(t, err)
// 	})
// 	t.Run("Test New MSTransaction Fails", func(t *testing.T) {
// 		mockWalletCallback := mocks.TransactionCallbackImpl{}
// 		mockWalletCallback.On("OnTransactionComplete", &Transaction{}, 0).Return()
// 		resp, err := NewMSTransaction("walletString", mockWalletCallback)

// 		require.Nil(t, resp)
// 		expectedErrorMsg := "invalid character 'w' looking for beginning of value"
// 		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
// 	})
// }
func TestRegisterVote(t *testing.T) {
	t.Run("Test Register Vote Success", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.RegisterVote(walletString, msv)

		require.NoError(t, err)
	})
	t.Run("Test Register Vote Fails", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.RegisterVote("walletString", msv)
		expectedErrorMsg := "invalid character 'w' looking for beginning of value"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestVerifyContentHash(t *testing.T) {
	t.Run("Test Verify Content Hash Fails", func(t *testing.T) {
		resp, err := VerifyContentHash(`{"txn_id": "test"}`)

		require.False(t, resp)
		expectedErrorMsg := "fetch_txm_details: Unable to fetch txn details"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	// handle case success
}
func TestFinalizeAllocation(t *testing.T) {
	t.Run("Test Finalize Allocation Success", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.FinalizeAllocation("allocID", 1)

		require.NoError(t, err)
	})
}
func TestCancelAllocation(t *testing.T) {
	t.Run("Test Cancel Allocation Success", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.CancelAllocation("allocID", 1)

		require.NoError(t, err)
	})
}
func TestCreateAllocation(t *testing.T) {
	t.Run("Test Create Allocation Success", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.CreateAllocation(&CreateAllocationRequest{}, 1, 1)

		require.NoError(t, err)
	})
}
func TestCreateReadPool(t *testing.T) {
	t.Run("Test Create Read Pool Success", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.CreateReadPool(1)

		require.NoError(t, err)
	})
}
func TestReadPoolLock(t *testing.T) {
	t.Run("Test Read Pool Lock Success", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.ReadPoolLock("allocID", "blobberID", 1, 1, 1)

		require.NoError(t, err)
	})
}
func TestReadPoolUnlock(t *testing.T) {
	t.Run("Test Read Pool Unlock Success", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.ReadPoolUnlock("poolID", 1)

		require.NoError(t, err)
	})
}
func TestStakePoolLock(t *testing.T) {
	t.Run("Test Stake Pool Lock Success", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.StakePoolLock("blobberID", 1, 1)

		require.NoError(t, err)
	})
}
func TestStakePoolUnlock(t *testing.T) {
	t.Run("Test Stake Pool Unlock Success", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.StakePoolUnlock("blobberID", "poolID", 1)

		require.NoError(t, err)
	})

}
func TestStakePoolPayInterests(t *testing.T) {
	t.Run("Stake Pool Pay Interests Success", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.StakePoolPayInterests("blobberID", 1)

		require.NoError(t, err)
	})

}
func TestUpdateBlobberSettings(t *testing.T) {
	t.Run("Update Blobber Settings Success", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.UpdateBlobberSettings(&Blobber{}, 1)

		require.NoError(t, err)
	})

}
func TestUpdateAllocation(t *testing.T) {
	t.Run("Update Allocation Success", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.UpdateAllocation("allocID", 1, 1, 1, 1)

		require.NoError(t, err)
	})

}
func TestWritePoolLock(t *testing.T) {
	t.Run("Write Pool Lock Success", func(t *testing.T) {

		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.WritePoolLock("allocID", "blobberID", 1, 1, 1)

		require.NoError(t, err)
	})

}
func TestWritePoolUnlock(t *testing.T) {
	t.Run("Write Pool Unlock Success", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.WritePoolUnlock("poolID", 100)

		require.NoError(t, err)
	})

}

// func TestTransaction_VestingTrigger(t *testing.T) {
// 	type fields struct {
// 		txn          *transaction.Transaction
// 		txnOut       string
// 		txnHash      string
// 		txnStatus    int
// 		txnError     error
// 		txnCb        TransactionCallback
// 		verifyStatus int
// 		verifyOut    string
// 		verifyError  error
// 	}
// 	type args struct {
// 		poolID string
// 	}
// 	tests := []struct {
// 		name       string

// 		setup      func(*testing.T, string, string)
// 		wantErr    bool
// 		errMsg     string
// 	}{
// 		name: "Vesting_Trigger",

// 		setup: func(t *testing.T, name string, p parameters, errMsg string) {
// 			mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
// 				return strings.HasPrefix(req.URL.Path, "Test_Http_Error")
// 			})).Return(&http.Response{
// 				Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
// 				StatusCode: p.respStatusCode,
// 			}, fmt.Errorf(mockErrorMessage))
// 		},
// 		wantErr: true,
// 		errMsg:  mockErrorMessage,
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tr := &Transaction{
// 				txn:          tt.fields.txn,
// 				txnOut:       tt.fields.txnOut,
// 				txnHash:      tt.fields.txnHash,
// 				txnStatus:    tt.fields.txnStatus,
// 				txnError:     tt.fields.txnError,
// 				txnCb:        tt.fields.txnCb,
// 				verifyStatus: tt.fields.verifyStatus,
// 				verifyOut:    tt.fields.verifyOut,
// 				verifyError:  tt.fields.verifyError,
// 			}
// 			if err := tr.VestingTrigger(tt.args.poolID); (err != nil) != tt.wantErr {
// 				t.Errorf("Transaction.VestingTrigger() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }
