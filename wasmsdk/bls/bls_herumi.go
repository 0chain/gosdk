//go:build js && wasm
// +build js,wasm

package bls

import (
	"errors"
	"io"
	"syscall/js"

	"github.com/0chain/gosdk/core/zcncrypto"
)

func init() {

	jsBls := NewJsObject(js.Global().Get("bls"))

	//	bls.BN25 == 0
	jsBls.Init(0)

	zcncrypto.BlsSignerInstance = &herumiBls{
		JsObject: jsBls,
	}
}

type herumiBls struct {
	JsObject
}

func (b *herumiBls) NewFr() zcncrypto.Fr {
	return &herumiFr{
		Fr: b.JsObject.NewFr(),
	}
}
func (b *herumiBls) NewSecretKey() zcncrypto.SecretKey {
	return &herumiSecretKey{
		SecretKey: b.JsObject.NewSecretKey(),
	}
}

func (b *herumiBls) NewPublicKey() zcncrypto.PublicKey {
	return &herumiPublicKey{
		PublicKey: b.JsObject.NewPublicKey(),
	}
}

func (b *herumiBls) NewSignature() zcncrypto.Signature {
	sg := &herumiSignature{
		Sign: b.JsObject.NewSignature(),
	}

	return sg
}

func (b *herumiBls) NewID() zcncrypto.ID {
	id := &herumiID{
		ID: b.JsObject.NewID(),
	}

	return id
}

func (b *herumiBls) SetRandFunc(randReader io.Reader) {
	b.JsObject.SetRandFunc(randReader)
}

func (b *herumiBls) FrSub(out zcncrypto.Fr, x zcncrypto.Fr, y zcncrypto.Fr) {
	out1, _ := out.(*herumiFr)
	x1, _ := x.(*herumiFr)
	y1, _ := y.(*herumiFr)

	b.JsObject.FrSub(out1.Fr, x1.Fr, y1.Fr)
}

type herumiFr struct {
	Fr JsObject
}

func (fr *herumiFr) Serialize() []byte {
	return fr.Fr.Serialize()
}

func (fr *herumiFr) SetLittleEndian(buf []byte) error {
	return fr.Fr.SetLittleEndian(buf)
}

type herumiSecretKey struct {
	SecretKey JsObject
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

func (sk *herumiSecretKey) GetPublicKey() zcncrypto.PublicKey {
	pk := sk.SecretKey.GetPublicKey()
	return &herumiPublicKey{
		PublicKey: pk,
	}
}

func (sk *herumiSecretKey) Add(rhs zcncrypto.SecretKey) {
	i, _ := rhs.(*herumiSecretKey)
	sk.SecretKey.Add(i.SecretKey)
}

func (sk *herumiSecretKey) Sign(m string) zcncrypto.Signature {
	sig := sk.SecretKey.Sign(m)

	return &herumiSignature{
		Sign: sig,
	}
}

func (sk *herumiSecretKey) GetMasterSecretKey(k int) []zcncrypto.SecretKey {
	list := sk.SecretKey.GetMasterSecretKey(k)

	msk := make([]zcncrypto.SecretKey, len(list))

	for i, it := range list {
		msk[i] = &herumiSecretKey{SecretKey: it}

	}

	return msk
}

func (sk *herumiSecretKey) Set(msk []zcncrypto.SecretKey, id zcncrypto.ID) error {

	blsMsk := make([]JsObject, len(msk))

	for i, it := range msk {
		k, ok := it.(*herumiSecretKey)
		if !ok {
			return errors.New("invalid herumi secret key")
		}

		blsMsk[i] = k.SecretKey
	}

	blsID, _ := id.(*herumiID)

	return sk.SecretKey.SetKeys(blsMsk, blsID.ID)
}

type herumiPublicKey struct {
	PublicKey JsObject
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
	Sign JsObject
}

// SerializeToHexStr --
func (sg *herumiSignature) SerializeToHexStr() string {
	return sg.Sign.SerializeToHexStr()
}

func (sg *herumiSignature) DeserializeHexStr(s string) error {
	return sg.Sign.DeserializeHexStr(s)
}

func (sg *herumiSignature) Add(rhs zcncrypto.Signature) {
	sg2, _ := rhs.(*herumiSignature)

	sg.Sign.Add(sg2.Sign)
}

func (sg *herumiSignature) Verify(pk zcncrypto.PublicKey, m string) bool {
	pub, _ := pk.(*herumiPublicKey)

	return pub.PublicKey.Verify(pub.PublicKey, m)
}

type herumiID struct {
	ID JsObject
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
