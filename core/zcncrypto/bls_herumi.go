//go:build !js && !wasm
// +build !js,!wasm

package zcncrypto

import (
	"errors"
	"io"

	"github.com/herumi/bls-go-binary/bls"
)

func init() {
	err := bls.Init(bls.CurveFp254BNb)
	if err != nil {
		panic(err)
	}
	BlsSignerInstance = &herumiBls{}
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

func (b *herumiBls) NewID() ID {
	id := &herumiID{}

	return id
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

func (sk *herumiSecretKey) GetMasterSecretKey(k int) ([]SecretKey, error) {
	if k < 1 {
		return nil, errors.New("cannot get master secret key for threshold less than 1")
	}

	list := sk.SecretKey.GetMasterSecretKey(k)

	msk := make([]SecretKey, len(list))

	for i, it := range list {
		msk[i] = &herumiSecretKey{SecretKey: it}

	}

	return msk, nil
}

func (sk *herumiSecretKey) Set(msk []SecretKey, id ID) error {

	blsMsk := make([]bls.SecretKey, len(msk))

	for i, it := range msk {
		k, ok := it.(*herumiSecretKey)
		if !ok {
			return errors.New("invalid herumi secret key")
		}

		blsMsk[i] = k.SecretKey
	}

	blsID, _ := id.(*herumiID)

	return sk.SecretKey.Set(blsMsk, &blsID.ID)
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

type herumiID struct {
	bls.ID
}

func (id *herumiID) SetHexString(s string) error {
	return id.ID.SetHexString(s)
}
func (id *herumiID) GetHexString() string {
	return id.ID.GetHexString()
}

func (id *herumiID) SetDecString(s string) error {
	return id.ID.SetDecString(s)
}
