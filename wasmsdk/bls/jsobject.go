//go:build js && wasm
// +build js,wasm

package bls

import (
	"encoding/hex"
	"fmt"
	"io"
	"syscall/js"
)

type JsObject struct {
	js.Value
}

func NewJsObject(v js.Value) JsObject {
	return JsObject{
		Value: v,
	}
}

func (jo JsObject) Init(curve int) {
	//window.bls.BN254 = 0
	jo.Value.Call("init", js.ValueOf(curve))
}

func (jo JsObject) SetRandFunc(reader io.Reader) {
	//FIXME: not sure how to pass io.Reader to js
}

func (jo JsObject) FrSub(out JsObject, x JsObject, y JsObject) {

	//FIXME: _mclBnFr_sub is not exported on bls.js yet
	jo.Value.Call("_mclBnFr_sub", out.Value, x.Value, y.Value)
}

func (jo JsObject) NewFr() JsObject {
	return NewJsObject(js.Global().Get("Fr").New())
}

func (jo JsObject) NewSecretKey() JsObject {
	return NewJsObject(js.Global().Get("SecretKey").New())
}

func (jo JsObject) NewPublicKey() JsObject {
	return NewJsObject(js.Global().Get("PublicKey").New())
}

func (jo JsObject) NewSignature() JsObject {
	return NewJsObject(js.Global().Get("Signature").New())
}

func (jo JsObject) NewID() JsObject {
	return NewJsObject(js.Global().Get("Id").New())
}

func (jo JsObject) Serialize() []byte {
	buf := jo.Value.Call("serialize")
	var dst []byte
	js.CopyBytesToGo(dst, buf)

	return dst
}

func (jo JsObject) Deserialize(buf []byte) {

	var dst js.Value

	js.CopyBytesToJS(dst, buf)

	jo.Value.Call("deserialize", dst)

}

func (jo JsObject) SerializeToHexStr() string {
	buf := jo.Serialize()
	return hex.EncodeToString(buf)
}

func (jo JsObject) DeserializeHexStr(s string) error {
	buf, err := hex2byte(s)
	if err != nil {
		return err
	}

	jo.Deserialize(buf)

	return nil
}

func (jo JsObject) SetLittleEndian(buf []byte) error {
	var dst js.Value

	js.CopyBytesToJS(dst, buf)

	jo.Value.Call("setLittleEndian", dst)

	return nil
}

func (jo JsObject) GetLittleEndian() []byte {
	//FIXME: _mclBnFr_getLittleEndian is not exported on bls.js yet
	buf := jo.Value.Call("getLittleEndian")
	var dst []byte
	js.CopyBytesToGo(dst, buf)

	return dst
}

func (jo JsObject) SetByCSPRNG() {
	jo.Value.Call("setByCSPRNG")
}

func (jo JsObject) Add(sign JsObject) {
	jo.Value.Call("add", sign.Value)
}

func (jo JsObject) GetPublicKey() JsObject {
	pk := jo.Value.Call("getPublicKey")

	return NewJsObject(pk)
}

func (jo JsObject) Sign(m string) JsObject {
	sign := jo.Value.Call("sign", js.ValueOf(m))

	return NewJsObject(sign)
}

func (jo JsObject) Verify(sign JsObject, m string) bool {
	jv := jo.Value.Call("verify", sign.Value, js.ValueOf(m))

	if jv.IsNull() {
		return true
	}

	return false

}

func (jo JsObject) GetMasterSecretKey(k int) []JsObject {
	//FIXME: GetMasterSecretKey is not exported on bls_c.js yet
	return nil
}

func (jo JsObject) SetKeys(msk []JsObject, id JsObject) error {
	//FIXME: Set is not exported on bls_c.js yet
	return nil
}

func hex2byte(s string) ([]byte, error) {
	if (len(s) & 1) == 1 {
		return nil, fmt.Errorf("odd length")
	}
	return hex.DecodeString(s)
}
