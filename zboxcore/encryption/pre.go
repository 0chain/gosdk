package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/oasisprotocol/curve25519-voi/curve"
	"github.com/oasisprotocol/curve25519-voi/curve/scalar"
	"strings"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/edwards25519"
)

// Suite represents the set of functionalities needed by the package schnorr.
type Suite interface {
	kyber.Group
	kyber.Random
}

type PREEncryptionScheme struct {
	PublicKey    *curve.EdwardsPoint
	PrivateKey   *scalar.Scalar
	SuiteObj     Suite
	Tag          []byte
	T            *curve.EdwardsPoint
	Ht           *curve.EdwardsPoint
	EncryptedKey *curve.EdwardsPoint
}

type PREEncryptedMessage struct {
	EncryptedKey    *curve.EdwardsPoint `json:"-"`
	EncryptedData   []byte
	MessageChecksum []byte
	OverallChecksum []byte
	TagA            []byte
}

type ReKey struct {
	R1 *curve.EdwardsPoint
	R2 *curve.EdwardsPoint
	R3 *scalar.Scalar
}

type ReKeyBytes struct {
	R1Bytes []byte `json:"r1"`
	R2Bytes []byte `json:"r2"`
	R3Bytes []byte `json:"r3"`
}

type ReEncryptedMessage struct {
	D1 *curve.EdwardsPoint
	D2 []byte
	D3 []byte
	D4 *curve.EdwardsPoint
	D5 *curve.EdwardsPoint
}

func (u *ReKey) MarshalJSON() ([]byte, error) {
	r1Bytes, err := u.R1.MarshalBinary()
	if err != nil {
		return nil, err
	}
	r2Bytes, err := u.R2.MarshalBinary()
	if err != nil {
		return nil, err
	}
	r3Bytes, err := u.R3.MarshalBinary()
	fmt.Println(r3Bytes)
	if err != nil {
		return nil, err
	}
	return json.Marshal(&ReKeyBytes{
		R1Bytes: r1Bytes,
		R2Bytes: r2Bytes,
		R3Bytes: r3Bytes,
	})
}

func (u *ReKey) UnmarshalJSON(data []byte) error {
	rbytes := &ReKeyBytes{}
	err := json.Unmarshal(data, rbytes)
	fmt.Println("unm", err)
	if err != nil {
		return err
	}
	fmt.Println("unm 2", err)
	err = u.R1.UnmarshalBinary(rbytes.R1Bytes)
	if err != nil {
		return err
	}
	fmt.Println("unm 3", err)
	err = u.R2.UnmarshalBinary(rbytes.R2Bytes)
	if err != nil {
		return err
	}
	fmt.Println("unm 4", err, len(rbytes.R3Bytes))
	err = u.R3.UnmarshalBinary(rbytes.R3Bytes)
	fmt.Println("unm 5", err)
	if err != nil {
		return err
	}
	return nil
}

func (u *PREEncryptedMessage) MarshalJSON() ([]byte, error) {
	type Alias PREEncryptedMessage
	c1Bytes, err := u.EncryptedKey.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return json.Marshal(&struct {
		EncryptedKeyBytes []byte `json:"c1"`
		*Alias
	}{
		EncryptedKeyBytes: c1Bytes,
		Alias:             (*Alias)(u),
	})
}

func (reEncMsg *ReEncryptedMessage) Marshal() ([]byte, error) {
	D1Bytes, err := reEncMsg.D1.MarshalBinary()
	if err != nil {
		return nil, err
	}

	D4Bytes, err := reEncMsg.D4.MarshalBinary()
	if err != nil {
		return nil, err
	}

	D5Bytes, err := reEncMsg.D5.MarshalBinary()
	if err != nil {
		return nil, err
	}

	headerBytes := make([]byte, 256)
	// 44 + 88 + 44 + 44 => 220, so rest of the 36 bytes minus the commas are padding
	header := base64.StdEncoding.EncodeToString(D1Bytes) + ","+ base64.StdEncoding.EncodeToString(reEncMsg.D3)
	header += "," + base64.StdEncoding.EncodeToString(D4Bytes) + ","+ base64.StdEncoding.EncodeToString(D5Bytes)
	copy(headerBytes, header)

	return append(headerBytes, reEncMsg.D2...), nil
}

