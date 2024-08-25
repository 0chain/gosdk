// Methods and types for client and wallet operations.
package client

import (
	"encoding/json"
	"fmt"

	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/core/zcncrypto"
)

type SignFunc func(hash string) (string, error)

// Client represents client information
type Client struct {
	*zcncrypto.Wallet
	SignatureScheme string
	txnFee          uint64
}

var (
	client  *Client
	clients []*Client

	// Sign is a function to sign a hash
	Sign SignFunc
	sigC = make(chan struct{}, 1)
)

func init() {
	client = &Client{
		Wallet: &zcncrypto.Wallet{},
	}

	sigC <- struct{}{}

	sys.Sign = signHash
	sys.SignWithAuth = signHash

	// initialize SignFunc as default implementation
	Sign = func(hash string) (string, error) {
		if client.PeerPublicKey == "" {
			return sys.Sign(hash, client.SignatureScheme, GetClientSysKeys())
		}

		// get sign lock
		<-sigC
		fmt.Println("Sign: with sys.SignWithAuth:", sys.SignWithAuth, "sysKeys:", GetClientSysKeys())
		sig, err := sys.SignWithAuth(hash, client.SignatureScheme, GetClientSysKeys())
		sigC <- struct{}{}
		return sig, err
	}

	sys.Verify = VerifySignature
	sys.VerifyWith = VerifySignatureWith
}

func SetClient(w *zcncrypto.Wallet, signatureScheme string, txnFee uint64) {
	client.Wallet = w
	client.SignatureScheme = signatureScheme
	client.txnFee = txnFee
}

// PopulateClient populates single client
//   - clientjson: client json string
//   - signatureScheme: signature scheme
func PopulateClient(clientjson string, signatureScheme string) error {
	err := json.Unmarshal([]byte(clientjson), &client)
	client.SignatureScheme = signatureScheme
	return err
}

// SetClientNonce sets client nonce
func SetClientNonce(nonce int64) {
	client.Nonce = nonce
}

// SetTxnFee sets general transaction fee
func SetTxnFee(fee uint64) {
	client.txnFee = fee
}

// TxnFee gets general txn fee
func TxnFee() uint64 {
	return client.txnFee
}

// PopulateClients This is a workaround for blobber tests that requires multiple clients to test authticket functionality
//   - clientJsons: array of client json strings
//   - signatureScheme: signature scheme
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

// GetClient returns client instance
func GetClient() *Client {
	return client
}

// GetClients returns all clients
func GetClients() []*Client {
	return clients
}

// GetClientID returns client id
func GetClientID() string {
	return client.ClientID
}

// GetClientPublicKey returns client public key
func GetClientPublicKey() string {
	return client.ClientKey
}

func GetClientPeerPublicKey() string {
	return client.PeerPublicKey

}

// GetClientPrivateKey returns client private key
func GetClientPrivateKey() string {
	for _, kv := range client.Keys {
		return kv.PrivateKey
	}

	return ""
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

func signHash(hash string, signatureScheme string, keys []sys.KeyPair) (string, error) {
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

// VerifySignature verifies signature of a message with client public key and signature scheme
//   - signature: signature to use for verification
//   - msg: message to verify
func VerifySignature(signature string, msg string) (bool, error) {
	ss := zcncrypto.NewSignatureScheme(client.SignatureScheme)
	if err := ss.SetPublicKey(client.ClientKey); err != nil {
		return false, err
	}

	return ss.Verify(signature, msg)
}

// VerifySignatureWith verifies signature of a message with a given public key, and the client's signature scheme
//   - pubKey: public key to use for verification
//   - signature: signature to use for verification
//   - hash: message to verify
func VerifySignatureWith(pubKey, signature, hash string) (bool, error) {
	sch := zcncrypto.NewSignatureScheme(client.SignatureScheme)
	err := sch.SetPublicKey(pubKey)
	if err != nil {
		return false, err
	}
	return sch.Verify(signature, hash)
}
