package gosdkError

const (
	TooManyRequests        = "too_many_requests"
	AttrChangeProcessError = "attributes_change_process"
	FileNotFound           = "file_not_found"
	InvalidReferencePath   = "invalid_reference_path"
	MarshallError          = "marshall_error"
	UnmarshallError        = "unmarshall_error"
	DecodeError            = "decode_error"
	// Erasure Coding errors
	ECError              = "ec_error"
	ECSplitError         = "ec_split_error"
	ECVerifyError        = "ec_verify_error"
	ECReconstructError   = "ec_reconstruct_error"
	ECEncodingError      = "ec_encoding_error"
	ECInvalidInputLength = "ec_invalid_input_length"
	ECJoinError          = "ec_join_error"

	// Encryption/Decryption
	EncryptError             = "encrypt_error"
	DecryptError             = "decrypt_error"
	ReEncryptError           = "reencrypt_error"
	ReDecryptError           = "redecrypt_error"
	InvalidCipherText        = "invalid_cipher_text"
	SymmetricDecryptionError = "sym_decryption_error"
	SymmetricEncryptionError = "sym_encryption_error"
	InvalidHeaderChecksums   = "invalid_header_checksums"
)
