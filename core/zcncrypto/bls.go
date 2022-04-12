package zcncrypto

import "io"

var BlsSignerInstance BlsSigner

type BlsSigner interface {
	SetRandFunc(randReader io.Reader)
	FrSub(out Fr, x Fr, y Fr)

	NewFr() Fr
	NewSecretKey() SecretKey
	NewPublicKey() PublicKey
	NewSignature() Signature
	NewID() ID
}

// Fr --
type Fr interface {
	Serialize() []byte

	SetLittleEndian(buf []byte) error
}

type SecretKey interface {
	SerializeToHexStr() string
	DeserializeHexStr(s string) error

	Serialize() []byte

	GetLittleEndian() []byte
	SetLittleEndian(buf []byte) error

	SetByCSPRNG()

	GetPublicKey() PublicKey

	Sign(m string) Signature
	Add(rhs SecretKey)

	GetMasterSecretKey(k int) (msk []SecretKey, err error)
	Set(msk []SecretKey, id ID) error
}

type PublicKey interface {
	SerializeToHexStr() string
	DeserializeHexStr(s string) error

	Serialize() []byte
}

type Signature interface {
	SerializeToHexStr() string
	DeserializeHexStr(s string) error

	Add(rhs Signature)

	Verify(pk PublicKey, m string) bool
}

type ID interface {
	SetHexString(s string) error
	GetHexString() string

	SetDecString(s string) error
}