func (reEncMsg *ReEncryptedMessage) Unmarshal(data []byte) error {
	headerBytes := data[:256]
	headerBytes = bytes.Trim(headerBytes, "\x00")
	encryptedData := data[256:]

	headerString := string(headerBytes)
	headerChecksums := strings.Split(headerString, ",")
	if len(headerChecksums) != 4 {
		return errors.New("Invalid data received for unmarsalling of reEncrypted data")
	}

	d1, d3, d4, d5 := headerChecksums[0], headerChecksums[1], headerChecksums[2], headerChecksums[3]

	d1Bytes, err := base64.StdEncoding.DecodeString(d1)
	if err != nil {
		return err
	}

	d3Bytes, err := base64.StdEncoding.DecodeString(d3)
	if err != nil {
		return err
	}

	d4Bytes, err := base64.StdEncoding.DecodeString(d4)
	if err != nil {
		return err
	}

	d5Bytes, err := base64.StdEncoding.DecodeString(d5)
	if err != nil {
		return err
	}

	err = reEncMsg.D1.UnmarshalBinary(d1Bytes)
	if err != nil {
		return err
	}

	reEncMsg.D2 = encryptedData
	reEncMsg.D3 = d3Bytes

	err = reEncMsg.D4.UnmarshalBinary(d4Bytes)
	if err != nil {
		return err
	}

	err = reEncMsg.D5.UnmarshalBinary(d5Bytes)
	if err != nil {
		return err
	}

	return nil
}

func (pre *PREEncryptionScheme) Initialize(mnemonic string) error {
	suite := edwards25519.NewBlakeSHA256Ed25519()
	// Create a public/private keypair (X,x)

	key, err := scalar.New().SetRandom(crand.Reader)
	if err != nil {
		return err
	}
	pre.PrivateKey = key
	pre.PublicKey = curve.NewEdwardsPoint().Mul(curve.NewEdwardsPoint().Identity(), pre.PrivateKey)
	pre.SuiteObj = suite

	return nil
}

func (pre *PREEncryptionScheme) InitForEncryption(tag string) {
	pre.Tag = []byte(tag)

	var p curve.EdwardsPoint
	s, _ := scalar.New().SetRandom(crand.Reader)
	pre.T = p.MulBasepoint(curve.ED25519_BASEPOINT_TABLE, s)
	pre.Ht = pre.hash1(pre.SuiteObj, pre.Tag, pre.PrivateKey)      // Ht  = H1(tagA,skA)
	fmt.Println("OK", pre.Ht, pre.T)
	pre.EncryptedKey = curve.NewEdwardsPoint().Add(pre.Ht, pre.T) // C1  = T + Ht
}

func (pre *PREEncryptionScheme) InitForDecryption(tag string, encryptedKey string) error {
	pre.Tag = []byte(tag)

	keyBytes, err := base64.StdEncoding.DecodeString(encryptedKey)
	if err != nil {
		return err
	}
	p := curve.NewEdwardsPoint()
	err = p.UnmarshalBinary(keyBytes)
	if err != nil {
		return err
	}
	pre.EncryptedKey = p
	return nil
}

//--------------------------------H1: Maps to Point on Elliptic Curve---------------------------------------
func (pre *PREEncryptionScheme) hash1(s Suite, tagA []byte, skA *scalar.Scalar) *curve.EdwardsPoint {
	h := sha256.New()
	if _, err := h.Write(tagA); err != nil {
		return nil
	}
	bytes, err := skA.MarshalBinary()
	if err != nil {
		return nil
	}
	h.Write(bytes)
	h1, err := scalar.NewFromBytesModOrder(h.Sum(nil)[:32])
	fmt.Println("i am here 2", err)
	if err != nil {
		return nil
	}
	p1 := curve.NewEdwardsPoint().Mul(curve.NewEdwardsPoint().Identity(), h1)
	return p1
}

