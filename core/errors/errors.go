package gosdkError

import "github.com/0chain/errors"

var (
	ErrTooManyRequests      = errors.New(TooManyRequests, "")
	ErrFileNotFound         = errors.New(FileNotFound, "")
	ErrInvalidReferencePath = errors.New(InvalidReferencePath, "")
	ErrInvalidValue         = errors.New(InvalidValue, "")
	ErrMarshall             = errors.New(MarshallError, "")
	ErrUnmarshall           = errors.New(UnmarshallError, "")
	ErrDecode               = errors.New(DecodeError, "")
	// Erasure Coding Errors
	ErrEC                   = errors.New(ECError, "")
	ErrECSplit              = errors.New(ECSplitError, "")
	ErrECVerify             = errors.New(ECVerifyError, "")
	ErrECReconstruct        = errors.New(ECReconstructError, "")
	ErrECInvalidInputLength = errors.New(ECInvalidInputLength, "")
	ErrECJoin               = errors.New(ECJoinError, "")
	// Encryption/decryption errors
	ErrDecrypt                = errors.New(DecryptError, "")
	ErrEncrypt                = errors.New(EncryptError, "")
	ErrReEncrypt              = errors.New(ReEncryptError, "")
	ErrReDecrypt              = errors.New(ReDecryptError, "")
	ErrInvalidCipherText      = errors.New(InvalidCipherText, "")
	ErrSymmetricDecryption    = errors.New(SymmetricDecryptionError, "")
	ErrSymmetricEncryption    = errors.New(SymmetricEncryptionError, "")
	ErrInvalidHeaderChecksums = errors.New(InvalidHeaderChecksums, "")

	ErrInvalidListPath = errors.New(InvalidListPath, "")
	ErrInvalidRefPath  = errors.New(InvalidRefPath, "")
)
