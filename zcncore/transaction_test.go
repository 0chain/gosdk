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
	"time"

	"github.com/0chain/gosdk/core/block"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	mockHash = "15743192c9f78cf56824f83d92d54e48f50ca53c305a316ad7070b9ba4fac486"
)

func setupMockHttpResponse(body []byte) {
	var mockClient = mocks.HttpClient{}
	util.Client = &mockClient
	mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		fmt.Println("**********", req.URL.Path)
		return strings.HasPrefix(req.URL.Path, "TestTransaction")
	})).Return(&http.Response{
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(body))),
		StatusCode: http.StatusOK,
	}, nil)
}

func setupMockSubmitTxn() {
	_config.chain.SignatureScheme = "bls0chain"
	_config.wallet = zcncrypto.Wallet{
		Keys: []zcncrypto.KeyPair{
			{
				PublicKey:  mockPublicKey,
				PrivateKey: mockPrivateKey,
			},
		},
	}
}

func TestSignFn(t *testing.T) {
	t.Run("Test_Success", func(t *testing.T) {
		_config.chain.SignatureScheme = "bls0chain"
		_config.wallet = zcncrypto.Wallet{
			Keys: []zcncrypto.KeyPair{
				{
					PublicKey:  mockPublicKey,
					PrivateKey: mockPrivateKey,
				},
			},
		}
		resp, err := signFn("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
		require.NoError(t, err)
		require.NotNil(t, resp)
	})
}

func TestSignWithWallet(t *testing.T) {
	t.Run("Sign With Wallet Success", func(t *testing.T) {
		_config.chain.SignatureScheme = "bls0chain"
		resp, err := signWithWallet("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", &zcncrypto.Wallet{
			Keys: []zcncrypto.KeyPair{
				{
					PublicKey:  mockPublicKey,
					PrivateKey: mockPrivateKey,
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
	})
}

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

func TestTransaction_submitTxn(t *testing.T) {
	_config.chain.SignatureScheme = "bls0chain"
	_config.wallet = zcncrypto.Wallet{
		Keys: []zcncrypto.KeyPair{
			{
				PublicKey:  mockPublicKey,
				PrivateKey: mockPrivateKey,
			},
		},
	}
	var mockClient = mocks.HttpClient{}
	util.Client = &mockClient
	tests := []struct {
		name     string
		setup    func(*testing.T)
		wantFunc func(*Transaction)
	}{
		{
			name: "Test_Error_Submit_Transaction_Failed",
			setup: func(t *testing.T) {
				_config.chain.Miners = []string{"TestTransaction_submitTxnTest_Error_Submit_Transaction_Failed"}
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, "TestTransaction_submitTxnTest_Error_Submit_Transaction_Failed")
				})).Return(&http.Response{
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
					StatusCode: http.StatusBadRequest,
				}, nil)
			},
			wantFunc: func(trans *Transaction) {
				require.EqualValues(t, trans.txnStatus, StatusError)
				require.Contains(t, trans.txnError.Error(), "submit transaction failed. ")
			},
		},
		{
			name: "Test_Success",
			setup: func(t *testing.T) {
				_config.chain.Miners = []string{"TestTransaction_submitTxnTest_Success"}
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, "TestTransaction_submitTxnTest_Success")
				})).Return(&http.Response{
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
					StatusCode: http.StatusOK,
				}, nil)
			},
			wantFunc: func(trans *Transaction) {
				require.EqualValues(t, trans.txnStatus, StatusSuccess)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}
			trans := &Transaction{
				txnOut: "test",
				txn:    &transaction.Transaction{},
			}
			trans.submitTxn()
			tt.wantFunc(trans)
		})
	}
}

func TestNewTransaction(t *testing.T) {
	t.Run("Test_Success", func(t *testing.T) {
		_config.chain.ChainID = "mock chain id"
		_config.wallet = zcncrypto.Wallet{
			ClientID:  "mock client id",
			ClientKey: "mock client key",
			Keys: []zcncrypto.KeyPair{
				{
					PublicKey:  mockPublicKey,
					PrivateKey: mockPrivateKey,
				},
			},
		}
		mockWalletCallback := MockTransactionCallback{}
		mockWalletCallback.On("OnTransactionComplete", &Transaction{}, 0).Return()
		trans, err := newTransaction(mockWalletCallback, 100)
		expected := &Transaction{
			txn: &transaction.Transaction{},
		}
		expected.txn.TransactionFee = 100
		require.EqualValues(t, trans.txn.TransactionFee, expected.txn.TransactionFee)
		require.NoError(t, err)
	})
}