//------------------------------------------H2: Maps to string-----------------------------------------------
func (pre *PREEncryptionScheme) hash2(g kyber.Group, T *curve.EdwardsPoint) []byte {
	h := sha512.New()
	bytes, err := T.MarshalBinary()
	if err != nil {
		return nil
	}
	h.Write(bytes)
	h1 := h.Sum(nil)
	return h1
}

//------------------------------------------H3: Maps to string-----------------------------------------------
func (pre *PREEncryptionScheme) hash3(g kyber.Group, msg []byte, T *curve.EdwardsPoint) []byte {
	h := sha512.New()
	bytes, err := T.MarshalBinary()
	if err != nil {
		return nil
	}
	h.Write(bytes)
	if _, err := h.Write(msg); err != nil {
		return nil
	}
	h1 := h.Sum(nil)
	return h1
}

//------------------------------------------H4: Maps to string-----------------------------------------------
func (pre *PREEncryptionScheme) hash5(g kyber.Group, C1 *curve.EdwardsPoint, C2 []byte, C3 []byte, alp *scalar.Scalar) []byte {
	h := sha512.New()
	bytes, err := C1.MarshalBinary()
	if err != nil {
		return nil
	}
	h.Write(bytes)
	if _, err := h.Write(C2); err != nil {
		return nil
	}
	if _, err := h.Write(C3); err != nil {
		return nil
	}
	bytes, err = alp.MarshalBinary()
	if err != nil {
		return nil
	}
	h.Write(bytes)
	h1 := h.Sum(nil)
	return h1

}

//------------------------------------------H6: Maps to Scalar-----------------------------------------------
func (pre *PREEncryptionScheme) hash6(g kyber.Group, tagA []byte, skA *scalar.Scalar) *scalar.Scalar {
	h := sha512.New()
	bytes, err := skA.MarshalBinary()
	if err != nil {
		return nil
	}
	h.Write(bytes)
	if _, err := h.Write(tagA); err != nil {
		return nil
	}
	fmt.Println(len(h.Sum(nil)))
	res, err := scalar.NewFromBytesModOrder(h.Sum(nil)[:32])
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return res
}

//------------------------------------------H7: Maps to Scalar-----------------------------------------------
func (pre *PREEncryptionScheme) hash7(g kyber.Group, X *curve.EdwardsPoint, D2 []byte, D3 []byte, D4 *curve.EdwardsPoint, D5 *curve.EdwardsPoint) *scalar.Scalar {
	h := sha512.New()
	bytes, err := X.MarshalBinary()
	h.Write(bytes)
	if err != nil {
		return nil
	}
	if _, err := h.Write(D2); err != nil {
		return nil
	}
	if _, err := h.Write(D3); err != nil {
		return nil
	}
	bytes, err = D4.MarshalBinary()
	h.Write(bytes)
	if err != nil {
		return nil
	}
	bytes, err = D5.MarshalBinary()
	h.Write(bytes)
	if err != nil {
		return nil
	}
	result, err := scalar.NewFromBytesModOrder(h.Sum(nil)[:32])
	fmt.Println("from bits error", err)
	if err != nil {
		return nil
	}
	return result
}

func (pre *PREEncryptionScheme) encrypt(msg []byte) (*PREEncryptedMessage, error) {
	var C = new(PREEncryptedMessage)
	C.TagA = pre.Tag
	T := pre.T
	var g kyber.Group = pre.SuiteObj
	C.EncryptedKey = pre.EncryptedKey
	fmt.Println(1)

	key := pre.hash2(g, T)             // key = H2(T)
	fmt.Println(2)
	C2, err := pre.SymEnc(g, msg, key) // C2  = Sym.Encrypt(msg,key)
	fmt.Println(3)
	C.EncryptedData = C2
	if err != nil {
		return nil, err
	}

	C.MessageChecksum = pre.hash3(g, msg, T)                                                  // C3  = H3(msg,T)
	fmt.Println(4)
	alp := pre.hash6(g, pre.Tag, pre.PrivateKey)                                              // alp = H6(tagA,skA)
	fmt.Println(5, alp)
	C.OverallChecksum = pre.hash5(g, C.EncryptedKey, C.EncryptedData, C.MessageChecksum, alp) // C4  = H5(C1,C2,C3,alp)
	fmt.Println(6)
	return C, nil                                                                             // return C = (C1,C2,C3,C4,tagA)
}

