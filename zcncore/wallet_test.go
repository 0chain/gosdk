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
	"github.com/0chain/gosdk/zboxcore/mocks"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCheckSdkInit(t *testing.T) {
	t.Run("Test check Sdk Init SDK not initialized", func(t *testing.T) {
		_config.isConfigured = false
		err := checkSdkInit()
		expectedErrorMsg := "SDK not initialized"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test check Sdk Init", func(t *testing.T) {
		_config.isConfigured = true
		err := checkSdkInit()
		require.NoError(t, err)
	})
}
func TestCheckWalletConfig(t *testing.T) {
	t.Run("Test check Wallet Config", func(t *testing.T) {
		_config.isConfigured = false
		err := checkWalletConfig()
		expectedErrorMsg := "wallet info not found. set wallet info."
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test check Sdk Init", func(t *testing.T) {
		_config.isConfigured = true
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID
		err := checkWalletConfig()
		require.NoError(t, err)
	})
}
func TestCheckConfig(t *testing.T) {
	t.Run("Test check Config", func(t *testing.T) {
		_config.isConfigured = false
		err := checkConfig()
		expectedErrorMsg := "SDK not initialized"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test check Config", func(t *testing.T) {
		_config.isConfigured = true
		err := checkConfig()
		require.NoError(t, err)
	})
}
func TestAssertConfig(t *testing.T) {
	t.Run("Test assert Config", func(t *testing.T) {
		_config.isConfigured = false
		assertConfig()
	})
}
func TestgetMinMinersSubmit(t *testing.T) {
	t.Run("Test get Min Miners Submit", func(t *testing.T) {
		_config.isConfigured = false
		resp := getMinMinersSubmit()
		require.Equal(t, 1, resp)
	})
}
func TestGetMinShardersVerify(t *testing.T) {
	t.Run("Test Get Min Sharders Verify", func(t *testing.T) {
		_config.isConfigured = false
		resp := GetMinShardersVerify()
		require.Equal(t, 1, resp)
	})
}
func TestGetMinRequiredChainLength(t *testing.T) {
	t.Run("Test get Min Required Chain Length", func(t *testing.T) {
		_config.isConfigured = false
		resp := getMinRequiredChainLength()
		require.Equal(t, int64(3), resp)
	})
}
func TestCalculateMinRequired(t *testing.T) {
	t.Run("Test calculate Min Required", func(t *testing.T) {
		_config.isConfigured = false
		resp := calculateMinRequired(1, 1)
		require.Equal(t, 1, resp)
	})
}
func TestGetVersion(t *testing.T) {
	t.Run("Test calculate Min Required", func(t *testing.T) {
		_config.isConfigured = false
		resp := GetVersion()
		require.Equal(t, "v1.2.6", resp)
	})
}
func TestSetLogLevel(t *testing.T) {
	t.Run("Test Set Log Level", func(t *testing.T) {
		_config.isConfigured = false
		SetLogLevel(1)
	})
}
func TestSetLogFile(t *testing.T) {
	t.Run("Test Set Log Level", func(t *testing.T) {
		_config.isConfigured = false
		SetLogFile("logFile", true)
	})
}
func TestCloseLog(t *testing.T) {
	t.Run("Test Close Log", func(t *testing.T) {
		_config.isConfigured = false
		CloseLog()
	})
}

func TestInit(t *testing.T) {
	t.Run("Test Init", func(t *testing.T) {
		var mockClient = mocks.HttpClient{}

		util.Client = &mockClient
		_config.chain.BlockWorker = "TestInit"
		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			if strings.HasPrefix(req.URL.Path, "TestInit") || strings.HasPrefix(req.URL.Path, "/dns/network") || strings.HasPrefix(req.URL.Path, "/v1/transaction/put") {
				return true
			}
			return false

		})).Return(&http.Response{
			Body: func() io.ReadCloser {
				jsonFR, err := json.Marshal(&Network{
					Miners:   []string{""},
					Sharders: []string{""},
				})
				require.NoError(t, err)
				return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
			}(),
			StatusCode: http.StatusOK,
		}, nil)

		_config.isConfigured = false
		jsonFR, err := json.Marshal(&ChainConfig{
			ChainID:         "",
			BlockWorker:     "",
			SignatureScheme: "bls0chain",
		})
		require.NoError(t, err)

		err = Init(string(jsonFR))
		require.NoError(t, err)
	})
}

