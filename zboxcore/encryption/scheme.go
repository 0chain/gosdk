package encryption

// EncryptionScheme interface for encryption schemes
type EncryptionScheme interface {
	Initialize(mnemonic string) ([]byte, error)
	InitializeWithPrivateKey(privateKey []byte) error
	InitForEncryption(tag string)
	InitForEncryptionWithPoint(tag, point string) error
	InitForDecryption(tag string, encryptedKey string) error
	InitForDecryptionWithPoint(tag, point string) error
	Encrypt(data []byte) (*EncryptedMessage, error)
	Decrypt(*EncryptedMessage) ([]byte, error)
	ReDecrypt(D *ReEncryptedMessage) ([]byte, error)
	GetEncryptedKey() string
	GetReGenKey(encPublicKey string, tag string) (string, error)
	ReEncrypt(encMsg *EncryptedMessage, reGenKey string, clientPublicKey string) (*ReEncryptedMessage, error)
	GetPublicKey() (string, error)
	GetPrivateKey() (string, error)
	GetEncryptedKeyPoint() string
}

// NewEncryptionScheme create a new encryption scheme
func NewEncryptionScheme() EncryptionScheme {
	return new(PREEncryptionScheme)
}

// EncryptedMessage encrypted message structure
type EncryptedMessage struct {
	EncryptedKey    string
	EncryptedData   []byte
	ReEncryptionKey string
	MessageChecksum string
	OverallChecksum string
}
