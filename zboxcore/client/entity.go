package client

import (
	"encoding/json"

	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/core/zcncrypto"
)

type SignFunc func(hash string) (string, error)

type Client struct {
	*zcncrypto.Wallet
	SignatureScheme string
	providedFee     uint64
}

var (
	client  *Client
	clients []*Client
	Sign    SignFunc
)

func init() {
	client = &Client{
		Wallet: &zcncrypto.Wallet{},
	}

	sys.Sign = SignHash
	// initialize SignFunc as default implementation
	Sign = func(hash string) (string, error) {
		return sys.Sign(hash, client.SignatureScheme, GetClientSysKeys())
	}

	sys.Verify = VerifySignature
}

// Populate Single Client
func PopulateClient(clientjson string, signatureScheme string) error {
	err := json.Unmarshal([]byte(clientjson), &client)
	client.SignatureScheme = signatureScheme
	return err
}

func SetClientNonce(nonce int64) {
	client.Nonce = nonce
}

func SetTxFee(fee uint64) {
	client.providedFee = fee
}

func TxFee() uint64 {
	return client.providedFee
}

// PopulateClients This is a workaround for blobber tests that requires multiple clients to test authticket functionality
func PopulateClients(clientJsons []string, signatureScheme string) error {
	for _, clientJson := range clientJsons {
		c := new(Client)
		if err := json.Unmarshal([]byte(clientJson), c); err != nil {
			return err
		}
		c.SignatureScheme = signatureScheme
		clients = append(clients, c)
	}
	return nil
}

func GetClient() *Client {
	return client
}

func GetClients() []*Client {
	return clients
}

func GetClientID() string {
	return client.ClientID
}

func GetClientPublicKey() string {
	return client.ClientKey
}

// GetClientSysKeys convert client.KeyPair to sys.KeyPair
func GetClientSysKeys() []sys.KeyPair {
	var keys []sys.KeyPair
	if client != nil {
		for _, kv := range client.Keys {
			keys = append(keys, sys.KeyPair{
				PrivateKey: kv.PrivateKey,
				PublicKey:  kv.PublicKey,
			})
		}
	}

	return keys
}

func SignHash(hash string, signatureScheme string, keys []sys.KeyPair) (string, error) {
	retSignature := ""
	for _, kv := range keys {
		ss := zcncrypto.NewSignatureScheme(signatureScheme)
		err := ss.SetPrivateKey(kv.PrivateKey)
		if err != nil {
			return "", err
		}

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
	ss := zcncrypto.NewSignatureScheme(client.SignatureScheme)
	if err := ss.SetPublicKey(client.Keys[0].PublicKey); err != nil {
		return false, err
	}

	return ss.Verify(signature, msg)
}
