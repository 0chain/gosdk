package bls

import (
  "io"
  "bytes"
  "encoding/binary"
  "math/rand"
  "errors"
  "unsafe"
  "fmt"
  "encoding/hex"
  "github.com/0chain/gosdk/miracl"
)

func Init() error {
  if BN254.Init() == BN254.BLS_FAIL {
    return errors.New("Couldn't initialize BLS")
  }
  return nil
}

// https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L711
var sRandReader io.Reader

// Basically entirely from
// <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L729>
func SetRandFunc(randReader io.Reader) {
	// if nil, uses default random generator. See getRandomValue.
	sRandReader = randReader
}

// TODO: remove when done porting.
// // Reads a random value from the function set in `SetRandFunc`.
// func getRandomValue() (byte, error) {
//   var b [1]byte
//   var n int
//   var err error
//   if sRandReader == nil {
//     n, err = rand.Read(b[:])
//   } else {
//     n, err = sRandReader.Read(b[:])
//   }
//   if n > 0 {
//     return b[0], nil
//   }
//   return 0, errors.New("getRandomValue(): End of stream")
// }

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
  t := make([]byte, int(BN254.MODBYTES))
  E.ToBytes(t, false /*compress*/)
  return t
}

func CloneFP(fp *BN254.FP) *BN254.FP {
  result := BN254.NewFP()
  result.Copy(fp)
  return result
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

func (sig *Sign) Add(rhs *Sign) {
  sig.v.Add(rhs.v)
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

// Starting from herumi's library:
// <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L454>
func (sig *Sign) Serialize() []byte {
  return ToBytes(sig.v)
}

// Starting from herumi's library:
// <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L475>
func (sig *Sign) SerializeToHexStr() string {
	return hex.EncodeToString(sig.Serialize())
}

// Porting over <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L553>
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

func (pk *PublicKey) SerializeToHexStr() string {
  return hex.EncodeToString(pk.Serialize())
}

func (pk *PublicKey) Serialize() []byte {
  return ToBytes(pk.v)
}

//-----------------------------------------------------------------------------
// SecretKey.
//-----------------------------------------------------------------------------

// Copied directly from herumi's source code.
// SecretKey: <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L154>
// blsSecretKey: <https://github.com/herumi/bls/blob/1b48de51f4f76deb204d108f6126c1507623f739/include/bls/bls.h#L56>
// FP as Fr: <https://github.com/miracl/core/blob/master/go/FP.go#L26>
type SecretKey struct {
  v *BN254.FP
}

func SecretKey_fromFP(fp *BN254.FP) *SecretKey {
  sk := new(SecretKey)
  sk.v = fp
  return sk
}

func (sk *SecretKey) GetFP() *BN254.FP {
  return sk.v
}

func (sk *SecretKey) Clone() *SecretKey {
  result := new(SecretKey)
  result.v = sk.CloneFP()
  return result
}

func (sk *SecretKey) CloneFP() *BN254.FP {
  return BN254.NewFPcopy(sk.v)
}

func (sk *SecretKey) SetByCSPRNG() error {
  var w [BN254.NLEN]BN254.Chunk
  if sRandReader == nil {
    b := make([]byte, BN254.NLEN*int(unsafe.Sizeof(w[0])))
    rand.Read(b)
    buf := bytes.NewBuffer(b)
    binary.Read(buf, binary.LittleEndian, w)
  } else {
    binary.Read(sRandReader, binary.LittleEndian, w)
  }
  sk.v = BN254.NewFPbig(BN254.NewBIGints(w))
  return nil
}

func (sk *SecretKey) DeserializeHexStr(s string) error {
  b, err := hex2byte(s)
  if err != nil {
    return err
  }
  sk.v = BN254.FP_fromBytes(b)
  return nil
}

func (sk *SecretKey) SerializeToHexStr() string {
  return hex.EncodeToString(sk.Serialize())
}

// Note: herumi's SecretKey.GetLittleEndian is just aliased to Serialize().
func (sk *SecretKey) Serialize() []byte {
  var _a BN254.Chunk
  b := make([]byte, BN254.NLEN*int(unsafe.Sizeof(_a)))
  sk.v.ToBytes(b)
  return b
}

func (sk *SecretKey) Sign(m []byte) *Sign {
  // We're just using this miracl/core function to port over the Sign function.
  // func Core_Sign(SIG []byte, M []byte, S []byte) int {...}

  var _a BN254.Chunk

  b1 := make([]byte, int(BN254.MODBYTES))
  b3 := make([]byte, BN254.NLEN*int(unsafe.Sizeof(_a)))
  sk.v.ToBytes(b3)
  BN254.Core_Sign(b1, m, b3)

  sig := new(Sign)
  sig.v = BN254.ECP_fromBytes(b1)
  return sig
}

// Turns out this is just MPIN_GET_SERVER_SECRET
func (sk *SecretKey) GetPublicKey() *PublicKey {
  // Taken from:
  // https://github.com/miracl/core/blob/fda3416694d153f900b617d7bc42038df34a2da6/go/TestMPIN.go#L41
  // https://github.com/miracl/core/blob/fda3416694d153f900b617d7bc42038df34a2da6/go/TestMPIN.go#L79
	const MGS = BN254.MGS
	const MFS = BN254.MFS
	const G1S = 2*MFS + 1 /* Group 1 Size */
	const G2S = 4*MFS + 1  /* Group 2 Size */
	var S [MGS]byte
	var SST [G2S]byte
  sk.v.ToBytes(S[:])
  BN254.MPIN_GET_SERVER_SECRET(S[:], SST[:])
  result := new(PublicKey)
  result.v = BN254.ECP_fromBytes(SST[:])
  return result
}

func (sk *SecretKey) Add(rhs *SecretKey) {
  sk.v.Add(rhs.v)
}

func (sk *SecretKey) SubFP(fp *BN254.FP) {
  sk.v.Sub(fp)
}