func TestNewTransactionFunction(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*testing.T) (teardown func(*testing.T))
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Test_Error_Config",
			wantErr: true,
			errMsg:  "SDK not initialized",
		},
		{
			name: "Test_Error_Auth_Url_Not_Set",
			setup: func(t *testing.T) (teardown func(t *testing.T)) {
				_config.chain.Sharders = []string{"TestNewTransactionFunction"}
				_config.chain.Miners = []string{"0", "1"}
				_config.isValidWallet = true
				_config.isSplitWallet = true
				_config.wallet.ClientID = "test"
				_config.isConfigured = true
				return func(t *testing.T) {
					_config.chain.Sharders = nil
				}
			},
			wantErr: true,
			errMsg:  "auth url not set",
		},
		{
			name: "Test_Success_With_Auth",
			setup: func(t *testing.T) (teardown func(t *testing.T)) {
				_config.chain.Sharders = []string{"TestNewTransactionFunction"}
				_config.chain.Miners = []string{"0", "1"}
				_config.isValidWallet = true
				_config.isSplitWallet = true
				_config.wallet.ClientID = "test"
				_config.isConfigured = true
				_config.authUrl = "mockauthurl"
				return func(t *testing.T) {
					_config.chain.Sharders = nil
				}
			},
		},
		{
			name: "Test_Success",
			setup: func(t *testing.T) (teardown func(t *testing.T)) {
				_config.chain.Sharders = []string{"TestNewTransactionFunction"}
				_config.chain.Miners = []string{"0", "1"}
				_config.isValidWallet = true
				_config.wallet.ClientID = "test"
				_config.isConfigured = true
				return func(t *testing.T) {
					_config.chain.Sharders = nil
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if tt.setup != nil {
				if teardown := tt.setup(t); teardown != nil {
					defer teardown(t)
				}
			}
			mockWalletCallback := MockTransactionCallback{}
			mockWalletCallback.On("OnTransactionComplete", &Transaction{}, 0).Return()
			_, err := NewTransaction(mockWalletCallback, 100)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.Contains(err.Error(), tt.errMsg)
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestSetTransactionCallback(t *testing.T) {
	type parameters struct {
		txnStatus int
	}
	tests := []struct {
		name       string
		parameters parameters
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "Test_Error_Transaction_Already_Exists",
			parameters: parameters{},
			wantErr:    true,
			errMsg:     "transaction already exists. cannot set transaction hash.",
		},
		{
			name: "Test_Success",
			parameters: parameters{
				txnStatus: StatusUnknown,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			trans := &Transaction{
				txnOut:    "test",
				txnStatus: tt.parameters.txnStatus,
			}
			mockWalletCallback := MockTransactionCallback{}
			mockWalletCallback.On("OnTransactionComplete", &Transaction{}, 0).Return()
			err := trans.SetTransactionCallback(mockWalletCallback)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.Contains(err.Error(), tt.errMsg)
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
		})
	}
}

func TestSetTransactionFee(t *testing.T) {
	t.Run("Test_Success", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.SetTransactionFee(100)
		require.NoError(t, err)
	})
	t.Run("Test_Error_Transaction_Already_Exists", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
		}

		err := trans.SetTransactionFee(100)
		expectedErrorMsg := "transaction already exists. cannot set transaction fee."
		require.Contains(t, err.Error(), expectedErrorMsg)
	})
}

func TestTransaction_Send(t *testing.T) {
	t.Run("Test_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_Send"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.Send(clientID, 100, "desc")
		require.NoError(t, err)
	})
}

func TestSendWithSignatureHash(t *testing.T) {
	t.Run("Test_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestSendWithSignatureHash"}
		_config.chain.MinSubmit = 9
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
		trans := &Transaction{
			txn: &transaction.Transaction{
				TransactionType: 0,
				Value:           0,
				CreationDate:    0,
			},
		}
		err := trans.SendWithSignatureHash(clientID, 100, "desc", "sig", 1, "hash")
		require.NoError(t, err)
	})
}

func TestTransaction_StoreData(t *testing.T) {
	t.Run("New Transaction Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_StoreData"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
		trans := &Transaction{
			txnOut: "test",
			txn: &transaction.Transaction{
				Signature: "signature",
			},
		}
		err := trans.StoreData("a")
		require.NoError(t, err)
	})
}

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
		require.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
		require.Nil(t, wallet)
	})
}

func TestExecuteFaucetSCWallet(t *testing.T) {
	_config.chain.Miners = []string{"TestTransaction_TestExecuteFaucetSCWallet"}
	setupMockSubmitTxn()
	setupMockHttpResponse([]byte(""))
	trans := &Transaction{
		txnOut: "test",
		txn:    &transaction.Transaction{},
	}
	trans.txnStatus = StatusUnknown

	err := trans.ExecuteFaucetSCWallet(walletString, "get", []byte("input"))
	require.NoError(t, err)
}

