package authorizer_test

import (
	"encoding/json"
	"testing"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zcnbridge/authorizer"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const walletString = `{
  "client_id": "39ac4c04401762778de45d00c51934b2ff15d8168722585daa422195a1ebc5d5",
  "client_key": "9cf22d230343a68a149a2ccea1137846dd6d0fe653985eb8bebfbe2583ab0e1e2281b78218f480a50cac78cd0c716c3aafa8d741af813a94ca0a5b4e0b8e7a22",
  "keys": [
    {
      "public_key": "9cf22d230343a68a149a2ccea1137846dd6d0fe653985eb8bebfbe2583ab0e1e2281b78218f480a50cac78cd0c716c3aafa8d741af813a94ca0a5b4e0b8e7a22",
      "private_key": "9c6c9022bdbd05670c07eae339dbb5491b1c035c2287a83c56581c505b824623"
    }
  ],
  "mnemonics": "fortune guitar marine bachelor ocean raven hunt silver pass hurt industry forget cradle shuffle render used used order chat shallow aerobic cry exercise junior",
  "version": "1.0",
  "date_created": "2022-07-17T16:12:25+05:00",
  "nonce": 0
}
`

type TicketTestSuite struct {
	suite.Suite
	w *zcncrypto.Wallet
}

func TestTicketTestSuite(t *testing.T) {
	suite.Run(t, new(TicketTestSuite))
}

func (suite *TicketTestSuite) SetupTest() {
	w := &zcncrypto.Wallet{}
	err := json.Unmarshal([]byte(walletString), w)
	require.NoError(suite.T(), err)
	suite.w = w
}

func (suite *TicketTestSuite) TestBasicSignature() {
	hash := zcncrypto.Sha3Sum256("test")
	signScheme := zcncrypto.NewSignatureScheme("bls0chain")
	err := signScheme.SetPrivateKey(suite.w.Keys[0].PrivateKey)
	require.NoError(suite.T(), err)
	sign, err := signScheme.Sign(hash)
	require.NoError(suite.T(), err)
	require.NotEmpty(suite.T(), sign)
}

func (suite *TicketTestSuite) TestTicketSignature() {
	pb := &authorizer.ProofOfBurn{
		TxnID:           "TxnID",
		Nonce:           100,
		Amount:          10,
		EthereumAddress: "0xBEEF",
		Signature:       nil,
	}

	conf.InitClientConfig(&conf.Config{
		SignatureScheme: constants.BLS0CHAIN.String(),
	})
	err := pb.SignWith0Chain(suite.w)
	require.NoError(suite.T(), err)
	require.NotEmpty(suite.T(), pb.Signature)
}
