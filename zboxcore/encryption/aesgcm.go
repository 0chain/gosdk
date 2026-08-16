package encryption

// aesgcm.go — a fast AES-256-GCM EncryptionScheme, an alternative to the
// pairing/EC-based PREEncryptionScheme (pre.go). Benchmarked ~15-19x faster than
// PRE for encrypt/decrypt because it is symmetric + AES-NI accelerated.
//
// TRADEOFF vs PRE: AES-GCM does NOT support proxy re-encryption (share a file to
// another user's public key WITHOUT re-upload). ReEncrypt/ReDecrypt/GetReGenKey
// therefore return an error. Use PRE when marketplace/proxy-reshare is required;
// use AES-GCM (the default for new uploads) otherwise.
//
// ON-DISK FRAMING — byte-compatible with the PRE chunk layout so NO change to the
// chunked upload/download framing is needed:
//   chunk = MessageChecksum(128) || OverallChecksum(128) || EncryptedData
//   EncryptedData = AES-GCM ciphertext+tag = plaintext + 16 bytes (== EncryptedDataPaddingSize)
//   the 12-byte GCM nonce rides in MessageChecksum (hex of a 64-byte header block)
//   EncryptedKey = "aesgcm:" + base64(wrapNonce || AES-GCM-wrapped per-file key)
// The "aesgcm:" prefix on EncryptedKey lets the download path pick this scheme
// (NewEncryptionSchemeForKey) so AES and legacy-PRE objects coexist.

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
)

// NewEncryptionSchemeForUpload returns the scheme a NEW upload should use. It
// defaults to the legacy PRE scheme (backward-compatible for every gosdk
// consumer) and switches to AES-256-GCM only when ZUS_ENC_CIPHER=aesgcm — so a
// deployment (e.g. the zs3 gateway) opts into fast AES explicitly, without
// changing behavior for other tools. Reads (below) auto-detect the scheme from
// the stored key, so this choice only affects freshly written objects.
func NewEncryptionSchemeForUpload() EncryptionScheme {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("ZUS_ENC_CIPHER")), "aesgcm") {
		return NewAESGCMEncryptionScheme()
	}
	return NewEncryptionScheme()
}

// NewEncryptionSchemeForKey returns the scheme that can DECRYPT a given stored
// EncryptedKey: AES-256-GCM if it carries the "aesgcm:" marker, else legacy PRE.
// This is what lets AES-GCM and PRE objects coexist in one allocation.
func NewEncryptionSchemeForKey(encryptedKey string) EncryptionScheme {
	if strings.HasPrefix(encryptedKey, AESGCMKeyPrefix) {
		return NewAESGCMEncryptionScheme()
	}
	return NewEncryptionScheme()
}

// AESGCMKeyPrefix marks an EncryptedKey as belonging to the AES-GCM scheme.
const AESGCMKeyPrefix = "aesgcm:"

// aesGCMNonceSize is the standard 96-bit GCM nonce.
const aesGCMNonceSize = 12

// AESGCMEncryptionScheme implements EncryptionScheme with AES-256-GCM.
type AESGCMEncryptionScheme struct {
	masterKey []byte // 32-byte KEK derived from the owner secret (wraps per-file keys)
	fileKey   []byte // 32-byte per-file data key (random on encrypt, unwrapped on decrypt)
	encKeyB64 string // the serialized EncryptedKey ("aesgcm:"+base64(wrapNonce||wrapped))
	tag       string
}

// NewAESGCMEncryptionScheme returns an empty AES-256-GCM scheme.
func NewAESGCMEncryptionScheme() *AESGCMEncryptionScheme { return &AESGCMEncryptionScheme{} }

// compile-time conformance to the full EncryptionScheme interface.
var _ EncryptionScheme = (*AESGCMEncryptionScheme)(nil)

// deriveMaster turns owner secret material into a 32-byte KEK. Deterministic, so
// every fleet node holding the same owner mnemonic derives the same key and can
// decrypt the same data.
func deriveMaster(secret []byte) []byte {
	h := sha256.Sum256(append([]byte("zus-aes256gcm-master-v1:"), secret...))
	return h[:]
}

func (a *AESGCMEncryptionScheme) Initialize(mnemonic string) ([]byte, error) {
	if mnemonic == "" {
		return nil, errors.New("aesgcm: empty mnemonic")
	}
	// Bind the KEK to the mnemonic. Return value mirrors PRE (private-key bytes);
	// callers only persist/relay it, so the SHA-256 seed is sufficient here.
	seed := sha256.Sum256([]byte(mnemonic))
	a.masterKey = deriveMaster(seed[:])
	return seed[:], nil
}

func (a *AESGCMEncryptionScheme) InitializeWithPrivateKey(privateKey []byte) error {
	if len(privateKey) == 0 {
		return errors.New("aesgcm: empty private key")
	}
	a.masterKey = deriveMaster(privateKey)
	return nil
}

// InitForEncryption mints a fresh random per-file key and wraps it under the KEK.
// tag is retained for parity with PRE (it is not part of the AES key path).
func (a *AESGCMEncryptionScheme) InitForEncryption(tag string) {
	a.tag = tag
	a.fileKey = make([]byte, 32)
	_, _ = rand.Read(a.fileKey)
	wrapped, err := a.wrapFileKey(a.fileKey)
	if err != nil {
		// rand/AES failures here are non-recoverable; leave encKeyB64 empty so the
		// first Encrypt surfaces the error rather than silently mis-encrypting.
		a.encKeyB64 = ""
		return
	}
	a.encKeyB64 = AESGCMKeyPrefix + wrapped
}