//---------------------------------Symmetric Encryption using AES with GCM mode---------------------------------
func (pre *PREEncryptionScheme) SymEnc(group kyber.Group, message []byte, keyhash []byte) ([]byte, error) {
	len := 32 + 12
	key := keyhash[:32]
	nonce := keyhash[32:len]
	aes, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(aes)
	if err != nil {
		return nil, err
	}
	c := aesgcm.Seal(nil, nonce, message, nil)
	var ctx bytes.Buffer
	_, err = ctx.Write(c)
	if err != nil {
		return nil, err
	}
	return ctx.Bytes(), nil
}

func UnmarshallPublicKey(publicKey string) (*curve.EdwardsPoint, error) {
	point := curve.NewEdwardsPoint()
	decoded, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return nil, err
	}
	err = point.UnmarshalBinary([]byte(decoded))
	if err != nil {
		return nil, err
	}
	return point, nil
}

//---------------------------------Symmetric Decryption using AES with GCM mode---------------------------------
func (pre *PREEncryptionScheme) SymDec(group kyber.Group, ctx []byte, keyhash []byte) ([]byte, error) {
	len := 32 + 12
	key := keyhash[:32]
	nonce := keyhash[32:len]

	// Decrypt message using AES-GCM
	aes, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(aes)
	if err != nil {
		return nil, err
	}
	return aesgcm.Open(nil, nonce, ctx, nil)
}

func (pre *PREEncryptionScheme) Encrypt(data []byte) (*EncryptedMessage, error) {
	//condA := []byte("filetype:audio")
	encryptedMsg, err := pre.encrypt(data)
	if err != nil {
		return nil, err
	}
	// msgData, err := json.Marshal(encryptedMsg)
	// if err != nil {
	// 	return nil, err
	// }
	return &EncryptedMessage{EncryptedData: encryptedMsg.EncryptedData, EncryptedKey: pre.GetEncryptedKey(), MessageChecksum: hex.EncodeToString(encryptedMsg.MessageChecksum), OverallChecksum: hex.EncodeToString(encryptedMsg.OverallChecksum)}, err
	//return encryptedMsg.EncryptedData, err
}

func (pre *PREEncryptionScheme) decrypt(encMsg *EncryptedMessage) ([]byte, error) {
	var g kyber.Group = pre.SuiteObj
	C := &PREEncryptedMessage{}
	C.EncryptedKey = pre.EncryptedKey
	C.EncryptedData = encMsg.EncryptedData
	decodedChecksum, err := hex.DecodeString(encMsg.MessageChecksum)
	if err != nil {
		return nil, err
	}
	C.MessageChecksum = decodedChecksum

	decodedChecksum, err = hex.DecodeString(encMsg.OverallChecksum)
	if err != nil {
		return nil, err
	}
	C.OverallChecksum = decodedChecksum
	C.TagA = pre.Tag
	// err := json.Unmarshal(data, C)
	// if err != nil {
	// 	return nil, err
	// }
	alp := pre.hash6(g, C.TagA, pre.PrivateKey) // alp = H6(tagA,skA)
	chk1 := pre.hash5(g, C.EncryptedKey, C.EncryptedData, C.MessageChecksum, alp)
	if !bytes.Equal(chk1, C.OverallChecksum) { // Check if C4 = H5(C1,C2,C3,alp)
		return nil, fmt.Errorf("Invalid Ciphertext in decrypt, C4 != H5")
	}
	Ht := pre.hash1(pre.SuiteObj, pre.Tag, pre.PrivateKey) // Ht  = H1(tagA,skA)
	T := curve.NewEdwardsPoint().Sub(C.EncryptedKey, Ht)
	key := pre.hash2(g, T)                                 // key = H2(T)
	recmsg, err2 := pre.SymDec(g, C.EncryptedData, key)    // recover message using Sym.Decrypt(C2,key)
	if err2 == nil {
		chk2 := pre.hash3(g, recmsg, T)
		if !bytes.Equal(chk2, C.MessageChecksum) { // Check if C3 = H3(m,T)
			return nil, fmt.Errorf("Invalid Ciphertext in decrypt, C3 != H3")
		} else {
			//fmt.Println("First level ciphertext decrypted successfully")
			return recmsg, nil
		}
	}
	return nil, err2
}