func TestTransaction_ExecuteSmartContract(t *testing.T) {
	_config.chain.Miners = []string{"TestTransaction_ExecuteSmartContract"}
	setupMockSubmitTxn()
	setupMockHttpResponse([]byte(""))
	trans := &Transaction{
		txnOut: "test",
		txn:    &transaction.Transaction{},
	}
	trans.txnStatus = StatusUnknown

	err := trans.ExecuteSmartContract("mockaddress", "get", `{"input": "mockInput"}`, 0)
	time.Sleep(5 * time.Second)
	require.NoError(t, err)
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
		require.Contains(t, err.Error(), expectedErrorMsg)
	})
}

func TestGetTransactionHash(t *testing.T) {
	t.Run("Get_Transaction_Hash_Error_Parsing", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}

		resp := trans.GetTransactionHash()
		require.Empty(t, resp)
	})
	t.Run("Get_Transaction_Hash_Success", func(t *testing.T) {
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
	t.Run("Get_Transaction_Hash_Empty", func(t *testing.T) {
		trans := &Transaction{
			txnOut:    "test",
			txn:       &transaction.Transaction{},
			txnStatus: 1,
		}

		resp := trans.GetTransactionHash()
		require.Empty(t, resp)
	})
}

func Test_getBlockHeaderFromTransactionConfirmation(t *testing.T) {
	type parameters struct {
		txnHash  string
		cfmBlock map[string]json.RawMessage
	}
	tests := []struct {
		name       string
		parameters parameters
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "Test_Error_Txn_Confirmation_Not_Found",
			parameters: parameters{},
			wantErr:    true,
			errMsg:     "txn confirmation not found.",
		},
		{
			name: "Test_Error_Txn_Confirmation_Parse",
			parameters: parameters{
				cfmBlock: map[string]json.RawMessage{
					"confirmation": []byte(`{error}`),
				},
			},
			wantErr: true,
			errMsg:  "txn confirmation parse error. invalid character 'e' looking for beginning of object key string",
		},
		{
			name: "Test_Error_Missing_Transaction",
			parameters: parameters{
				cfmBlock: map[string]json.RawMessage{
					"confirmation": []byte(`{}`),
				},
			},
			wantErr: true,
			errMsg:  "missing transaction  in block confirmation",
		},
		{
			name: "Test_Error_Invalid_Transaction_Hash",
			parameters: parameters{
				cfmBlock: map[string]json.RawMessage{
					"confirmation": []byte(`{"txn":{"hash":"mockhash"}}`),
				},
			},
			wantErr: true,
			errMsg:  "invalid transaction hash. Expected: . Received: mockhash",
		},
		{
			name: "Test_Error_Txn_Merkle_Validation_Failed",
			parameters: parameters{
				cfmBlock: map[string]json.RawMessage{
					"confirmation": []byte(`{"txn":{"hash":"mockhash"},"merkle_tree_path":{"nodes":["mockmerkletreepath"]},"merkle_tree_root":"mockmerkletreepath"}`),
				},
				txnHash: "mockhash",
			},
			wantErr: true,
			errMsg:  "txn merkle validation failed.",
		},
		{
			name: "Test_Error_Txn_Receipt_Merkle_Validation_Failed",
			parameters: parameters{
				cfmBlock: map[string]json.RawMessage{
					"confirmation": []byte(`{"txn":{"hash":"mockhash"},"merkle_tree_path":{"nodes":["mockmerkletreepath"]},"merkle_tree_root":"29e7836bba52aa41c929e5018cf5eaee051ab8f21ec471fa09d7a240e469cf7b","receipt_merkle_tree_path":{"nodes":["mockreceiptmerkletreepath"]},"receipt_merkle_tree_root":"mockreceiptmerkletreepath"}`),
				},
				txnHash: "mockhash",
			},
			wantErr: true,
			errMsg:  "txn receipt cmerkle validation failed.",
		},
		{
			name: "Test_Error_Block_Hash_Verification_Failed",
			parameters: parameters{
				cfmBlock: map[string]json.RawMessage{
					"confirmation": []byte(`{"txn":{"hash":"mockhash"},"merkle_tree_path":{"nodes":["mockmerkletreepath"]},"merkle_tree_root":"29e7836bba52aa41c929e5018cf5eaee051ab8f21ec471fa09d7a240e469cf7b","receipt_merkle_tree_path":{"nodes":["mockreceiptmerkletreepath"]},"receipt_merkle_tree_root":"1af88a7aadb2543f58f0ed07f56c5cf1bb52c9274aa4231a7900ccf1da3dc7c5"}`),
				},
				txnHash: "mockhash",
			},
			wantErr: true,
			errMsg:  "block hash verification failed in confirmation",
		},
		{
			name: "Test_Success",
			parameters: parameters{
				cfmBlock: map[string]json.RawMessage{
					"confirmation": []byte(`{"block_hash":"0a80e736fb4dd8b9b976695a67d8baf0b6184f3ba7eb51d2e99caf320cb0e2af","txn":{"hash":"mockhash"},"merkle_tree_path":{"nodes":["mockmerkletreepath"]},"merkle_tree_root":"29e7836bba52aa41c929e5018cf5eaee051ab8f21ec471fa09d7a240e469cf7b","receipt_merkle_tree_path":{"nodes":["mockreceiptmerkletreepath"]},"receipt_merkle_tree_root":"1af88a7aadb2543f58f0ed07f56c5cf1bb52c9274aa4231a7900ccf1da3dc7c5"}`),
				},
				txnHash: "mockhash",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			got, err := getBlockHeaderFromTransactionConfirmation(tt.parameters.txnHash, tt.parameters.cfmBlock)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.Contains(err.Error(), tt.errMsg)
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
			expectedValue := &blockHeader{
				Hash:                  "0a80e736fb4dd8b9b976695a67d8baf0b6184f3ba7eb51d2e99caf320cb0e2af",
				MerkleTreeRoot:        "29e7836bba52aa41c929e5018cf5eaee051ab8f21ec471fa09d7a240e469cf7b",
				ReceiptMerkleTreeRoot: "1af88a7aadb2543f58f0ed07f56c5cf1bb52c9274aa4231a7900ccf1da3dc7c5",
			}
			require.EqualValues(expectedValue, got)
		})
	}
}

