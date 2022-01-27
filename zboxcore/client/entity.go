package client

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/0chain/gosdk/core/zcncrypto"
)

type SignFunc func(hash string) (string, error)

type Client struct {
	*zcncrypto.Wallet
	signatureSchemeString string
}

var (
	client  *Client
	clients *[]Client
	Sign    SignFunc
)

func init() {
	client = &Client{
		Wallet: &zcncrypto.Wallet{},
	}

	clients = &[]Client{
		{
			Wallet: &zcncrypto.Wallet{},
		},
	}

	Sign = defaultSignFunc
}

// Populate Single Client
func PopulateClient(clientjson string, signatureScheme string) error {
	err := json.Unmarshal([]byte(clientjson), &client)
	client.signatureSchemeString = signatureScheme
	return err
}

// Populate multiple Client through a slice of JSON strings
func PopulateClients(clientjsons []string, signatureScheme string) error {
	allClients := strings.Join(clientjsons, "")
	err := json.Unmarshal([]byte(allClients), &clients)

	for _, c := range *clients {
		c.signatureSchemeString = signatureScheme
	}

	return err
}

func GetClient() *Client {
	return client
}

func GetClients() *[]Client {
	return clients
}

func GetClientID() string {
	return client.ClientID
}

func GetClientIDByIndex(index int) (string, error) {
	for i, c := range *clients {
		if i == index {
			return c.ClientID, nil
		}
	}
	return "", errors.New("input index is out of bounds")
}

func GetClientPublicKey() string {
	return client.ClientKey
}

func GetClientPublicKeyByIndex(index int) (string, error) {
	for i, c := range *clients {
		if i == index {
			return c.ClientKey, nil
		}
	}
	return "", errors.New("input index is out of bounds")
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