func TestGetBalanceWallet(t *testing.T) {
	t.Run("Test Init", func(t *testing.T) {
		var mockClient = mocks.HttpClient{}

		util.Client = &mockClient
		_config.chain.BlockWorker = "TestGetBalanceWallet"
		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			if strings.HasPrefix(req.URL.Path, "TestGetBalanceWallet") || strings.HasPrefix(req.URL.Path, "/dns/network") || strings.HasPrefix(req.URL.Path, "/v1/transaction/put") {
				return true
			}
			return false

		})).Return(&http.Response{
			Body: func() io.ReadCloser {
				jsonFR, err := json.Marshal(&Network{
					Miners:   []string{""},
					Sharders: []string{""},
				})
				require.NoError(t, err)
				return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
			}(),
			StatusCode: http.StatusOK,
		}, nil)
		var mockGetBalanceCallback = mocks.GetBalanceCallback{}

		mockGetBalanceCallback.On("OnBalanceAvailable", 2, int64(0), "").Return()
	

		err := GetBalanceWallet(walletString,mockGetBalanceCallback)
		require.NoError(t, err)
	})
}
func TestWithChainID(t *testing.T) {
	t.Run("Test With Chain ID", func(t *testing.T) {
		_config.isConfigured = false
		resp := WithChainID("ID")
		require.NotNil(t, resp)
	})
}
func TestWithMinSubmit(t *testing.T) {
	t.Run("Test With Min Submit", func(t *testing.T) {
		resp := WithMinSubmit(1)
		require.NotNil(t, resp)
	})
}
func TestWithMinConfirmation(t *testing.T) {
	t.Run("Test With Min Confirmation", func(t *testing.T) {
		resp := WithMinConfirmation(1)
		require.NotNil(t, resp)
	})
}
func TestWithConfirmationChainLength(t *testing.T) {
	t.Run("Test With Min Confirmation", func(t *testing.T) {
		resp := WithConfirmationChainLength(1)
		require.NotNil(t, resp)
	})
}

// func TestInitZCNSDK(t *testing.T) {
// 	t.Run("Test Init ZCN SDK", func(t *testing.T) {
// 		resp := InitZCNSDK("", "bls0chain")
// 		require.Nil(t, resp)
// 	})
// }
func TestGetNetwork(t *testing.T) {
	t.Run("Test Get Network", func(t *testing.T) {
		resp := GetNetwork()
		require.NotNil(t, resp)
	})
}
func TestSetNetwork(t *testing.T) {
	t.Run("Test Set Network", func(t *testing.T) {
		SetNetwork([]string{"", ""}, []string{"", ""})

	})
}
func TestGetNetworkJSON(t *testing.T) {
	t.Run("Test Get Net work JSON", func(t *testing.T) {
		resp := GetNetworkJSON()
		jsonFR, err := json.Marshal(&Network{
			Miners:   []string{"", ""},
			Sharders: []string{"", ""},
		})
		require.NoError(t, err)
		require.Equal(t, string(jsonFR), resp)
	})
}

