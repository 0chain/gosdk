package encryption

type EncryptionScheme interface {
	Initialize(mnemonic string) ([]byte, error)
	InitializeWithPrivateKey(privateKey []byte) error
	InitForEncryption(tag string)
	InitForDecryption(tag string, encryptedKey string) error
	Encrypt(data []byte) (*EncryptedMessage, error)
	Decrypt(*EncryptedMessage) ([]byte, error)
	ReDecrypt(D *ReEncryptedMessage) ([]byte, error)
	GetEncryptedKey() string
	GetReGenKey(encPublicKey string, tag string) (string, error)
	ReEncrypt(encMsg *EncryptedMessage, reGenKey string, clientPublicKey string) (*ReEncryptedMessage, error)
	GetPublicKey() (string, error)
	GetPrivateKey() (string, error)
}

func NewEncryptionScheme() EncryptionScheme {
	return new(PREEncryptionScheme)
}

type EncryptedMessage struct {
	EncryptedKey    string
	EncryptedData   []byte
	ReEncryptionKey string
	MessageChecksum string
	OverallChecksum string
}