func TestGetTransactionConfirmation(t *testing.T) {
	var mockClient = mocks.HttpClient{}
	util.Client = &mockClient
	tests := []struct {
		name    string
		setup   func(*testing.T) (teardown func(*testing.T))
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Test_Failed",
			wantErr: true,
			errMsg:  "transaction not found",
		},
		{
			name: "Test_Success",
			setup: func(t *testing.T) (teardown func(t *testing.T)) {
				_config.chain.Sharders = []string{"TestGetTransactionConfirmation"}
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, "TestGetTransactionConfirmation")
				})).Return(&http.Response{
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"confirmation":{"block_hash":"0a80e736fb4dd8b9b976695a67d8baf0b6184f3ba7eb51d2e99caf320cb0e2af","txn":{"hash":"mockhash"},"merkle_tree_path":{"nodes":["mockmerkletreepath"]},"merkle_tree_root":"29e7836bba52aa41c929e5018cf5eaee051ab8f21ec471fa09d7a240e469cf7b","receipt_merkle_tree_path":{"nodes":["mockreceiptmerkletreepath"]},"receipt_merkle_tree_root":"1af88a7aadb2543f58f0ed07f56c5cf1bb52c9274aa4231a7900ccf1da3dc7c5"}}`))),
					StatusCode: http.StatusOK,
				}, nil).Once()
				return func(t *testing.T) {
					_config.chain.Sharders = nil
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if tt.setup != nil {
				if teardown := tt.setup(t); teardown != nil {
					defer teardown(t)
				}
			}
			got, _, _, err := getTransactionConfirmation(1, "mockhash")
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.Contains(err.Error(), tt.errMsg)
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
			expectedValue := &blockHeader{
				Hash:                  "0a80e736fb4dd8b9b976695a67d8baf0b6184f3ba7eb51d2e99caf320cb0e2af",
				MerkleTreeRoot:        "29e7836bba52aa41c929e5018cf5eaee051ab8f21ec471fa09d7a240e469cf7b",
				ReceiptMerkleTreeRoot: "1af88a7aadb2543f58f0ed07f56c5cf1bb52c9274aa4231a7900ccf1da3dc7c5",
			}
			require.EqualValues(expectedValue, got)
		})
	}
}

func TestGetLatestFinalized(t *testing.T) {
	var mockClient = mocks.HttpClient{}
	util.Client = &mockClient
	tests := []struct {
		name    string
		setup   func(*testing.T) (teardown func(*testing.T))
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Test_Failed",
			wantErr: true,
			errMsg:  "block info not found",
		},
		{
			name: "Test_Success",
			setup: func(t *testing.T) (teardown func(t *testing.T)) {
				_config.chain.Sharders = []string{"TestGetLatestFinalized"}
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, "TestGetLatestFinalized")
				})).Return(&http.Response{
					Body: func() io.ReadCloser {
						jsonFR, err := json.Marshal(&block.Header{
							Hash: "mockhash",
						})
						require.NoError(t, err)
						return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
					}(),
					StatusCode: http.StatusOK,
				}, nil).Once()
				return func(t *testing.T) {
					_config.chain.Sharders = nil
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if tt.setup != nil {
				if teardown := tt.setup(t); teardown != nil {
					defer teardown(t)
				}
			}
			ctx := context.Background()
			got, err := GetLatestFinalized(ctx, 1)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.Contains(err.Error(), tt.errMsg)
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
			expectedValue := &block.Header{
				Hash: "mockhash",
			}
			require.EqualValues(expectedValue, got)
		})
	}
}

func TestGetLatestFinalizedMagicBlock(t *testing.T) {
	var mockClient = mocks.HttpClient{}
	util.Client = &mockClient
	tests := []struct {
		name    string
		setup   func(*testing.T) (teardown func(*testing.T))
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Test_Failed",
			wantErr: true,
			errMsg:  "magic block info not found",
		},
		{
			name: "Test_Success",
			setup: func(t *testing.T) (teardown func(t *testing.T)) {
				_config.chain.Sharders = []string{"TestGetLatestFinalizedMagicBlock"}
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, "TestGetLatestFinalizedMagicBlock")
				})).Return(&http.Response{
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"magic_block":{"hash":"mockhash"}}`))),
					StatusCode: http.StatusOK,
				}, nil).Once()
				return func(t *testing.T) {
					_config.chain.Sharders = nil
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if tt.setup != nil {
				if teardown := tt.setup(t); teardown != nil {
					defer teardown(t)
				}
			}
			ctx := context.Background()
			got, err := GetLatestFinalizedMagicBlock(ctx, 1)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.Contains(err.Error(), tt.errMsg)
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
			expectedValue := &block.MagicBlock{
				Hash: "mockhash",
			}
			require.EqualValues(expectedValue, got)
		})
	}
}