func TestCreateWallet(t *testing.T) {
	var (
		// mockRecoverKey = "snake mixed bird cream cotton trouble small fee finger catalog measure spoon private second canal pact unable close predict dream mask delay path inflict"
		mockPublicKey    = "mock public key"
		mockPrivateKey   = "mock private key"
		mockSignature    = "mock signature"
		mockTxnData      = "mock txn data"
		mockCreationDate = int64(1625030157)
		mockValue        = int64(1)
		mockClientID     = "mock client id"

		mockTxn = &transaction.Transaction{
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

	t.Run("Test Create Wallet", func(t *testing.T) {
		var mockClient = mocks.HttpClient{}
		mockSignatureScheme := &zcncryptomock.SignatureScheme{}
		mockSignatureScheme.On("GenerateKeys").Return(&zcncrypto.Wallet{}, nil)
		mockSignatureScheme.On("SetPrivateKey", mockPrivateKey).Return(nil)
		mockSignatureScheme.On("SetPublicKey", mockPublicKey).Return(nil)
		mockSignatureScheme.On("Sign", mockTxn.Hash).Return(mockSignature, nil)
		mockSignatureScheme.On("Verify", mockSignature, mockTxn.Hash).Return(true, nil)
		mockSignatureScheme.On("Add", mockSignature, mockTxn.Hash).Return(mockSignature, nil)
		setupSignatureSchemeMock(mockSignatureScheme)
		wallet, err := json.Marshal(&zcncrypto.Wallet{})
		require.NoError(t, err)

		util.Client = &mockClient
		var mockWalletCallback = mocks.WalletCallback{}
		mockWalletCallback.On("OnWalletCreateComplete", 0, string(wallet), "").Return()
		_config.chain.Miners = []string{""}
		_config.chain.Sharders = []string{""}
		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return true
		})).Return(&http.Response{
			Body: func() io.ReadCloser {
				jsonFR, err := json.Marshal(&Network{
					Miners:   []string{""},
					Sharders: []string{""},
				})
				require.NoError(t, err)
				return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
			}(),
			StatusCode: http.StatusOK,
		}, nil)
		resp := CreateWallet(mockWalletCallback)
		require.Nil(t, resp)
	})
	t.Run("Test Create Wallet", func(t *testing.T) {

		var mockWalletCallback = mocks.WalletCallback{}
		wallet, err := json.Marshal(&zcncrypto.Wallet{})
		require.NoError(t, err)
		mockWalletCallback.On("OnWalletCreateComplete", 0, string(wallet), "").Return()

		resp := CreateWallet(mockWalletCallback)
		require.Nil(t, resp)
	})
}

func TestRecoverWallet(t *testing.T) {
	var (
		mockRecoverKey   = "snake mixed bird cream cotton trouble small fee finger catalog measure spoon private second canal pact unable close predict dream mask delay path inflict"
		mockPublicKey    = "mock public key"
		mockPrivateKey   = "mock private key"
		mockSignature    = "mock signature"
		mockTxnData      = "mock txn data"
		mockCreationDate = int64(1625030157)
		mockValue        = int64(1)
		mockClientID     = "mock client id"

		mockTxn = &transaction.Transaction{
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

	t.Run("Test Recover Wallet", func(t *testing.T) {
		var mockClient = mocks.HttpClient{}
		mockSignatureScheme := &zcncryptomock.SignatureScheme{}
		mockSignatureScheme.On("GenerateKeys").Return(&zcncrypto.Wallet{}, nil)
		mockSignatureScheme.On("RecoverKeys", mockRecoverKey).Return(&zcncrypto.Wallet{}, nil)
		mockSignatureScheme.On("SetPrivateKey", mockPrivateKey).Return(nil)
		mockSignatureScheme.On("SetPublicKey", mockPublicKey).Return(nil)
		mockSignatureScheme.On("Sign", mockTxn.Hash).Return(mockSignature, nil)
		mockSignatureScheme.On("Verify", mockSignature, mockTxn.Hash).Return(true, nil)
		mockSignatureScheme.On("Add", mockTxn.Signature, mockTxn.Hash).Return(mockSignature, nil)
		setupSignatureSchemeMock(mockSignatureScheme)
		wallet, err := json.Marshal(&zcncrypto.Wallet{})
		require.NoError(t, err)

		util.Client = &mockClient
		var mockWalletCallback = mocks.WalletCallback{}
		mockWalletCallback.On("OnWalletCreateComplete", 0, string(wallet), "").Return()
		_config.chain.Miners = []string{""}
		_config.chain.Sharders = []string{""}
		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return true
		})).Return(&http.Response{
			Body: func() io.ReadCloser {
				jsonFR, err := json.Marshal(&Network{
					Miners:   []string{""},
					Sharders: []string{""},
				})
				require.NoError(t, err)
				return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
			}(),
			StatusCode: http.StatusOK,
		}, nil)
		resp := RecoverWallet(mnemonic, mockWalletCallback)
		require.Nil(t, resp)
	})
}

func TestSplitKeys(t *testing.T) {
	t.Run("Test Recover Wallet signature key doesn't support split key", func(t *testing.T) {

		_config.chain.SignatureScheme = "fail signature"
		resp, err := SplitKeys(private_key, 1)

		expectedErrorMsg := "signature key doesn't support split key"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
		require.Equal(t, "", resp)
	})
	t.Run("Test Recover Wallet", func(t *testing.T) {
		_config.chain.SignatureScheme = "bls0chain"
		resp, err := SplitKeys(private_key, 1)

		require.NoError(t, err)
		require.NotNil(t, resp)
	})
}