func (pre *PREEncryptionScheme) ReEncrypt(encMsg *EncryptedMessage, reGenKey string, clientPublicKey string) (*ReEncryptedMessage, error) {
	key, err := UnmarshallPublicKey(clientPublicKey)
	if err != nil {
		return nil, err
	}
	return pre.reEncrypt(encMsg, reGenKey, key)
}

//-----------------------------------------------ReEncryption-------------------------------------------------
//reencrypt the data, cancelling the previous encryption by using the new regenkey
func (pre *PREEncryptionScheme) reEncrypt(encMsg *EncryptedMessage, reGenKey string, clientPublicKey *curve.EdwardsPoint) (*ReEncryptedMessage, error) {
	var g kyber.Group = pre.SuiteObj
	C := &PREEncryptedMessage{}
	C.EncryptedKey = pre.EncryptedKey
	C.EncryptedData = encMsg.EncryptedData
	decodedChecksum, err := hex.DecodeString(encMsg.MessageChecksum)
	if err != nil {
		return nil, err
	}
	C.MessageChecksum = decodedChecksum

	decodedChecksum, err = hex.DecodeString(encMsg.OverallChecksum)
	if err != nil {
		return nil, err
	}
	C.OverallChecksum = decodedChecksum
	C.TagA = pre.Tag

	rk := new(ReKey)
	rk.R1 = curve.NewEdwardsPoint()
	rk.R2 = curve.NewEdwardsPoint()
	rk.R3 = scalar.New()
	err = rk.UnmarshalJSON([]byte(reGenKey))
	fmt.Println(888, err)
	if err != nil {
		return nil, err
	}

	var reEncMsg = new(ReEncryptedMessage)

	fmt.Println(889)
	chk1 := pre.hash5(g, C.EncryptedKey, C.EncryptedData, C.MessageChecksum, rk.R3)
	fmt.Println(890)
	if !bytes.Equal(chk1, C.OverallChecksum) { // Check if C4 = H5(C1,C2,C3,alp)
		return nil, fmt.Errorf("Invalid Ciphertext in reEncrypt, C4 != H5")
	}
	fmt.Println(891)
	t, err := scalar.New().SetRandom(crand.Reader)
	fmt.Println(892, err)
	if err != nil {
		return nil, err
	}
	fmt.Println(893)
	reEncMsg.D5 = curve.NewEdwardsPoint().Mul(curve.NewEdwardsPoint().Identity(), t)    // D5    = tP
	fmt.Println(894)
	tXj := curve.NewEdwardsPoint().Mul(clientPublicKey, t)
	fmt.Println(895)
	reEncMsg.D1 = curve.NewEdwardsPoint().Add(C.EncryptedKey, rk.R1)
	fmt.Println(896)
	reEncMsg.D2 = C.EncryptedData                                                // D2    = C2
	reEncMsg.D3 = C.MessageChecksum                                              // D3    = C3
	reEncMsg.D4 = rk.R2                                                          // D4    = R2
	bet := pre.hash7(g, tXj, reEncMsg.D2, reEncMsg.D3, reEncMsg.D4, reEncMsg.D5) // bet   = H7(tXj,D2,D3,D4,D5)
	fmt.Println(897, bet)
	reEncMsg.D1 = curve.NewEdwardsPoint().Mul(reEncMsg.D1, bet)                    // D1    = bet.(C1 + R1)
	fmt.Println(898)
	return reEncMsg, nil                                                         // Return D = (D1,D2,D3,D4,D5)
}