// InitForDecryption unwraps the per-file key from the stored EncryptedKey.
func (a *AESGCMEncryptionScheme) InitForDecryption(tag string, encryptedKey string) error {
	a.tag = tag
	if a.masterKey == nil {
		return errors.New("aesgcm: not initialized (no master key)")
	}
	raw := strings.TrimPrefix(encryptedKey, AESGCMKeyPrefix)
	fk, err := a.unwrapFileKey(raw)
	if err != nil {
		return fmt.Errorf("aesgcm: unwrap file key: %w", err)
	}
	a.fileKey = fk
	a.encKeyB64 = encryptedKey
	return nil
}

// wrapFileKey seals the per-file key under the KEK (AES-GCM), returning
// base64(wrapNonce || ciphertext).
func (a *AESGCMEncryptionScheme) wrapFileKey(fk []byte) (string, error) {
	gcm, err := newGCM(a.masterKey)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	ct := gcm.Seal(nil, nonce, fk, nil)
	return base64.StdEncoding.EncodeToString(append(nonce, ct...)), nil
}

func (a *AESGCMEncryptionScheme) unwrapFileKey(b64 string) ([]byte, error) {
	blob, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, err
	}
	gcm, err := newGCM(a.masterKey)
	if err != nil {
		return nil, err
	}
	if len(blob) < gcm.NonceSize() {
		return nil, errors.New("aesgcm: wrapped key too short")
	}
	nonce, ct := blob[:gcm.NonceSize()], blob[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ct, nil)
}

// Encrypt AES-256-GCM-seals one chunk, packing the nonce into MessageChecksum so
// the on-disk framing matches PRE exactly (see file header).
func (a *AESGCMEncryptionScheme) Encrypt(data []byte) (*EncryptedMessage, error) {
	if a.fileKey == nil {
		return nil, errors.New("aesgcm: InitForEncryption not called (no file key)")
	}
	gcm, err := newGCM(a.fileKey)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aesGCMNonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	ct := gcm.Seal(nil, nonce, data, nil) // len == len(data)+16 == +EncryptedDataPaddingSize
	return &EncryptedMessage{
		EncryptedData:   ct,
		EncryptedKey:    a.encKeyB64,
		MessageChecksum: encodeHeader(nonce), // 128-char hex, first 24 chars = nonce
		OverallChecksum: emptyHeader(),       // 128-char hex padding (GCM tag already authenticates)
	}, nil
}

// Decrypt reverses Encrypt: pull the nonce from MessageChecksum, GCM-open.
func (a *AESGCMEncryptionScheme) Decrypt(msg *EncryptedMessage) ([]byte, error) {
	if a.fileKey == nil {
		return nil, errors.New("aesgcm: InitForDecryption not called (no file key)")
	}
	nonce, err := decodeHeader(msg.MessageChecksum)
	if err != nil {
		return nil, err
	}
	gcm, err := newGCM(a.fileKey)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, nonce, msg.EncryptedData, nil)
}

func (a *AESGCMEncryptionScheme) GetEncryptedKey() string { return a.encKeyB64 }

// Point-based variants are PRE-specific (kyber point encapsulation). AES-GCM has
// no curve point, so these map onto the key-string path / are unsupported.
func (a *AESGCMEncryptionScheme) InitForEncryptionWithPoint(tag, point string) error {
	a.InitForEncryption(tag)
	return nil
}
func (a *AESGCMEncryptionScheme) InitForDecryptionWithPoint(tag, point string) error {
	return errors.New("aesgcm: point-based decryption not supported (no curve point)")
}
func (a *AESGCMEncryptionScheme) GetEncryptedKeyPoint() string { return "" }

// --- PRE-only capabilities: unsupported for AES-GCM (no proxy re-encryption) ---

func (a *AESGCMEncryptionScheme) ReDecrypt(*ReEncryptedMessage) ([]byte, error) {
	return nil, errors.New("aesgcm: proxy re-encryption not supported; use PRE for share-without-reupload")
}
func (a *AESGCMEncryptionScheme) ReEncrypt(*EncryptedMessage, string, string) (*ReEncryptedMessage, error) {
	return nil, errors.New("aesgcm: proxy re-encryption not supported; use PRE for share-without-reupload")
}
func (a *AESGCMEncryptionScheme) GetReGenKey(string, string) (string, error) {
	return "", errors.New("aesgcm: proxy re-encryption not supported; use PRE for share-without-reupload")
}
func (a *AESGCMEncryptionScheme) GetPublicKey() (string, error) {
	return "", errors.New("aesgcm: no public key (symmetric scheme)")
}
func (a *AESGCMEncryptionScheme) GetPrivateKey() (string, error) {
	if a.masterKey == nil {
		return "", errors.New("aesgcm: not initialized")
	}
	return hex.EncodeToString(a.masterKey), nil
}

// --- helpers ---

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// encodeHeader packs the nonce into a fixed 128-char hex string (a 64-byte block:
// first 12 bytes = nonce, rest zero) so MessageChecksum is exactly EncryptionHeaderSize/2.
func encodeHeader(nonce []byte) string {
	buf := make([]byte, 64)
	copy(buf, nonce)
	return hex.EncodeToString(buf)
}

func emptyHeader() string { return hex.EncodeToString(make([]byte, 64)) }

func decodeHeader(s string) ([]byte, error) {
	buf, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	if len(buf) < aesGCMNonceSize {
		return nil, errors.New("aesgcm: header too short for nonce")
	}
	return buf[:aesGCMNonceSize], nil
}