func TestDecrypt(t *testing.T) {
	t.Run("Test Decrypt invalid key", func(t *testing.T) {
		resp, err := Decrypt(hash, "text")

		expectedErrorMsg := "crypto/aes: invalid key size 64"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
		require.Equal(t, "", resp)
	})
}

func TestEncrypt(t *testing.T) {
	t.Run("Test Encrypt", func(t *testing.T) {
		resp, err := Encrypt(hash, "text")

		expectedErrorMsg := "crypto/aes: invalid key size 64"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
		require.Equal(t, "", resp)
	})
}

func TestGetWritePoolInfo(t *testing.T) {
	t.Run("Test Get Write Pool Info", func(t *testing.T) {
		var mockClient = mocks.HttpClient{}

		util.Client = &mockClient

		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID
		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return strings.HasPrefix(req.URL.Path, "")
		})).Return(&http.Response{
			Body: func() io.ReadCloser {
				jsonFR, err := json.Marshal(&Network{
					Miners:   []string{""},
					Sharders: []string{""},
				})
				require.NoError(t, err)
				return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
			}(),
			StatusCode: http.StatusOK,
		}, nil)

		mockGetInfoCallback.On("OnInfoAvailable", 12, 0, "", "").Return()
		err := GetWritePoolInfo(clientID, mockGetInfoCallback)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestGetBlobber(t *testing.T) {
	t.Run("Test Get Blobber", func(t *testing.T) {
		var mockClient = mocks.HttpClient{}

		util.Client = &mockClient

		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID
		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return strings.HasPrefix(req.URL.Path, "")
		})).Return(&http.Response{
			Body: func() io.ReadCloser {
				jsonFR, err := json.Marshal(&Network{
					Miners:   []string{""},
					Sharders: []string{""},
				})
				require.NoError(t, err)
				return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
			}(),
			StatusCode: http.StatusOK,
		}, nil)

		mockGetInfoCallback.On("OnInfoAvailable", 11, 0, "", "").Return()
		err := GetBlobber("bloberID", mockGetInfoCallback)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}

