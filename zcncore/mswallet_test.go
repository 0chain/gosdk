package zcncore

import (
	"encoding/json"
	"testing"

	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	walletString = `{"client_id":"679b06b89fc418cfe7f8fc908137795de8b7777e9324901432acce4781031c93","client_key":"2c2aaca87c9d80108c4d5dc27fc8eefc57be98af55d26a548ebf92a86cd90615d19d715a9ed6d009798877189babf405384a2980e102ce72f824890b20f8ce1e","keys":[{"public_key":"2c2aaca87c9d80108c4d5dc27fc8eefc57be98af55d26a548ebf92a86cd90615d19d715a9ed6d009798877189babf405384a2980e102ce72f824890b20f8ce1e","private_key":""}],"mnemonics":"bamboo list citizen release bronze duck woman moment cart crucial extra hip witness mixture flash into priority length pattern deposit title exhaust flush addict","version":"1.0","date_created":"2021-06-15 11:11:40.306922176 +0700 +07 m=+1.187131283"}`
	mnemonic     = `snake mixed bird cream cotton trouble small fee finger catalog measure spoon private second canal pact unable close predict dream mask delay path inflict`
	clientID     = `0bc96a0980170045863d826f9eb579d8144013210602e88426408e9f83c236f6`
	public_key   = `2c2aaca87c9d80108c4d5dc27fc8eefc57be98af55d26a548ebf92a86cd90615d19d715a9ed6d009798877189babf405384a2980e102ce72f824890b20f8ce1e`
	private_key  = `private_key`
	msw          = `{
		"id":1,
		"signature_scheme":"test",
		"group_client_id":"test",
		"group_key":{
		   "public_key":"2c2aaca87c9d80108c4d5dc27fc8eefc57be98af55d26a548ebf92a86cd90615d19d715a9ed6d009798877189babf405384a2980e102ce72f824890b20f8ce1e",
		   "private_key":"",
		   "mnemonic":""
		},
		"signer_client_ids":[
		   "test0",
		   "test1"
		],
		"signer_keys":[
		   
		],
		"t":1,
		"n":2
	 }`
	mswFail = ""
	msv     = `{
		"proposal_id": "",
		"transfer": {
			"from": "",
			"to": "",
			"amount": 100
		},
		"signature": ""
	}`
	msvFail = ""
	hash    = "127e6fbfe24a750e72930c220a8e138275656b8e5d8f48a98c3c92df2caba935"
)

func TestCreateMSWallet(t *testing.T) {
	t.Run("Create MSWallet encryption scheme fails", func(t *testing.T) {
		smsw, groupClientID, wallets, err := CreateMSWallet(1, 1)
		expectedErrorMsg := "encryption scheme for this blockchain is not bls0chain"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
		assert.Equal(t, "", smsw)
		assert.Equal(t, "", groupClientID)
		assert.Equal(t, []string([]string(nil)), wallets)
	})
	t.Run("Success create MSWallet", func(t *testing.T) {
		_config.chain.SignatureScheme = "bls0chain"
		smsw, groupClientID, wallets, err := CreateMSWallet(1, 1)
		require.NoError(t, err)
		require.NotNil(t, smsw)
		require.NotNil(t, groupClientID)
		require.NotNil(t, wallets)
	})
}

func TestRegisterWallet(t *testing.T) {
	t.Run("Success Register MSWallet", func(t *testing.T) {
		_config.chain.SignatureScheme = "bls0chain"

		var mockWalletCallback = mocks.WalletCallback{}
		mockWalletCallback.On("OnWalletCreateComplete", 0, walletString, "").Return()
		RegisterWallet(walletString, mockWalletCallback)
	})
}

func TestCreateMSVote(t *testing.T) {
	t.Run("Field empty Create  MSVote", func(t *testing.T) {
		_config.chain.SignatureScheme = "bls0chain"
		resp, err := CreateMSVote("", "", "", "", 123)

		expectedErrorMsg := "proposal or groupClient or signer wallet or toClientID cannot be empty"

		assert.Equal(t, "", resp)
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Token less than 1", func(t *testing.T) {
		_config.chain.SignatureScheme = "bls0chain"
		resp, err := CreateMSVote("proposal", "grpClientID", mnemonic, clientID, 0)

		expectedErrorMsg := "Token cannot be less than 1"

		assert.Equal(t, "", resp)
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Error while parsing the signer wallet", func(t *testing.T) {
		_config.chain.SignatureScheme = "bls0chain"
		resp, err := CreateMSVote("proposal", "grpClientID", mnemonic, clientID, 0)

		expectedErrorMsg := "Token cannot be less than 1"

		assert.Equal(t, "", resp)
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
	})
	t.Run("Error while parsing the signer wallet", func(t *testing.T) {
		_config.chain.SignatureScheme = "bls0chain"
		wallet := &zcncrypto.Wallet{
			Keys: []zcncrypto.KeyPair{
				{
					PublicKey:  mockPublicKey,
					PrivateKey: mockPrivateKey,
				},
			},
		}
		str, err := json.Marshal(wallet)
		require.NoError(t, err)

		resp, err := CreateMSVote("proposal", "grpClientID", string(str), clientID, 2)

		assert.NotEmpty(t, resp)
		require.Nil(t, err)
	})
}

func TestGetWallets(t *testing.T) {
	t.Run("Get Wallets Success", func(t *testing.T) {
		msw := MSWallet{
			Id: 123,
			GroupKey: &zcncrypto.BLS0ChainScheme{
				PublicKey:  public_key,
				PrivateKey: private_key,
				Mnemonic:   mnemonic,
			},
		}
		resp, err := getWallets(msw)
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})
}
func TestMakeWallet(t *testing.T) {
	t.Run("Make Wallets Success", func(t *testing.T) {

		resp, err := makeWallet(private_key, public_key, mnemonic)
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})
}
func TestGetClientID(t *testing.T) {
	t.Run("Get Client ID Success", func(t *testing.T) {

		resp := GetClientID(public_key)
		// require.NoError(t, err)
		assert.NotNil(t, resp)
	})
}
func TestGetMultisigPayload(t *testing.T) {
	t.Run("Get Client ID Success", func(t *testing.T) {

		resp, err := GetMultisigPayload(msw)
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})
	t.Run("Get Client ID Success", func(t *testing.T) {

		resp, err := GetMultisigPayload(mswFail)
		expectedErrorMsg := "unexpected end of JSON input"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
		assert.Equal(t, "", resp)
	})
}
func TestGetMultisigVotePayload(t *testing.T) {
	t.Run("Get Multisig Vote Payload Success", func(t *testing.T) {

		resp, err := GetMultisigVotePayload(msv)
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})
	t.Run("Get Multisig Vote Payload Fail", func(t *testing.T) {

		resp, err := GetMultisigVotePayload(msvFail)
		expectedErrorMsg := "unexpected end of JSON input"
		assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
		assert.Equal(t, nil, resp)
	})
}