//-----------------------------------------------ReDecryption-------------------------------------------------
func (pre *PREEncryptionScheme) ReDecrypt(D *ReEncryptedMessage) ([]byte, error) {
	s := pre.SuiteObj
	tXj := curve.NewEdwardsPoint().Mul(D.D5, pre.PrivateKey) // tXj   = skB.D5
	var g kyber.Group = s
	bet := pre.hash7(g, tXj, D.D2, D.D3, D.D4, D.D5) // bet   = H7(tXj,D2,D3,D4,D5)
	binv := scalar.New().Invert(bet)
	xinv := scalar.New().Invert(pre.PrivateKey)

	T1 := curve.NewEdwardsPoint().Mul(D.D1, binv)
	T2 := curve.NewEdwardsPoint().Mul(D.D4, xinv)
	T := curve.NewEdwardsPoint().Sub(T1, T2) // T     = bet^(-1).D1 - skB^(-1).D4
	key := pre.hash2(g, T)     // key   = H2(T)

	recmsg, err2 := pre.SymDec(g, D.D2, key) // recover message using Sym.Decrypt(D2,key)
	if err2 == nil {
		chk2 := pre.hash3(g, recmsg, T)
		if !bytes.Equal(chk2, D.D3) { // Check if D3 = H3(m,T)
			return nil, fmt.Errorf("Invalid Ciphertext in reDecrypt, D3 != H3")
		} else {
			return recmsg, nil
		}
	}
	return nil, err2
}

func (pre *PREEncryptionScheme) Decrypt(encMsg *EncryptedMessage) ([]byte, error) {
	if len(encMsg.ReEncryptionKey) > 0 {
		fmt.Println("OK")
		reEncMsg, err := pre.reEncrypt(encMsg, encMsg.ReEncryptionKey, pre.PublicKey)
		fmt.Println("Done")
		if err != nil {
			return nil, err
		}
		decryptedMessage, err := pre.ReDecrypt(reEncMsg)
		if err != nil {
			return nil, err
		}
		return decryptedMessage, nil
	}
	decryptedMessage, err := pre.decrypt(encMsg)
	if err != nil {
		return nil, err
	}
	return decryptedMessage, nil
}

func (pre *PREEncryptionScheme) GetEncryptedKey() string {
	keyBytes, _ := pre.EncryptedKey.MarshalBinary()
	keyString := base64.StdEncoding.EncodeToString(keyBytes)
	return keyString
}

func (pre *PREEncryptionScheme) GetPublicKey() (string, error) {
	keyBytes, err := pre.PublicKey.MarshalBinary()
	if err != nil {
		return "", err
	}
	keyString := base64.StdEncoding.EncodeToString(keyBytes)
	return keyString, nil
}

func (pre *PREEncryptionScheme) GetPrivateKey() (string, error) {
	keyBytes, err := pre.PrivateKey.MarshalBinary()
	if err != nil {
		return "", err
	}
	keyString := base64.StdEncoding.EncodeToString(keyBytes)
	return keyString, nil
}

func (pre *PREEncryptionScheme) GetReGenKey(encPublicKey string, tag string) (string, error) {
	condA := []byte(tag)
	var RK = new(ReKey)
	var g kyber.Group = pre.SuiteObj
	r, err := scalar.New().SetRandom(crand.Reader)
	if err != nil {
		return "", err
	}
	Hc := pre.hash1(pre.SuiteObj, condA, pre.PrivateKey)         // Hc   = H1(condA,skA)
	RK.R1 = curve.NewEdwardsPoint().Mul(curve.NewEdwardsPoint().Identity(), r)
	RK.R1 = curve.NewEdwardsPoint().Sub(RK.R1, Hc) // R1   = rP - Hc

	keyBytes, err := base64.StdEncoding.DecodeString(encPublicKey)
	if err != nil {
		return "", err
	}
	p := curve.NewEdwardsPoint()
	err = p.UnmarshalBinary(keyBytes)
	if err != nil {
		return "", err
	}

	RK.R2 = curve.NewEdwardsPoint().Mul(p, r)      // R2   = r.pkB
	RK.R3 = pre.hash6(g, condA, pre.PrivateKey) // R3   = H6(condA,skA)
	rkBytes, err := RK.MarshalJSON()
	if err != nil {
		return "", err
	}
	return string(rkBytes), nil
}
