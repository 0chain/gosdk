package bls

import (
  "fmt"
  "encoding/hex"
  "github.com/0chain/gosdk/miracl"
)

// Taken directly from herumi's library:
// <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L29>
func hex2byte(s string) ([]byte, error) {
	if (len(s) & 1) == 1 {
		return nil, fmt.Errorf("odd length")
	}
	return hex.DecodeString(s)
}

func DeserializeHexStr(s string) (*BN254.ECP, error) {
  b, err := hex2byte(s)
	if err != nil {
		return nil, err
	}
  return BN254.ECP_fromBytes(b), nil
}

func ToBytes(E *BN254.ECP) []byte {
  var t [int(BN254.MODBYTES)]byte
  var R = t[:]
  E.ToBytes(R, false /*compress*/)
  return R
}

//-----------------------------------------------------------------------------
// Signature.
//-----------------------------------------------------------------------------

// Copied directly from herumi's source code.
// Sign: <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L449>
// blsSignature: <https://github.com/herumi/bls/blob/1b48de51f4f76deb204d108f6126c1507623f739/include/bls/bls.h#L68>
// mclBnG1: <https://github.com/herumi/mcl/blob/0114a3029f74829e79dc51de6dfb28f5da580632/include/mcl/bn.h#L96>
type Sign struct {
  v *BN254.ECP
}

// Starting from herumi's library:
// <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L480>
func (sig *Sign) DeserializeHexStr(s string) error {
  var err error
  sig.v, err = DeserializeHexStr(s);
  if err != nil {
    return err
  }
  return nil
}

func (sig *Sign) Verify(pub *PublicKey, m []byte) bool {
  b := BN254.Core_Verify(ToBytes(sig.v), m, ToBytes(pub.v))
  return b == BN254.BLS_OK
}

//-----------------------------------------------------------------------------
// PublicKey.
//-----------------------------------------------------------------------------

// Copied directly from herumi's source code.
// PublicKey: <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L334>
// blsPublicKey: <https://github.com/herumi/bls/blob/1b48de51f4f76deb204d108f6126c1507623f739/include/bls/bls.h#L60>
// mclBnG1: <https://github.com/herumi/mcl/blob/0114a3029f74829e79dc51de6dfb28f5da580632/include/mcl/bn.h#L96>
type PublicKey struct {
  v *BN254.ECP
}

// Starting from herumi's library:
// <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L480>
func (pk *PublicKey) DeserializeHexStr(s string) error {
  var err error
  pk.v, err = DeserializeHexStr(s);
  if err != nil {
    return err
  }
  return nil
}

func Init() {
  fmt.Println("hello world")
}