func TestGetChainStats(t *testing.T) {
	var mockClient = mocks.HttpClient{}
	util.Client = &mockClient
	tests := []struct {
		name    string
		setup   func(*testing.T) (teardown func(*testing.T))
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Test_Failed",
			wantErr: true,
			errMsg:  "http_request_failed: Request failed with status not 200",
		},
		{
			name: "Test_Success",
			setup: func(t *testing.T) (teardown func(t *testing.T)) {
				_config.chain.Sharders = []string{"TestGetChainStats"}
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, "TestGetChainStats")
				})).Return(&http.Response{
					Body: func() io.ReadCloser {
						jsonFR, err := json.Marshal(&block.ChainStats{
							Count: 10,
						})
						require.NoError(t, err)
						return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
					}(),
					StatusCode: http.StatusOK,
				}, nil).Once()
				return func(t *testing.T) {
					_config.chain.Sharders = nil
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if tt.setup != nil {
				if teardown := tt.setup(t); teardown != nil {
					defer teardown(t)
				}
			}
			ctx := context.Background()
			got, err := GetChainStats(ctx)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.Contains(err.Error(), tt.errMsg)
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
			expectedValue := &block.ChainStats{
				Count: 10,
			}
			require.EqualValues(expectedValue, got)
		})
	}
}

func TestGetBlockByRound(t *testing.T) {
	var mockClient = mocks.HttpClient{}
	util.Client = &mockClient
	tests := []struct {
		name    string
		setup   func(*testing.T) (teardown func(*testing.T))
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Test_Failed",
			wantErr: true,
			errMsg:  "round info not found",
		},
		{
			name: "Test_Success",
			setup: func(t *testing.T) (teardown func(t *testing.T)) {
				_config.chain.Sharders = []string{"TestGetBlockByRound"}
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, "TestGetBlockByRound")
				})).Return(&http.Response{
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"block":{"hash":"mockhash"},"header":{"hash":"mockhash"}}`))),
					StatusCode: http.StatusOK,
				}, nil).Once()
				return func(t *testing.T) {
					_config.chain.Sharders = nil
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if tt.setup != nil {
				if teardown := tt.setup(t); teardown != nil {
					defer teardown(t)
				}
			}
			ctx := context.Background()
			got, err := GetBlockByRound(ctx, 1, 1)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.Contains(err.Error(), tt.errMsg)
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
			expectedValue := &block.Block{
				Header: &block.Header{
					Hash: "mockhash",
				},
				Hash: "mockhash",
			}
			require.EqualValues(expectedValue, got)
		})
	}
}

func TestGetMagicBlockByNumber(t *testing.T) {
	var mockClient = mocks.HttpClient{}
	util.Client = &mockClient
	tests := []struct {
		name    string
		setup   func(*testing.T) (teardown func(*testing.T))
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Test_Failed",
			wantErr: true,
			errMsg:  "magic block info not found",
		},
		{
			name: "Test_Success",
			setup: func(t *testing.T) (teardown func(t *testing.T)) {
				_config.chain.Sharders = []string{"TestGetMagicBlockByNumber"}
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, "TestGetMagicBlockByNumber")
				})).Return(&http.Response{
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"magic_block":{"hash":"mockhash"}}`))),
					StatusCode: http.StatusOK,
				}, nil).Once()
				return func(t *testing.T) {
					_config.chain.Sharders = nil
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if tt.setup != nil {
				if teardown := tt.setup(t); teardown != nil {
					defer teardown(t)
				}
			}
			ctx := context.Background()
			got, err := GetMagicBlockByNumber(ctx, 1, 1)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.Contains(err.Error(), tt.errMsg)
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
			expectedValue := &block.MagicBlock{
				Hash: "mockhash",
			}
			require.EqualValues(expectedValue, got)
		})
	}
}

