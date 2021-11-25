//go:build !js && !wasm
// +build !js,!wasm

package zcncrypto

import (
	"io"

	"github.com/herumi/bls-go-binary/bls"
)

func init() {
	err := bls.Init(bls.CurveFp254BNb)
	if err != nil {
		panic(err)
	}
	blsInstance = &herumiBls{}
}

type herumiBls struct {
}

func (b *herumiBls) NewFr() Fr {
	return &herumiFr{}
}
func (b *herumiBls) NewSecretKey() SecretKey {
	return &herumiSecretKey{}
}

func (b *herumiBls) NewPublicKey() PublicKey {
	return &herumiPublicKey{
		PublicKey: &bls.PublicKey{},
	}
}

func (b *herumiBls) NewSignature() Signature {
	sg := &herumiSignature{
		Sign: &bls.Sign{},
	}

	return sg
}

func (b *herumiBls) SetRandFunc(randReader io.Reader) {
	bls.SetRandFunc(randReader)
}

func (b *herumiBls) FrSub(out Fr, x Fr, y Fr) {
	o1, _ := out.(*herumiFr)
	x1, _ := x.(*herumiFr)
	y1, _ := y.(*herumiFr)

	bls.FrSub(&o1.Fr, &x1.Fr, &y1.Fr)
}

type herumiFr struct {
	bls.Fr
}

func (fr *herumiFr) Serialize() []byte {
	return fr.Fr.Serialize()
}

func (fr *herumiFr) SetLittleEndian(buf []byte) error {
	return fr.Fr.SetLittleEndian(buf)
}

type herumiSecretKey struct {
	bls.SecretKey
}

func (sk *herumiSecretKey) SerializeToHexStr() string {
	return sk.SecretKey.SerializeToHexStr()
}
func (sk *herumiSecretKey) DeserializeHexStr(s string) error {
	return sk.SecretKey.DeserializeHexStr(s)
}

func (sk *herumiSecretKey) Serialize() []byte {
	return sk.SecretKey.Serialize()
}

func (sk *herumiSecretKey) GetLittleEndian() []byte {
	return sk.SecretKey.GetLittleEndian()
}
func (sk *herumiSecretKey) SetLittleEndian(buf []byte) error {
	return sk.SecretKey.SetLittleEndian(buf)
}

func (sk *herumiSecretKey) SetByCSPRNG() {
	sk.SecretKey.SetByCSPRNG()
}

func (sk *herumiSecretKey) GetPublicKey() PublicKey {
	pk := sk.SecretKey.GetPublicKey()
	return &herumiPublicKey{
		PublicKey: pk,
	}
}

func (sk *herumiSecretKey) Add(rhs SecretKey) {
	i, _ := rhs.(*herumiSecretKey)
	sk.SecretKey.Add(&i.SecretKey)
}

func (sk *herumiSecretKey) Sign(m string) Signature {
	sig := sk.SecretKey.Sign(m)

	return &herumiSignature{
		Sign: sig,
	}
}

type herumiPublicKey struct {
	*bls.PublicKey
}

func (pk *herumiPublicKey) SerializeToHexStr() string {
	return pk.PublicKey.SerializeToHexStr()
}

func (pk *herumiPublicKey) DeserializeHexStr(s string) error {
	return pk.PublicKey.DeserializeHexStr(s)
}

func (pk *herumiPublicKey) Serialize() []byte {
	return pk.PublicKey.Serialize()
}

type herumiSignature struct {
	*bls.Sign
}

// SerializeToHexStr --
func (sg *herumiSignature) SerializeToHexStr() string {
	return sg.Sign.SerializeToHexStr()
}

func (sg *herumiSignature) DeserializeHexStr(s string) error {
	return sg.Sign.DeserializeHexStr(s)
}

func (sg *herumiSignature) Add(rhs Signature) {
	sg2, _ := rhs.(*herumiSignature)

	sg.Sign.Add(sg2.Sign)
}

func (sg *herumiSignature) Verify(pk PublicKey, m string) bool {
	pub, _ := pk.(*herumiPublicKey)

	return sg.Sign.Verify(pub.PublicKey, m)
}
