package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/edwards25519"
)

// Suite represents the set of functionalities needed by the package schnorr.
type Suite interface {
	kyber.Group
	kyber.Random
}

type PREEncryptionScheme struct {
	PublicKey    kyber.Point
	PrivateKey   kyber.Scalar
	SuiteObj     Suite
	Tag          []byte
	T            kyber.Point
	Ht           kyber.Point
	EncryptedKey kyber.Point
}

type PREEncryptedMessage struct {
	EncryptedKey    kyber.Point `json:"-"`
	EncryptedData   []byte
	MessageChecksum []byte
	OverallChecksum []byte
	TagA            []byte
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

func (pre *PREEncryptionScheme) Initialize(mnemonic string) error {
	suite := edwards25519.NewBlakeSHA256Ed25519()
	rand := suite.XOF([]byte(mnemonic))

	// Create a public/private keypair (X,x)
	pre.PrivateKey = suite.Scalar().Pick(rand)
	pre.PublicKey = suite.Point().Mul(pre.PrivateKey, nil)
	pre.SuiteObj = suite

	return nil
}

func (pre *PREEncryptionScheme) InitForEncryption(tag string) {
	pre.Tag = []byte(tag)

	pre.T = pre.SuiteObj.Point().Pick(pre.SuiteObj.RandomStream()) // Pick a random point T from the curve
	pre.Ht = pre.hash1(pre.SuiteObj, pre.Tag, pre.PrivateKey)      // Ht  = H1(tagA,skA)
	var g kyber.Group = pre.SuiteObj
	pre.EncryptedKey = g.Point().Add(pre.Ht, pre.T) // C1  = T + Ht
}

func (pre *PREEncryptionScheme) InitForDecryption(tag string, encryptedKey string) error {
	pre.Tag = []byte(tag)

	var g kyber.Group = pre.SuiteObj
	keyBytes, err := base64.StdEncoding.DecodeString(encryptedKey)
	if err != nil {
		return err
	}
	p := g.Point()
	err = p.UnmarshalBinary(keyBytes)
	if err != nil {
		return err
	}
	pre.EncryptedKey = p
	return nil
}

//--------------------------------H1: Maps to Point on Elliptic Curve---------------------------------------
func (pre *PREEncryptionScheme) hash1(s Suite, tagA []byte, skA kyber.Scalar) kyber.Point {
	var g kyber.Group = s
	h := sha256.New()
	if _, err := h.Write(tagA); err != nil {
		return nil
	}
	if _, err := skA.MarshalTo(h); err != nil {
		return nil
	}
	h1 := g.Scalar().SetBytes(h.Sum(nil))
	p1 := g.Point().Mul(h1, nil)
	return p1
}

//------------------------------------------H2: Maps to string-----------------------------------------------
func (pre *PREEncryptionScheme) hash2(g kyber.Group, T kyber.Point) []byte {
	h := sha512.New()
	if _, err := T.MarshalTo(h); err != nil {
		return nil
	}
	h1 := h.Sum(nil)
	return h1
}

//------------------------------------------H3: Maps to string-----------------------------------------------
func (pre *PREEncryptionScheme) hash3(g kyber.Group, msg []byte, T kyber.Point) []byte {
	h := sha512.New()
	if _, err := T.MarshalTo(h); err != nil {
		return nil
	}
	if _, err := h.Write(msg); err != nil {
		return nil
	}
	h1 := h.Sum(nil)
	return h1
}

//------------------------------------------H4: Maps to string-----------------------------------------------
func (pre *PREEncryptionScheme) hash5(g kyber.Group, C1 kyber.Point, C2 []byte, C3 []byte, alp kyber.Scalar) []byte {
	h := sha512.New()
	if _, err := C1.MarshalTo(h); err != nil {
		return nil
	}
	if _, err := h.Write(C2); err != nil {
		return nil
	}
	if _, err := h.Write(C3); err != nil {
		return nil
	}
	if _, err := alp.MarshalTo(h); err != nil {
		return nil
	}
	h1 := h.Sum(nil)
	return h1

}

//------------------------------------------H6: Maps to Scalar-----------------------------------------------
func (pre *PREEncryptionScheme) hash6(g kyber.Group, tagA []byte, skA kyber.Scalar) kyber.Scalar {
	h := sha512.New()
	if _, err := skA.MarshalTo(h); err != nil {
		return nil
	}
	if _, err := h.Write(tagA); err != nil {
		return nil
	}
	return g.Scalar().SetBytes(h.Sum(nil))
}

//------------------------------------------H7: Maps to Scalar-----------------------------------------------
func (pre *PREEncryptionScheme) hash7(g kyber.Group, X kyber.Point, D2 []byte, D3 []byte, D4 kyber.Point, D5 kyber.Point) kyber.Scalar {
	h := sha512.New()
	if _, err := X.MarshalTo(h); err != nil {
		return nil
	}
	if _, err := h.Write(D2); err != nil {
		return nil
	}
	if _, err := h.Write(D3); err != nil {
		return nil
	}
	if _, err := D4.MarshalTo(h); err != nil {
		return nil
	}
	if _, err := D5.MarshalTo(h); err != nil {
		return nil
	}
	return g.Scalar().SetBytes(h.Sum(nil))
}

func (pre *PREEncryptionScheme) encrypt(msg []byte) (*PREEncryptedMessage, error) {
	var C = new(PREEncryptedMessage)
	C.TagA = pre.Tag
	T := pre.T
	var g kyber.Group = pre.SuiteObj
	C.EncryptedKey = pre.EncryptedKey

	key := pre.hash2(g, T)             // key = H2(T)
	C2, err := pre.SymEnc(g, msg, key) // C2  = Sym.Encrypt(msg,key)
	C.EncryptedData = C2
	if err != nil {
		return nil, err
	}

	C.MessageChecksum = pre.hash3(g, msg, T)                                                  // C3  = H3(msg,T)
	alp := pre.hash6(g, pre.Tag, pre.PrivateKey)                                              // alp = H6(tagA,skA)
	C.OverallChecksum = pre.hash5(g, C.EncryptedKey, C.EncryptedData, C.MessageChecksum, alp) // C4  = H5(C1,C2,C3,alp)

	return C, nil // return C = (C1,C2,C3,C4,tagA)
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
		return nil, fmt.Errorf("Invalid Ciphertext")
	}
	Ht := pre.hash1(pre.SuiteObj, pre.Tag, pre.PrivateKey) // Ht  = H1(tagA,skA)
	T := g.Point().Sub(C.EncryptedKey, Ht)                 // T   = C1 - Ht
	key := pre.hash2(g, T)                                 // key = H2(T)
	recmsg, err2 := pre.SymDec(g, C.EncryptedData, key)    // recover message using Sym.Decrypt(C2,key)
	if err2 == nil {
		chk2 := pre.hash3(g, recmsg, T)
		if !bytes.Equal(chk2, C.MessageChecksum) { // Check if C3 = H3(m,T)
			return nil, fmt.Errorf("Invalid Ciphertext")
		} else {
			//fmt.Println("First level ciphertext decrypted successfully")
			return recmsg, nil
		}
	}
	return nil, err2
}

func (pre *PREEncryptionScheme) Decrypt(encMsg *EncryptedMessage) ([]byte, error) {
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