func TestGetBlobbers(t *testing.T) {
	t.Run("Test Get Blobber", func(t *testing.T) {
		var mockClient = mocks.HttpClient{}

		util.Client = &mockClient

		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID
		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			if strings.HasPrefix(req.URL.Path, "/v1/screst/") || strings.HasPrefix(req.URL.Path, "/v1/client") || strings.HasPrefix(req.URL.Path, "url/_nh/whoami"){
				return true
			}
			return false
		})).Return(&http.Response{
			Body: func() io.ReadCloser {
				jsonFR, err := json.Marshal(&Network{
					Miners:   []string{""},
					Sharders: []string{""},
				})
				require.NoError(t, err)
				return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
			}(),
			StatusCode: http.StatusOK,
		}, nil)

		mockGetInfoCallback.On("OnInfoAvailable", 10, 0, "", "").Return()
		err := GetBlobbers(mockGetInfoCallback)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestGetStakePoolUserInfo(t *testing.T) {
	t.Run("Test Get Stake Pool User Info", func(t *testing.T) {

		_config.isConfigured = false
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 9, 0, "", "").Return()
		err := GetStakePoolUserInfo(clientID, mockGetInfoCallback)
		expectedErrorMsg := "SDK not initialized"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Stake Pool User Info empty clientID", func(t *testing.T) {

		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 9, 0, "", "").Return()
		err := GetStakePoolUserInfo("", mockGetInfoCallback)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Stake Pool User Info", func(t *testing.T) {

		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 9, 0, "", "").Return()
		err := GetStakePoolUserInfo(clientID, mockGetInfoCallback)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestGetStakePoolInfo(t *testing.T) {
	t.Run("Test Get Stake Pool Info checkConfig fail", func(t *testing.T) {

		_config.isConfigured = false
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 9, 0, "", "").Return()
		err := GetStakePoolInfo("blobberID", mockGetInfoCallback)
		expectedErrorMsg := "SDK not initialized"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Stake Pool Info", func(t *testing.T) {

		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 9, 0, "", "").Return()
		err := GetStakePoolInfo("blobberID", mockGetInfoCallback)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestGetReadPoolInfo(t *testing.T) {
	t.Run("Test Get Read Pool Info checkConfig fails", func(t *testing.T) {

		_config.isConfigured = false
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 8, 0, "", "").Return()
		err := GetReadPoolInfo(clientID, mockGetInfoCallback)
		expectedErrorMsg := "SDK not initialized"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Read Pool Info empty clientID", func(t *testing.T) {

		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 8, 0, "", "").Return()
		err := GetReadPoolInfo("", mockGetInfoCallback)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Read Pool Info", func(t *testing.T) {

		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 8, 0, "", "").Return()
		err := GetReadPoolInfo(clientID, mockGetInfoCallback)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestGetAllocations(t *testing.T) {
	t.Run("Test Get Allocations empty clientID", func(t *testing.T) {

		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 7, 0, "", "").Return()
		err := GetAllocations("", mockGetInfoCallback)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Allocations checkConfig fails", func(t *testing.T) {

		_config.isConfigured = false
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 7, 0, "", "").Return()
		err := GetAllocations(clientID, mockGetInfoCallback)
		expectedErrorMsg := "SDK not initialized"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Allocations", func(t *testing.T) {

		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 7, 0, "", "").Return()
		err := GetAllocations(clientID, mockGetInfoCallback)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestGetAllocation(t *testing.T) {
	t.Run("Test Get Allocation checkConfig fails", func(t *testing.T) {

		_config.isConfigured = false
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 6, 0, "", "").Return()
		err := GetAllocation(clientID, mockGetInfoCallback)
		expectedErrorMsg := "SDK not initialized"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Allocation", func(t *testing.T) {

		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 6, 0, "", "").Return()
		err := GetAllocation(clientID, mockGetInfoCallback)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestGetChallengePoolInfo(t *testing.T) {
	t.Run("Test Get Challenge Pool Info", func(t *testing.T) {

		_config.isConfigured = false
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 5, 0, "", "").Return()
		err := GetChallengePoolInfo("allocID", mockGetInfoCallback)
		expectedErrorMsg := "SDK not initialized"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Challenge Pool Info", func(t *testing.T) {

		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 5, 0, "", "").Return()
		err := GetChallengePoolInfo("allocID", mockGetInfoCallback)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestGetStorageSCConfig(t *testing.T) {
	t.Run("Test Get Storage SC Config", func(t *testing.T) {

		_config.isConfigured = false
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 4, 0, "", "").Return()
		err := GetStorageSCConfig(mockGetInfoCallback)
		expectedErrorMsg := "SDK not initialized"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Storage SC Config", func(t *testing.T) {

		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 4, 0, "", "").Return()
		err := GetStorageSCConfig(mockGetInfoCallback)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestGetMinerSCConfig(t *testing.T) {
	t.Run("Test Get Miner SC Config checkConfig fail", func(t *testing.T) {

		_config.isConfigured = false
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 0, 0, "", "").Return()
		err := GetMinerSCConfig(mockGetInfoCallback)
		expectedErrorMsg := "SDK not initialized"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Miner SC Config", func(t *testing.T) {

		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 0, 0, "", "").Return()
		err := GetMinerSCConfig(mockGetInfoCallback)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestGetMinerSCUserInfo(t *testing.T) {
	t.Run("Test Get Miner SC Config checkConfig fail", func(t *testing.T) {

		_config.isConfigured = false
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 0, 0, "", "").Return()
		err := GetMinerSCUserInfo(clientID, mockGetInfoCallback)

		expectedErrorMsg := "SDK not initialized"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Miner SC Config empty clientID", func(t *testing.T) {

		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 0, 0, "", "").Return()
		err := GetMinerSCUserInfo("", mockGetInfoCallback)
		require.Nil(t, err)
		// expectedErrorMsg := "SDK not initialized"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Miner SC Config", func(t *testing.T) {

		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 0, 0, "", "").Return()
		err := GetMinerSCUserInfo(clientID, mockGetInfoCallback)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestGetMinerSCNodePool(t *testing.T) {
	t.Run("Test Get Miner SC Node Pool checkConfig fails", func(t *testing.T) {

		_config.isConfigured = false
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 0, 0, "", "").Return()
		err := GetMinerSCNodePool("id", "poolID", mockGetInfoCallback)
		expectedErrorMsg := "SDK not initialized"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Miner SC Node Pool", func(t *testing.T) {

		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 0, 0, "", "").Return()
		err := GetMinerSCNodePool("id", "poolID", mockGetInfoCallback)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestGetMinerSCNodeInfo(t *testing.T) {
	t.Run("Test Get Miner SC Node Info", func(t *testing.T) {
		_config.isConfigured = false
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 0, 0, "", "").Return()
		err := GetMinerSCNodeInfo("id", mockGetInfoCallback)
		expectedErrorMsg := "SDK not initialized"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Miner SC Node Info", func(t *testing.T) {
		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 0, 0, "", "").Return()
		err := GetMinerSCNodeInfo("id", mockGetInfoCallback)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestGetSharders(t *testing.T) {
	t.Run("Test Get Sharders checkConfig fails", func(t *testing.T) {
		_config.isConfigured = false
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 0, 0, "", "").Return()
		err := GetSharders(mockGetInfoCallback)
		expectedErrorMsg := "SDK not initialized"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Sharders", func(t *testing.T) {
		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 0, 0, "", "").Return()
		err := GetSharders(mockGetInfoCallback)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}

func TestGetMiners(t *testing.T) {
	t.Run("Test Get Miners checkConfig fails", func(t *testing.T) {
		_config.isConfigured = false
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 0, 0, "", "").Return()
		err := GetMiners(mockGetInfoCallback)
		expectedErrorMsg := "SDK not initialized"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Miners", func(t *testing.T) {
		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 0, 0, "", "").Return()
		err := GetMiners(mockGetInfoCallback)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestGetVestingSCConfig(t *testing.T) {
	t.Run("Test Get Vesting SC Config checkConfig fails", func(t *testing.T) {
		_config.isConfigured = false
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 0, 0, "", "").Return()
		err := GetVestingSCConfig(mockGetInfoCallback)
		expectedErrorMsg := "SDK not initialized"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Vesting SC Config", func(t *testing.T) {
		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 0, 0, "", "").Return()
		err := GetVestingSCConfig(mockGetInfoCallback)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestGetVestingClientList(t *testing.T) {
	t.Run("Test Get Vesting SC Config checkConfig fails", func(t *testing.T) {
		_config.isConfigured = false
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 0, 0, "", "").Return()
		err := GetVestingClientList(clientID, mockGetInfoCallback)
		expectedErrorMsg := "SDK not initialized"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Vesting SC Config empty clientID", func(t *testing.T) {
		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 0, 0, "", "").Return()
		err := GetVestingClientList("", mockGetInfoCallback)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Vesting SC Config", func(t *testing.T) {
		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 0, 0, "", "").Return()
		err := GetVestingClientList(clientID, mockGetInfoCallback)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}

func TestGetVestingPoolInfo(t *testing.T) {
	t.Run("Test Get Vesting Pool Info checkConfig fails", func(t *testing.T) {
		_config.isConfigured = false
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 0, 0, "", "").Return()
		err := GetVestingPoolInfo("poolID", mockGetInfoCallback)
		expectedErrorMsg := "SDK not initialized"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Vesting Pool Info", func(t *testing.T) {
		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 0, 0, "", "").Return()
		err := GetVestingPoolInfo("poolID", mockGetInfoCallback)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}

func TestGetIdForUrl(t *testing.T) {
	t.Run("Test Get Vesting Pool Info", func(t *testing.T) {
		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 0, 0, "", "").Return()
		resp := GetIdForUrl("url")
		require.NotNil(t, resp)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}

func TestSetupAuth(t *testing.T) {
	t.Run("Test Setup Auth", func(t *testing.T) {
		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockAuthCallback = mocks.AuthCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockAuthCallback.On("OnSetupComplete", 0, "").Return()
		resp := SetupAuth("authHost", "clientID", "clientKey", "publicKey", "privateKey", "localPublicKey", mockAuthCallback)
		require.Nil(t, resp)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestGetZcnUSDInfo(t *testing.T) {
	t.Run("Test Setup Auth", func(t *testing.T) {
		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetUSDInfoCallback = mocks.GetUSDInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetUSDInfoCallback.On("OnUSDInfoAvailable", 0, "", "").Return()
		resp := GetZcnUSDInfo(mockGetUSDInfoCallback)
		require.Nil(t, resp)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestGetWalletClientID(t *testing.T) {
	t.Run("Test Get Wallet Client ID", func(t *testing.T) {
		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetUSDInfoCallback = mocks.GetUSDInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetUSDInfoCallback.On("OnUSDInfoAvailable", 0, "", "").Return()
		resp, err := GetWalletClientID(walletString)
		require.Equal(t, "679b06b89fc418cfe7f8fc908137795de8b7777e9324901432acce4781031c93", resp)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestGetWallet(t *testing.T) {
	t.Run("Test Get Wallet Client ID", func(t *testing.T) {
		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetUSDInfoCallback = mocks.GetUSDInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetUSDInfoCallback.On("OnUSDInfoAvailable", 0, "", "").Return()
		resp, err := GetWallet(walletString)
		require.NotEmpty(t, resp)
		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Wallet Client ID error while parsing wallet string", func(t *testing.T) {
		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetUSDInfoCallback = mocks.GetUSDInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetUSDInfoCallback.On("OnUSDInfoAvailable", 0, "", "").Return()
		resp, err := GetWallet("walletString")
		require.Nil(t, resp)
		expectedErrorMsg := "invalid character 'w' looking for beginning of value"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestGetLockedTokens(t *testing.T) {
	t.Run("Test Get Wallet Client ID checkSdkInit fails", func(t *testing.T) {
		_config.isConfigured = false
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 1, 0, "", "").Return()
		err := GetLockedTokens(mockGetInfoCallback)

		expectedErrorMsg := "SDK not initialized"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Wallet Client ID", func(t *testing.T) {
		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 1, 0, "", "").Return()
		err := GetLockedTokens(mockGetInfoCallback)

		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestGetLockConfig(t *testing.T) {
	t.Run("Test Get Lock Config checkSdkInit fail", func(t *testing.T) {
		_config.isConfigured = false
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 0, 0, "", "").Return()
		err := GetLockConfig(mockGetInfoCallback)

		expectedErrorMsg := "SDK not initialized"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Get Lock Config", func(t *testing.T) {
		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 0, 0, "", "").Return()
		err := GetLockConfig(mockGetInfoCallback)

		require.NoError(t, err)
		// expectedErrorMsg := "crypto/aes: invalid key size 64"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestConvertUSDToToken(t *testing.T) {
	t.Run("Test Convert USD To Token", func(t *testing.T) {
		// var mockClient = mocks.HttpClient{}

		// util.Client = &mockClient
		_config.isConfigured = true
		_config.chain.Miners = []string{"", ""}
		_config.chain.Sharders = []string{"", ""}
		var mockGetInfoCallback = mocks.GetInfoCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetInfoCallback.On("OnInfoAvailable", 0, 0, "", "").Return()

		resp, err := ConvertUSDToToken(2)

		require.Equal(t, float64(0), resp)
		expectedErrorMsg := "unexpected end of JSON input"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestConvertToValue(t *testing.T) {
	t.Run("Test Convert USD To Token", func(t *testing.T) {
		resp := ConvertToValue(1)

		require.Equal(t, int64(10000000000), resp)
		// expectedErrorMsg := "unexpected end of JSON input"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestConvertToToken(t *testing.T) {
	t.Run("Test Convert USD To Token", func(t *testing.T) {
		resp := ConvertToToken(10000000)

		require.Equal(t, float64(0.001), resp)
		// expectedErrorMsg := "unexpected end of JSON input"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestGetBalance(t *testing.T) {
	t.Run("Test Convert USD To Token checkConfig fail", func(t *testing.T) {
		var mockGetBalanceCallback = mocks.GetBalanceCallback{}
		_config.isConfigured = false
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetBalanceCallback.On("OnBalanceAvailable", 2, int64(0), "").Return()
		err := GetBalance(mockGetBalanceCallback)

		expectedErrorMsg := "SDK not initialized"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Convert USD To Token wallet info not found", func(t *testing.T) {
		var mockGetBalanceCallback = mocks.GetBalanceCallback{}
		_config.isConfigured = true

		_config.isValidWallet = false
		_config.wallet.ClientID = clientID

		mockGetBalanceCallback.On("OnBalanceAvailable", 2, int64(0), "").Return()
		err := GetBalance(mockGetBalanceCallback)

		expectedErrorMsg := "wallet info not found. set wallet info."
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	//_config.isConfigured
	t.Run("Test Convert USD To Token", func(t *testing.T) {
		var mockGetBalanceCallback = mocks.GetBalanceCallback{}
		_config.isConfigured = true

		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetBalanceCallback.On("OnBalanceAvailable", 2, int64(0), "").Return()
		err := GetBalance(mockGetBalanceCallback)

		require.NoError(t, err)
	})
}

func TestSetAuthUrl(t *testing.T) {
	t.Run("Test Set Auth Url wallet type is not split key", func(t *testing.T) {
		var mockGetBalanceCallback = mocks.GetBalanceCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetBalanceCallback.On("OnBalanceAvailable", 2, int64(0), "").Return()
		err := SetAuthUrl("url")

		expectedErrorMsg := "wallet type is not split key"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Set Auth Url invalid auth url", func(t *testing.T) {
		var mockGetBalanceCallback = mocks.GetBalanceCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID
		_config.isSplitWallet = true

		mockGetBalanceCallback.On("OnBalanceAvailable", 2, int64(0), "").Return()
		err := SetAuthUrl("")

		expectedErrorMsg := "invalid auth url"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Test Set Auth Url", func(t *testing.T) {
		var mockGetBalanceCallback = mocks.GetBalanceCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID
		_config.isSplitWallet = true
		mockGetBalanceCallback.On("OnBalanceAvailable", 2, int64(0), "").Return()
		err := SetAuthUrl("url")

		require.NoError(t, err)
		// expectedErrorMsg := "invalid auth url"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}
func TestSetWalletInfo(t *testing.T) {
	t.Run("Test Convert USD To Token", func(t *testing.T) {
		var mockGetBalanceCallback = mocks.GetBalanceCallback{}
		_config.isValidWallet = true
		_config.wallet.ClientID = clientID

		mockGetBalanceCallback.On("OnBalanceAvailable", 2, int64(0), "").Return()

		jsonFR, err := json.Marshal(&zcncrypto.Wallet{
			ClientID:    "",
			ClientKey:   "",
			Keys:        []zcncrypto.KeyPair{},
			Mnemonic:    "",
			Version:     "",
			DateCreated: "",
		})
		err = SetWalletInfo(string(jsonFR), true)

		require.NoError(t, err)
	})
}

func TestIsMnemonicValid(t *testing.T) {
	t.Run("Test Is Mnemonic Valid", func(t *testing.T) {
		
		resp := IsMnemonicValid(mnemonic)
		require.True(t, resp)
		
		// expectedErrorMsg := "invalid auth url"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}

func TestConvertTokenToUSD(t *testing.T) {
	type CoinGeckoResponse struct {
		ID         string `json:"id"`
		Symbol     string `json:"symbol"`
		MarketData struct {
			CurrentPrice map[string]float64 `json:"current_price"`
		} `json:"market_data"`
	}
	t.Run("Test Convert Token To USD", func(t *testing.T) {

		resp, err := ConvertTokenToUSD(1)
		require.NotNil(t, resp)
		require.NoError(t, err)
		// expectedErrorMsg := "unexpected end of JSON input"
		// assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
}

func TestGetClientDetails(t *testing.T) {
	t.Run("Test Get Client Details", func(t *testing.T) {
		var mockClient = mocks.HttpClient{}
		_config.chain.Miners = []string{"TestGetClientDetails"}
		util.Client = &mockClient
		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			if req.Method == "GET" && (strings.HasPrefix(req.URL.Path, "TestGetClientDetails") || strings.HasPrefix(req.URL.Path, "/v1/client/get") || strings.HasPrefix(req.URL.Path, "/v1/screst/")) {

				return true
			}
			return false
		})).Return(&http.Response{
			Body: func() io.ReadCloser {
				jsonFR, err := json.Marshal(&GetClientResponse{
					ID:           "",
					Version:      "",
					CreationDate: 1,
					PublicKey:    "",
				})
				require.NoError(t, err)
				return ioutil.NopCloser(bytes.NewReader([]byte(jsonFR)))
			}(),
			StatusCode: http.StatusOK,
		}, nil)
		resp, err := GetClientDetails(clientID)
		require.NotEmpty(t, resp)
		require.Nil(t, err)
	})
}
