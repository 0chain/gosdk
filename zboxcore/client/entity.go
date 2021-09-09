package client

import (
	"encoding/json"

	"github.com/0chain/gosdk/core/zcncrypto"
)

type Client struct {
	*zcncrypto.Wallet
	signatureSchemeString string `json:"signature_scheme"`
}

var client *Client

func init() {
	client = &Client{
		Wallet: &zcncrypto.Wallet{},
	}
}

func PopulateClient(clientjson string, signatureScheme string) error {
	err := json.Unmarshal([]byte(clientjson), &client)
	client.signatureSchemeString = signatureScheme
	return err
}

func GetClient() *Client {
	return client
}

func GetClientID() string {
	return client.Wallet.ClientID
}

func GetClientPublicKey() string {
	return client.Wallet.ClientKey
}

func Sign(hash string) (string, error) {
	retSignature := ""
	for _, kv := range client.Wallet.Keys {
		ss := zcncrypto.NewSignatureScheme(client.signatureSchemeString)
		ss.SetPrivateKey(kv.PrivateKey)
		var err error
		if len(retSignature) == 0 {
			retSignature, err = ss.Sign(hash)
		} else {
			retSignature, err = ss.Add(retSignature, hash)
		}
		if err != nil {
			return "", err
		}
	}
	return retSignature, nil
}

func VerifySignature(signature string, msg string) (bool, error) {
	ss := zcncrypto.NewSignatureScheme(client.signatureSchemeString)
	ss.SetPublicKey(client.Wallet.ClientKey)
	return ss.Verify(signature, msg)
}
