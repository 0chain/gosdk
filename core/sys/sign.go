package sys

// KeyPair private and publickey
type KeyPair struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

// SignFunc sign method for request verification
type SignFunc func(hash string, signatureScheme string, keys []KeyPair) (string, error)

type VerifyFunc func(signature string, msg string) (bool, error)

type VerifyFuncWith func(pk, signature string, msg string) (bool, error)
