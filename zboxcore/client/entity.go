package client

import (
	"encoding/json"

	"github.com/0chain/gosdk/core/zcncrypto"
)

type SignFunc func(hash string) (string, error)

type Client struct {
	*zcncrypto.Wallet
	signatureSchemeString string
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

	Sign = defaultSignFunc
}

// Populate Single Client
func PopulateClient(clientjson string, signatureScheme string) error {
	err := json.Unmarshal([]byte(clientjson), &client)
	client.signatureSchemeString = signatureScheme
	return err
}

// PopulateClients This is a workaround for blobber tests that requires multiple clients to test authticket functionality
func PopulateClients(clientJsons []string, signatureScheme string) error {
	for _, clientJson := range clientJsons {
		c := new(Client)
		if err := json.Unmarshal([]byte(clientJson), c); err != nil {
			return err
		}
		c.signatureSchemeString = signatureScheme
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

func defaultSignFunc(hash string) (string, error) {
	retSignature := ""
	for _, kv := range client.Keys {
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
	ss.SetPublicKey(client.Keys[0].PublicKey)
	return ss.Verify(signature, msg)
}