func Test_getBlockInfoByRound(t *testing.T) {
	var mockClient = mocks.HttpClient{}
	util.Client = &mockClient
	tests := []struct {
		name    string
		setup   func(*testing.T) (teardown func(*testing.T))
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Test_Failed",
			wantErr: true,
			errMsg:  "round info not found.",
		},
		{
			name: "Test_Success",
			setup: func(t *testing.T) (teardown func(t *testing.T)) {
				_config.chain.Sharders = []string{"TestGetBlockInfoByRound"}
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, "TestGetBlockInfoByRound")
				})).Return(&http.Response{
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"header":{"hash":"mockhash"}}`))),
					StatusCode: http.StatusOK,
				}, nil).Once()
				return func(t *testing.T) {
					_config.chain.Sharders = nil
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if tt.setup != nil {
				if teardown := tt.setup(t); teardown != nil {
					defer teardown(t)
				}
			}
			got, err := getBlockInfoByRound(1, 1, "test")
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.Contains(err.Error(), tt.errMsg)
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
			expectedValue := &blockHeader{
				Hash: "mockhash",
			}
			require.EqualValues(expectedValue, got)
		})
	}
}

func Test_isBlockExtends(t *testing.T) {
	type parameters struct {
		block *blockHeader
	}
	tests := []struct {
		name       string
		parameters parameters
		want       bool
	}{
		{
			name: "Test_False",
			parameters: parameters{
				block: &blockHeader{
					MerkleTreeRoot:        "mockmerkletreeroot",
					ReceiptMerkleTreeRoot: "mockreceiptmerkletreeroot",
				},
			},
			want: false,
		},
		{
			name: "Test_False",
			parameters: parameters{
				block: &blockHeader{
					MerkleTreeRoot:        "mockmerkletreeroot",
					ReceiptMerkleTreeRoot: "mockreceiptmerkletreeroot",
					Hash:                  "0ea7929fb3185ec429db035e0d7b3337612d24f0fba64e3b253612be2f3017d7",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isBlockExtends("", tt.parameters.block); got != tt.want {
				t.Errorf("isBlockExtends() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateChain(t *testing.T) {
	t.Run("Test_Validate_Chain_Success", func(t *testing.T) {
		_config.chain.Sharders = []string{"TestValidateChain"}
		var mockClient = mocks.HttpClient{}
		util.Client = &mockClient
		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return strings.HasPrefix(req.URL.Path, "TestValidateChain")
		})).Return(&http.Response{
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"header":{"hash":"mockhash"}}`))),
			StatusCode: http.StatusOK,
		}, nil)
		got := validateChain(&blockHeader{
			CreationDate:    1,
			Round:           1,
			RoundRandomSeed: 1,
			NumTxns:         1,
		})
		require.True(t, got)
	})
}

func TestIsTransactionExpired(t *testing.T) {
	t.Run("Test_Is_Transaction_Expired_False", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown
		_config.wallet = zcncrypto.Wallet{
			Keys: []zcncrypto.KeyPair{
				{
					PublicKey:  public_key,
					PrivateKey: private_key,
				},
			},
		}
		resp := trans.isTransactionExpired(1, 1)

		require.False(t, resp)
	})
	t.Run("Test_Is_Transaction_Expired_False", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown
		_config.wallet = zcncrypto.Wallet{
			Keys: []zcncrypto.KeyPair{
				{
					PublicKey:  public_key,
					PrivateKey: private_key,
				},
			},
		}
		resp := trans.isTransactionExpired(0, 1)

		require.False(t, resp)
	})
	t.Run("Test_Is_Transaction_Expired_True", func(t *testing.T) {
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown
		_config.wallet = zcncrypto.Wallet{
			Keys: []zcncrypto.KeyPair{
				{
					PublicKey:  public_key,
					PrivateKey: private_key,
				},
			},
		}
		resp := trans.isTransactionExpired(100, 100)

		require.True(t, resp)
	})
}

