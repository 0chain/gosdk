package sdks

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/zcncrypto"
)

// ZBox  sdk client instance
type ZBox struct {
	// ClientID client id
	ClientID string
	// ClientKey client key
	ClientKey string
	// SignatureScheme signature scheme
	SignatureScheme string

	// Wallet wallet
	Wallet zcncrypto.Wallet

	// NewRequest create http request
	NewRequest func(method, url string, body io.Reader) (*http.Request, error)
}

// New create a sdk client instance
func New(clientID, clientKey, signatureScheme string) *ZBox {
	s := &ZBox{
		ClientID:        clientID,
		ClientKey:       clientKey,
		SignatureScheme: signatureScheme,
		NewRequest:      http.NewRequest,
	}

	return s
}

// InitWallet init wallet from json
func (z *ZBox) InitWallet(js string) error {
	return json.Unmarshal([]byte(js), &z.Wallet)
}

// SignRequest sign request with client_id, client_key and sign
func (z *ZBox) SignRequest(req *http.Request, allocationID string) error {

	if req == nil {
		return errors.Throw(constants.ErrInvalidParameter, "req")
	}

	req.Header.Set("X-App-Client-ID", z.ClientID)
	req.Header.Set("X-App-Client-Key", z.ClientKey)

	hash := encryption.Hash(allocationID)

	var err error
	sign := ""
	for _, kv := range z.Wallet.Keys {
		ss := zcncrypto.NewSignatureScheme(z.SignatureScheme)
		err = ss.SetPrivateKey(kv.PrivateKey)
		if err != nil {
			return err
		}

		if len(sign) == 0 {
			sign, err = ss.Sign(hash)
		} else {
			sign, err = ss.Add(sign, hash)
		}
		if err != nil {
			return err
		}
	}

	req.Header.Set("X-App-Client-Signature", sign)

	return nil
}
