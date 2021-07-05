package zcncore

import (
	"bytes"
	"encoding/json"
	"fmt"
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
			TransactionType: 0,
		}
	)
	mockTxn.ComputeHashData()
	fmt.Println("=============1", mockTxn.Hash)

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
	_config.authUrl = "TestSend"

	t.Run("Test Send", func(t *testing.T) {
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

		_config.chain.Miners = []string{"Send1", ""}

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			if strings.HasPrefix(req.URL.Path, "/dns") || strings.HasPrefix(req.URL.Path, "Send1") || strings.HasPrefix(req.URL.Path, "/v1/") {
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
		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			if strings.HasPrefix(req.URL.Path, "/dns") || strings.HasPrefix(req.URL.Path, "TestTransactionAuthCancelAllocation1") || strings.HasPrefix(req.URL.Path, "/v1/") {
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

		err := ta.Send(mockTxn.ToClientID, mockTxn.Value, mockTxn.TransactionData)
		require.NoError(t, err)
	})
}
//RUNOK
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
	fmt.Println("=============1", mockTxn.Hash)

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

		_config.chain.Miners = []string{"TestStoreData1", ""}

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			if strings.HasPrefix(req.URL.Path, "/dns") || strings.HasPrefix(req.URL.Path, "TestStoreData1") || strings.HasPrefix(req.URL.Path, "/v1/") {
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
		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			if strings.HasPrefix(req.URL.Path, "/dns") || strings.HasPrefix(req.URL.Path, "TestTransactionAuthCancelAllocation1") || strings.HasPrefix(req.URL.Path, "/v1/") {
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

//RUNOK
func TestTransactionAuthExecuteFaucetSCWallet(t *testing.T) {
	var (
		mockWalletString = `{"client_id":"679b06b89fc418cfe7f8fc908137795de8b7777e9324901432acce4781031c93","client_key":"2c2aaca87c9d80108c4d5dc27fc8eefc57be98af55d26a548ebf92a86cd90615d19d715a9ed6d009798877189babf405384a2980e102ce72f824890b20f8ce1e","keys":[{"public_key":"2c2aaca87c9d80108c4d5dc27fc8eefc57be98af55d26a548ebf92a86cd90615d19d715a9ed6d009798877189babf405384a2980e102ce72f824890b20f8ce1e","private_key":"mock private key"}],"mnemonics":"bamboo list citizen release bronze duck woman moment cart crucial extra hip witness mixture flash into priority length pattern deposit title exhaust flush addict","version":"1.0","date_created":"2021-06-15 11:11:40.306922176 +0700 +07 m=+1.187131283"}`

		mockPublicKey    = "mock public key"
		mockPrivateKey   = "mock private key"
		mockSignature    = "mock signature"
		mockClientID     = "mock client id"
		mockToClientID   = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d3"
		mockTxnData      = `{"name":"GET","input":"dGVzdA=="}`
		mockCreationDate = int64(1625030157)
		mockValue        = int64(0)
		mockTxn          = &transaction.Transaction{
			PublicKey:       mockPublicKey,
			ClientID:        mockClientID,
			TransactionData: mockTxnData,
			CreationDate:    mockCreationDate,
			ToClientID:      mockToClientID,
			Value:           mockValue,
			Signature:       mockSignature,
			TransactionType: 1000,
		}
	)
	mockTxn.ComputeHashData()
	fmt.Println("=============1", mockTxn.Hash)

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
	_config.authUrl = "TestExecuteFaucetSCWallet"

	t.Run("Test Execute Faucet SC Wallet", func(t *testing.T) {
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

		_config.chain.Miners = []string{"ExecuteFaucetSCWallet1", ""}

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			if strings.HasPrefix(req.URL.Path, "/dns") || strings.HasPrefix(req.URL.Path, "ExecuteFaucetSCWallet1") || strings.HasPrefix(req.URL.Path, "/v1/") {
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

		err := ta.ExecuteFaucetSCWallet(mockWalletString, "GET", []byte("test"))
		require.NoError(t, err)
	})
}

//RUNOK
func TestExecuteSmartContract(t *testing.T) {
	var (
		mockPublicKey    = "mock public key"
		mockPrivateKey   = "mock private key"
		mockSignature    = "mock signature"
		mockClientID     = "mock client id"
		mockToClientID   = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d3"
		mockTxnData      = `{"name":"GET","input":{}}`
		mockCreationDate = int64(1625030157)
		mockValue        = int64(1)
		mockTxn          = &transaction.Transaction{
			PublicKey:       mockPublicKey,
			ClientID:        mockClientID,
			TransactionData: mockTxnData,
			CreationDate:    mockCreationDate,
			ToClientID:      mockToClientID,
			Value:           mockValue,
			Signature:       mockSignature,
			TransactionType: 1000,
		}
	)
	mockTxn.ComputeHashData()
	fmt.Println("=============1", mockTxn.Hash)

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
	_config.authUrl = "TestExecuteSmartContract"

	t.Run("Test Execute Smart Contract", func(t *testing.T) {
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

		_config.chain.Miners = []string{"ExecuteSmartContract1", ""}

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			if strings.HasPrefix(req.URL.Path, "/dns") || strings.HasPrefix(req.URL.Path, "ExecuteSmartContract1") || strings.HasPrefix(req.URL.Path, "/v1/") {
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
		scData := make(map[string]interface{})
		out, err := json.Marshal(scData)
		require.NoError(t,err)
		err = ta.ExecuteSmartContract(mockToClientID,"GET",string(out),1)
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
// func TestExecuteSmartContract(t *testing.T) {
// 	t.Run("Test Execute Smart Contract", func(t *testing.T) {
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{},
// 			},
// 		}

// 		err := ta.ExecuteSmartContract("address", "GET", "{}", 1)
// 		require.NoError(t, err)
// 	})
// }
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
// 	var (
// 		mockPublicKey       = "mock public key"
// 		mockPrivateKey      = "mock private key"
// 		mockSignature       = "mock signature"
// 		mockClientID        = "mock client id"
// 		mockToClientID      = "cf8d0df9bd8cc637a4ff4e792ffe3686da6220c45f0e1103baa609f3f1751ef4"
// 		mockTxnData         = `{"name":"unlock","input":{"pool_id":"mock pool id"}}`
// 		mockCreationDate    = int64(1625030157)
// 		mockValue           = int64(0)
// 		mockTransactionType = int(1000)
// 		mockTxn             = &transaction.Transaction{
// 			PublicKey:       mockPublicKey,
// 			ClientID:        mockClientID,
// 			TransactionData: mockTxnData,
// 			CreationDate:    mockCreationDate,
// 			ToClientID:      mockToClientID,
// 			Value:           mockValue,
// 			Signature:       mockSignature,
// 			TransactionType: mockTransactionType,
// 		}
// 	)
// 	mockTxn.ComputeHashData()
// 	fmt.Println("=============1", mockTxn.Hash)

// 	_config.wallet = zcncrypto.Wallet{
// 		ClientID: mockClientID,
// 		Keys: []zcncrypto.KeyPair{
// 			{
// 				PublicKey:  mockPublicKey,
// 				PrivateKey: mockPrivateKey,
// 			},
// 		},
// 	}
// 	_config.chain.SignatureScheme = "bls0chain"
// 	_config.authUrl = "UnlockTokens"

// 	t.Run("Test Unlock Tokens", func(t *testing.T) {
// 		var mockClient = zcnmocks.HttpClient{}
// 		util.Client = &mockClient

// 		mockSignatureScheme := &zcncryptomock.SignatureScheme{}
// 		mockSignatureScheme.On("SetPrivateKey", mockPrivateKey).Return(nil)
// 		mockSignatureScheme.On("SetPublicKey", mockPublicKey).Return(nil)
// 		mockSignatureScheme.On("Sign", mockTxn.Hash).Return(mockSignature, nil)
// 		mockSignatureScheme.On("Verify", mockSignature, mockTxn.Hash).Return(true, nil)
// 		mockSignatureScheme.On("Add", mockTxn.Signature, mockTxn.Hash).Return(mockSignature, nil)
// 		setupSignatureSchemeMock(mockSignatureScheme)

// 		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
// 			if req.Method == "POST" && req.URL.Path == _config.authUrl+"/transaction" {
// 				require.EqualValues(t, "application/json", strings.Split(req.Header.Get("Content-Type"), ";")[0])
// 				defer req.Body.Close()
// 				body, err := ioutil.ReadAll(req.Body)
// 				require.NoError(t, err, "ioutil.ReadAll(req.Body)")
// 				var reqTxn *transaction.Transaction
// 				err = json.Unmarshal(body, &reqTxn)
// 				require.NoError(t, err, "json.Unmarshal(body, &reqTxn)")
// 				require.EqualValues(t, mockTxn, reqTxn)
// 				return true
// 			}
// 			return false
// 		})).Return(&http.Response{
// 			Body: func() io.ReadCloser {
// 				jsonFR, err := json.Marshal(mockTxn)
// 				require.NoError(t, err, "json.Marshal(mockTxn)")
// 				return ioutil.NopCloser(bytes.NewReader(jsonFR))
// 			}(),
// 			StatusCode: http.StatusOK,
// 		}, nil)

// 		_config.chain.Miners = []string{"UnlockTokens1", ""}

// 		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
// 			if strings.HasPrefix(req.URL.Path, "/dns") || strings.HasPrefix(req.URL.Path, "UnlockTokens1") || strings.HasPrefix(req.URL.Path, "/v1/") {
// 				return true
// 			}
// 			return false
// 		})).Return(&http.Response{
// 			Body: func() io.ReadCloser {
// 				jsonFR, err := json.Marshal(mockTxn)
// 				require.NoError(t, err, "json.Marshal(mockTxn)")
// 				return ioutil.NopCloser(bytes.NewReader(jsonFR))
// 			}(),
// 			StatusCode: http.StatusOK,
// 		}, nil)
// 		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
// 			if strings.HasPrefix(req.URL.Path, "/dns") || strings.HasPrefix(req.URL.Path, "TestTransactionAuthCancelAllocation1") || strings.HasPrefix(req.URL.Path, "/v1/") {
// 				return true
// 			}
// 			return false
// 		})).Return(&http.Response{
// 			Body: func() io.ReadCloser {
// 				jsonFR, err := json.Marshal(mockTxn)
// 				require.NoError(t, err, "json.Marshal(mockTxn)")
// 				return ioutil.NopCloser(bytes.NewReader(jsonFR))
// 			}(),
// 			StatusCode: http.StatusOK,
// 		}, nil)
// 		ta := &TransactionWithAuth{
// 			t: &Transaction{
// 				txn: &transaction.Transaction{
// 					Hash:            mockTxn.Hash,
// 					ClientID:        mockClientID,
// 					PublicKey:       mockPublicKey,
// 					ToClientID:      mockToClientID,
// 					CreationDate:    mockCreationDate,
// 					Value:           mockValue,
// 					TransactionData: mockTxnData,
// 					TransactionType: mockTxn.TransactionType,
// 					Signature:       mockSignature,
// 				},
// 			},
// 		}

// 		err := ta.UnlockTokens("mock pool id")
// 		require.NoError(t, err)
// 	})
// }
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
			TransactionData: `{"name":"finalize_allocation","input":{"allocation_id":"mock pool id"}}`,
			CreationDate:    mockCreationDate,
			ToClientID:      `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7`,
			Value:           0,
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
	_config.authUrl = "FinalizeAllocation"

	t.Run("Test Finalize Allocation", func(t *testing.T) {
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
			if strings.HasPrefix(req.URL.Path, _config.authUrl) || strings.HasPrefix(req.URL.Path, "FinalizeAllocation1") || strings.HasPrefix(req.URL.Path, "/v1") {
				// require.EqualValues(t, "application/json", strings.Split(req.Header.Get("Content-Type"), ";")[0])
				// defer req.Body.Close()
				// body, err := ioutil.ReadAll(req.Body)
				// require.NoError(t, err, "ioutil.ReadAll(req.Body)")
				// var reqTxn *transaction.Transaction
				// err = json.Unmarshal(body, &reqTxn)
				// require.NoError(t, err, "json.Unmarshal(body, &reqTxn)")
				// require.EqualValues(t, mockTxn, reqTxn)
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

		_config.chain.Miners = []string{"FinalizeAllocation1", ""}

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			if strings.HasPrefix(req.URL.Path, "/dns") || strings.HasPrefix(req.URL.Path, "FinalizeAllocation1") || strings.HasPrefix(req.URL.Path, "/v1/") {
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
		ta := &TransactionWithAuth{
			t: &Transaction{
				txn: &transaction.Transaction{
					ClientID:        mockClientID,
					PublicKey:       mockPublicKey,
					ToClientID:      mockClientID,
					CreationDate:    mockCreationDate,
					Value:           mockValue,
					TransactionData: mockTxnData,
				},
			},
		}

		err := ta.FinalizeAllocation("mock pool id", 1)
		require.NoError(t, err)
	})
}
//RUNOK
func TestTransactionAuthCancelAllocation(t *testing.T) {
	var (
		mockPublicKey    = "mock public key"
		mockPrivateKey   = "mock private key"
		mockSignature    = "mock signature"
		mockClientID     = "mock client id"
		mockToClientID   = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7"
		mockTxnData      = `{"name":"cancel_allocation","input":{"allocation_id":"mock allocation id"}}`
		mockCreationDate = int64(1625030157)
		mockValue        = int64(0)
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
	_config.authUrl = "TestTransactionAuthCancelAllocation"
	t.Run("Test Transaction Auth Cancel Allocation", func(t *testing.T) {
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
				jsonFR, err := json.Marshal(mockTxn)
				require.NoError(t, err, "json.Marshal(mockTxn)")
				return ioutil.NopCloser(bytes.NewReader(jsonFR))
			}(),
			StatusCode: http.StatusOK,
		}, nil)

		_config.chain.Miners = []string{"TestTransactionAuthCancelAllocation1", ""}

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			if strings.HasPrefix(req.URL.Path, "/dns") || strings.HasPrefix(req.URL.Path, "TestTransactionAuthCancelAllocation1") || strings.HasPrefix(req.URL.Path, "/v1/") {
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

		resp := ta.CancelAllocation("mock allocation id", 1)
		require.NoError(t, resp)
	})
}
//RUNOK
func TestTransactionAuthVestingTrigger(t *testing.T) {
	var (
		mockPublicKey    = "mock public key"
		mockPrivateKey   = "mock private key"
		mockSignature    = "mock signature"
		mockClientID     = "mock client id"
		mockToClientID   = "2bba5b05949ea59c80aed3ac3474d7379d3be737e8eb5a968c52295e48333ead"
		mockTxnData      = `{"name":"trigger","input":{"pool_id":"mock pool id"}}`
		mockCreationDate = int64(1625030157)
		mockValue        = int64(0)
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
	_config.authUrl = "TestVestingTrigger"
	t.Run("Test Vesting Trigger", func(t *testing.T) {
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
				jsonFR, err := json.Marshal(mockTxn)
				require.NoError(t, err, "json.Marshal(mockTxn)")
				return ioutil.NopCloser(bytes.NewReader(jsonFR))
			}(),
			StatusCode: http.StatusOK,
		}, nil)

		_config.chain.Miners = []string{"TestVestingTrigger1", ""}

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			if strings.HasPrefix(req.URL.Path, "/dns") || strings.HasPrefix(req.URL.Path, "TestVestingTrigger1") || strings.HasPrefix(req.URL.Path, "/v1/") {
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

		resp := ta.VestingTrigger("mock pool id")
		require.NoError(t, resp)
	})
}
//RUNOK
func TestTransactionAuthVestingStop(t *testing.T) {
	var (
		mockPublicKey    = "mock public key"
		mockPrivateKey   = "mock private key"
		mockSignature    = "mock signature"
		mockClientID     = "mock client id"
		mockToClientID   = "2bba5b05949ea59c80aed3ac3474d7379d3be737e8eb5a968c52295e48333ead"
		mockTxnData      = `{"name":"stop","input":{"pool_id":"","destination":""}}`
		mockCreationDate = int64(1625030157)
		mockValue        = int64(0)
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
	_config.authUrl = "TestVestingStop"
	t.Run("TestVesting Stop", func(t *testing.T) {
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
				jsonFR, err := json.Marshal(mockTxn)
				require.NoError(t, err, "json.Marshal(mockTxn)")
				return ioutil.NopCloser(bytes.NewReader(jsonFR))
			}(),
			StatusCode: http.StatusOK,
		}, nil)

		_config.chain.Miners = []string{"VestingStop1", ""}

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			if strings.HasPrefix(req.URL.Path, "/dns") || strings.HasPrefix(req.URL.Path, "VestingStop1") || strings.HasPrefix(req.URL.Path, "/v1/") {
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

		resp := ta.VestingStop(&VestingStopRequest{})
		require.NoError(t, resp)
	})
}
//RUNOK
func TestTransactionAuthVestingUnlock(t *testing.T) {
	var (
		mockPublicKey    = "mock public key"
		mockPrivateKey   = "mock private key"
		mockSignature    = "mock signature"
		mockClientID     = "mock client id"
		mockToClientID   = "2bba5b05949ea59c80aed3ac3474d7379d3be737e8eb5a968c52295e48333ead"
		mockTxnData      = `{"name":"unlock","input":{"pool_id":"mock pool id"}}`
		mockCreationDate = int64(1625030157)
		mockValue        = int64(0)
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
	_config.authUrl = "VestingUnlock"
	t.Run("Test Vesting Unlock", func(t *testing.T) {
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
				jsonFR, err := json.Marshal(mockTxn)
				require.NoError(t, err, "json.Marshal(mockTxn)")
				return ioutil.NopCloser(bytes.NewReader(jsonFR))
			}(),
			StatusCode: http.StatusOK,
		}, nil)

		_config.chain.Miners = []string{"VestingUnlock1", ""}

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			if strings.HasPrefix(req.URL.Path, "/dns") || strings.HasPrefix(req.URL.Path, "VestingUnlock1") || strings.HasPrefix(req.URL.Path, "/v1/") {
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

		resp := ta.VestingUnlock("mock pool id")
		require.NoError(t, resp)
	})
}
//RUNOK
func TestTransactionAuthVestingAdd(t *testing.T) {
	var (
		mockPublicKey    = "mock public key"
		mockPrivateKey   = "mock private key"
		mockSignature    = "mock signature"
		mockClientID     = "mock client id"
		mockToClientID   = "2bba5b05949ea59c80aed3ac3474d7379d3be737e8eb5a968c52295e48333ead"
		mockTxnData      = `{"name":"add","input":{"description":"","start_time":0,"duration":0,"destinations":null}}`
		mockCreationDate = int64(1625030157)
		mockValue        = int64(1)
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
	_config.authUrl = "VestingAdd"
	t.Run("Test Vesting Add", func(t *testing.T) {
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
				jsonFR, err := json.Marshal(mockTxn)
				require.NoError(t, err, "json.Marshal(mockTxn)")
				return ioutil.NopCloser(bytes.NewReader(jsonFR))
			}(),
			StatusCode: http.StatusOK,
		}, nil)

		_config.chain.Miners = []string{"VestingAdd1", ""}

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			if strings.HasPrefix(req.URL.Path, "/dns") || strings.HasPrefix(req.URL.Path, "VestingAdd1") || strings.HasPrefix(req.URL.Path, "/v1/") {
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

		resp := ta.VestingAdd(&VestingAddRequest{},1)
		require.NoError(t, resp)
	})
}
//RUNOK
func TestTransactionAuthVestingDelete(t *testing.T) {
	var (
		mockPublicKey    = "mock public key"
		mockPrivateKey   = "mock private key"
		mockSignature    = "mock signature"
		mockClientID     = "mock client id"
		mockToClientID   = "2bba5b05949ea59c80aed3ac3474d7379d3be737e8eb5a968c52295e48333ead"
		mockTxnData      = `{"name":"delete","input":{"pool_id":"mock pool id"}}`
		mockCreationDate = int64(1625030157)
		mockValue        = int64(0)
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
	_config.authUrl = "VestingDelete"
	t.Run("Test Vesting Delete", func(t *testing.T) {
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
				jsonFR, err := json.Marshal(mockTxn)
				require.NoError(t, err, "json.Marshal(mockTxn)")
				return ioutil.NopCloser(bytes.NewReader(jsonFR))
			}(),
			StatusCode: http.StatusOK,
		}, nil)

		_config.chain.Miners = []string{"VestingDelete1", ""}

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			if strings.HasPrefix(req.URL.Path, "/dns") || strings.HasPrefix(req.URL.Path, "VestingDelete1") || strings.HasPrefix(req.URL.Path, "/v1/") {
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

		resp := ta.VestingDelete("mock pool id")
		require.NoError(t, resp)
	})
}
//RUNOK
func TestTransactionAuthVestingUpdateConfig(t *testing.T) {
	var (
		mockPublicKey    = "mock public key"
		mockPrivateKey   = "mock private key"
		mockSignature    = "mock signature"
		mockClientID     = "mock client id"
		mockToClientID   = "2bba5b05949ea59c80aed3ac3474d7379d3be737e8eb5a968c52295e48333ead"
		mockTxnData      = `{"name":"update_config","input":{"min_lock":0,"min_duration":0,"max_duration":0,"max_destinations":0,"max_description_length":0}}`
		mockCreationDate = int64(1625030157)
		mockValue        = int64(0)
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
	_config.authUrl = "TestVestingUpdateConfig"
	t.Run("Test Vesting Update Config", func(t *testing.T) {
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
				jsonFR, err := json.Marshal(mockTxn)
				require.NoError(t, err, "json.Marshal(mockTxn)")
				return ioutil.NopCloser(bytes.NewReader(jsonFR))
			}(),
			StatusCode: http.StatusOK,
		}, nil)

		_config.chain.Miners = []string{"VestingUpdateConfig1", ""}

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			if strings.HasPrefix(req.URL.Path, "/dns") || strings.HasPrefix(req.URL.Path, "VestingUpdateConfig1") || strings.HasPrefix(req.URL.Path, "/v1/") {
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

		resp := ta.VestingUpdateConfig(&VestingSCConfig{})
		require.NoError(t, resp)
	})
}
//RUNOK
func TestTransactionAuthMinerSCSettings(t *testing.T) {
	var (
		mockPublicKey    = "mock public key"
		mockPrivateKey   = "mock private key"
		mockSignature    = "mock signature"
		mockClientID     = "mock client id"
		mockToClientID   = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d9"
		mockTxnData      = `{"name":"update_settings","input":{"simple_miner":null,"pending":null,"active":null,"deleting":null}}`
		mockCreationDate = int64(1625030157)
		mockValue        = int64(0)
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
	_config.authUrl = "MinerSCSettings"
	t.Run("Test Vesting Delete", func(t *testing.T) {
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
				jsonFR, err := json.Marshal(mockTxn)
				require.NoError(t, err, "json.Marshal(mockTxn)")
				return ioutil.NopCloser(bytes.NewReader(jsonFR))
			}(),
			StatusCode: http.StatusOK,
		}, nil)

		_config.chain.Miners = []string{"MinerSCSettings1", ""}

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			if strings.HasPrefix(req.URL.Path, "/dns") || strings.HasPrefix(req.URL.Path, "MinerSCSettings1") || strings.HasPrefix(req.URL.Path, "/v1/") {
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

		resp := ta.MinerSCSettings(&MinerSCMinerInfo{})
		require.NoError(t, resp)
	})
}
//RUNOK
func TestTransactionAuthMinerSCLock(t *testing.T) {
	var (
		mockPublicKey    = "mock public key"
		mockPrivateKey   = "mock private key"
		mockSignature    = "mock signature"
		mockClientID     = "mock client id"
		mockToClientID   = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d9"
		mockTxnData      = `{"name":"addToDelegatePool","input":{"id":"mock miner id"}}`
		mockCreationDate = int64(1625030157)
		mockValue        = int64(1)
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
	_config.authUrl = "MinerSCLock"
	t.Run("Test Vesting Delete", func(t *testing.T) {
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
				jsonFR, err := json.Marshal(mockTxn)
				require.NoError(t, err, "json.Marshal(mockTxn)")
				return ioutil.NopCloser(bytes.NewReader(jsonFR))
			}(),
			StatusCode: http.StatusOK,
		}, nil)

		_config.chain.Miners = []string{"MinerSCSettings1", ""}

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			if strings.HasPrefix(req.URL.Path, "/dns") || strings.HasPrefix(req.URL.Path, "MinerSCSettings1") || strings.HasPrefix(req.URL.Path, "/v1/") {
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

		resp := ta.MinerSCLock("mock miner id", 1)
		require.NoError(t, resp)
	})
}
//RUNOK
func TestTransactionAuthMienrSCUnlock(t *testing.T) {
	var (
		mockPublicKey    = "mock public key"
		mockPrivateKey   = "mock private key"
		mockSignature    = "mock signature"
		mockClientID     = "mock client id"
		mockToClientID   = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d9"
		mockTxnData      = `{"name":"deleteFromDelegatePool","input":{"id":"mock node id","pool_id":"mock pool id"}}`
		mockCreationDate = int64(1625030157)
		mockValue        = int64(0)
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
	_config.authUrl = "MienrSCUnlock"
	t.Run("Test Vesting Delete", func(t *testing.T) {
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
				jsonFR, err := json.Marshal(mockTxn)
				require.NoError(t, err, "json.Marshal(mockTxn)")
				return ioutil.NopCloser(bytes.NewReader(jsonFR))
			}(),
			StatusCode: http.StatusOK,
		}, nil)

		_config.chain.Miners = []string{"MienrSCUnlock1", ""}

		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			if strings.HasPrefix(req.URL.Path, "/dns") || strings.HasPrefix(req.URL.Path, "MienrSCUnlock1") || strings.HasPrefix(req.URL.Path, "/v1/") {
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

		resp := ta.MienrSCUnlock("mock node id", "mock pool id")
		require.NoError(t, resp)
	})
}
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