func TestVerify(t *testing.T) {
	var mockClient = mocks.HttpClient{}
	util.Client = &mockClient
	type parameters struct {
		txn     *transaction.Transaction
		txnHash string
	}
	tests := []struct {
		name       string
		parameters parameters
		wantErr    bool
		setup      func(*testing.T) (teardown func(*testing.T))
		errMsg     string
	}{
		{
			name: "Test_Invalid_Transaction",
			parameters: parameters{
				txn: &transaction.Transaction{},
			},
			wantErr: true,
			errMsg:  "invalid transaction. cannot be verified.",
		},
		{
			name: "Test_Success",
			parameters: parameters{
				txn: &transaction.Transaction{
					CreationDate: 1,
				},
				txnHash: "mockhash",
			},
			setup: func(t *testing.T) (teardown func(t *testing.T)) {
				_config.chain.Sharders = []string{"TestVerify"}

				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, "TestVerify")
				})).Return(&http.Response{
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"confirmation":{"block_hash":"0a80e736fb4dd8b9b976695a67d8baf0b6184f3ba7eb51d2e99caf320cb0e2af","txn":{"hash":"mockhash"},"merkle_tree_path":{"nodes":["mockmerkletreepath"]},"merkle_tree_root":"29e7836bba52aa41c929e5018cf5eaee051ab8f21ec471fa09d7a240e469cf7b","receipt_merkle_tree_path":{"nodes":["mockreceiptmerkletreepath"]},"receipt_merkle_tree_root":"1af88a7aadb2543f58f0ed07f56c5cf1bb52c9274aa4231a7900ccf1da3dc7c5"}}`))),
					StatusCode: http.StatusOK,
				}, nil).Once()

				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, "TestVerify")
				})).Return(&http.Response{
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"header":{"hash":"mockhash"}}`))),
					StatusCode: http.StatusOK,
				}, nil).Once()

				return func(t *testing.T) {
					_config.chain.Sharders = nil
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if tt.setup != nil {
				if teardown := tt.setup(t); teardown != nil {
					defer teardown(t)
				}
			}
			tr := &Transaction{
				txnOut:  "test",
				txn:     tt.parameters.txn,
				txnHash: tt.parameters.txnHash,
			}
			err := tr.Verify()
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.Contains(err.Error(), tt.errMsg)
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
		})
	}
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

func TestNewMSTransaction(t *testing.T) {
	type parameters struct {
		walletString string
	}
	tests := []struct {
		name       string
		parameters parameters
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Test_Error_Parse",
			parameters: parameters{
				walletString: "walletString",
			},
			wantErr: true,
			errMsg:  "invalid character 'w' looking for beginning of value",
		},
		{
			name: "Test_Success",
			parameters: parameters{
				walletString: walletString,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			mockWalletCallback := MockTransactionCallback{}
			mockWalletCallback.On("OnTransactionComplete", &Transaction{}, 0).Return()
			got, err := NewMSTransaction(tt.parameters.walletString, mockWalletCallback)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, err.Error())
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
			require.EqualValues(got.txnStatus, StatusUnknown)
			require.EqualValues(got.verifyStatus, StatusUnknown)
		})
	}
}

