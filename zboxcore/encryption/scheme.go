package encryption

type EncryptionScheme interface {
	Initialize(mnemonic string) ([]byte, error)
	InitializeWithPrivateKey(privateKey []byte) error
	InitForEncryption(tag string)
	InitForDecryption(tag string, encryptedKey string) error
	Encrypt(data []byte) (*EncryptedMessage, error)
	Decrypt(*EncryptedMessage) ([]byte, error)
	GetEncryptedKey() string
	GetReGenKey(encPublicKey string, tag string) (string, error)
	GetPublicKey() (string, error)
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
