package encryption

type EncryptionScheme interface {
	Initialize(mnemonic string) error
	InitForEncryption(tag string)
	InitForDecryption(tag string, encryptedKey string) error
	Encrypt(data []byte) (*EncryptedMessage, error)
	Decrypt(*EncryptedMessage) ([]byte, error)
	GetEncryptedKey() string
}

func NewEncryptionScheme() EncryptionScheme {
	return new(PREEncryptionScheme)
}

type EncryptedMessage struct {
	EncryptedKey    string
	EncryptedData   []byte
	MessageChecksum string
	OverallChecksum string
}