func TestVerifyContentHash(t *testing.T) {
	var mockClient = mocks.HttpClient{}
	util.Client = &mockClient
	type parameters struct {
		signerwalletstr string
	}
	tests := []struct {
		name       string
		parameters parameters
		setup      func(*testing.T)
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Test_Error_Decode",
			parameters: parameters{
				signerwalletstr: "{error}",
			},
			wantErr: true,
			errMsg:  "metaTxnData_decode_error: Unable to decode metaTxnData json",
		},
		{
			name: "Test_Error_Unable_To_Fetch_Txn_Details",
			parameters: parameters{
				signerwalletstr: "{}",
			},
			wantErr: true,
			errMsg:  "fetch_txm_details: Unable to fetch txn details",
		},
		{
			name: "Test_Error_Transaction_Data",
			parameters: parameters{
				signerwalletstr: `{"Metadata":{"Hash":"mockhash"}}`,
			},
			setup: func(t *testing.T) {
				blockchain.SetSharders([]string{"TestVerifyContentHash"})
				mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return strings.HasPrefix(req.URL.Path, "TestVerifyContentHash")
				})).Return(&http.Response{
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"txn":{"block_hash":"mockhash","signature":"mocksignature","transaction_data":"{\"MetaData\":{\"Hash\":\"mockhash\"}}"}}`))),
					StatusCode: http.StatusOK,
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if tt.setup != nil {
				tt.setup(t)
			}
			got, err := VerifyContentHash(tt.parameters.signerwalletstr)
			require.EqualValues(tt.wantErr, err != nil)
			if err != nil {
				require.EqualValues(tt.errMsg, err.Error())
				return
			}
			require.NoErrorf(err, "unexpected error: %v", err)
			require.True(got)
		})
	}
}

func TestVestingTrigger(t *testing.T) {
	t.Run("Test_Vesting_Trigger_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestVestingTrigger"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.VestingTrigger("poolID")

		require.Nil(t, err)
	})
}

func TestVestingStop(t *testing.T) {
	t.Run("Test_Vesting_Stop_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestVestingStop"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
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
	t.Run("Test_Vesting_Unlock_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestVestingUnlock"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
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
	t.Run("Test_Vesting_Add_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestVestingAdd"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
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
	t.Run("Test_Vesting_Update_Config_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestVestingDelete"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
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
	t.Run("Test_Vesting_Update_Config_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestVestingUpdateConfig"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
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
	t.Run("Test_Miner_SCSettings_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestMinerSCSettings"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.MinerSCSettings(&MinerSCMinerInfo{})

		require.NoError(t, err)
	})
}

func TestVestinggUpdateConfig(t *testing.T) {
	t.Run("Test_Vesting_Update_Config_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestVestingUpdateConfig"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.VestingUpdateConfig(&VestingSCConfig{})

		require.NoError(t, err)
	})
}

func TestMinerSCLock(t *testing.T) {
	t.Run("Test_Miner_SCLock_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestMinerSCLock"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusAuthTimeout

		err := trans.MinerSCLock("nodeID", 1)

		require.NoError(t, err)
	})
}

func TestMienrSCUnlock(t *testing.T) {
	t.Run("Test_Mienr_SCUnlock_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestMienrSCUnlock"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
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
	t.Run("Test_Unlock_Tokens_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestLockTokens"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
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
	t.Run("Test_Unlock_Tokens_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestUnlockTokens"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
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
		_config.chain.Miners = []string{"TestTransaction_TestRegisterMultiSig"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.RegisterMultiSig(walletString, msw)

		require.NoError(t, err)
	})
}
func TestRegisterVote(t *testing.T) {
	t.Run("Test_Register_Vote_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestRegisterVoteTest_Register_Vote_Success"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.RegisterVote(walletString, msv)

		require.NoError(t, err)
	})

	t.Run("Test_Register_Vote_Fails", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestRegisterVoteTest_Register_Vote_Fails"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.RegisterVote("walletString", msv)

		expectedErrorMsg := "invalid character 'w' looking for beginning of value"
		require.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}

func TestFinalizeAllocation(t *testing.T) {
	t.Run("Test_Finalize_Allocation_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestFinalizeAllocation"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
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
	t.Run("Test_Cancel_Allocation_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestCancelAllocation"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
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
	t.Run("Test_Create_Allocation_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestCreateAllocation"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
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
	t.Run("Test_Create_Read_Pool_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestCreateReadPool"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
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
	t.Run("Test_Read_Pool_Lock_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestReadPoolLock"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
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
	t.Run("Test_Read_Pool_Unlock_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestReadPoolUnlock"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
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
	t.Run("Test_Stake_Pool_Lock_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestStakePoolLock"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
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
	t.Run("Test_Stake_Pool_Unlock_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestStakePoolUnlock"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
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
	t.Run("Stake_Pool_Pay_Interests_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestStakePoolPayInterests"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
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
	t.Run("Update_Blobber_Settings_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestUpdateBlobberSettings"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
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
	t.Run("Update_Allocation_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestUpdateAllocation"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
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
	t.Run("Write_Pool_Lock_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestWritePoolLock"}
		setupMockSubmitTxn()
		setupMockHttpResponse([]byte(""))
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
	t.Run("Write_Pool_Unlock_Success", func(t *testing.T) {
		_config.chain.Miners = []string{"TestTransaction_TestWritePoolUnlock"}
		setupMockSubmitTxn()
		var mockClient = mocks.HttpClient{}
		util.Client = &mockClient
		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			fmt.Println("**********", req.URL.Path)
			return strings.HasPrefix(req.URL.Path, "TestTransaction_TestWritePoolUnlock")
		})).Return(&http.Response{
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
			StatusCode: http.StatusOK,
		}, nil)
		trans := &Transaction{
			txnOut: "test",
			txn:    &transaction.Transaction{},
		}
		trans.txnStatus = StatusUnknown

		err := trans.WritePoolUnlock("poolID", 100)

		require.NoError(t, err)
	})
}
