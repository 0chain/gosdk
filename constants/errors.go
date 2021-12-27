// Package constants provides constants.the convention of naming is to use MixedCaps or mixedCaps rather than underscores to write multiword names. https://golang.org/doc/effective_go#mixed-caps
package constants

import (
	"errors"

	zerrors "github.com/0chain/errors"
)

var (
	// ErrInvalidParameter parameter is not specified or invalid
	ErrInvalidParameter = errors.New("invalid parameter")

	// ErrUnableHash failed to hash with unknown exception
	ErrUnableHash = errors.New("unable to hash")

	// ErrUnableWriteFile failed to write bytes to file
	ErrUnableWriteFile = errors.New("unable to write file")

	// ErrNotImplemented feature/method is not implemented yet
	ErrNotImplemented = errors.New("Not Implemented")
)

// zerrors code
const (
	UnknownCodeError             = ""
	TooManyRequestsError         = "too_many_requests"
	ParseError                   = "parse_error"
	ValidateError                = "validate_error"
	TransactionVerificationError = "verify_transaction"
	TransactionSendError         = "transaction_send_error"
	LeafExistError               = "leaf_exist"
	RecoverKeysError             = "recover_keys"
	WasmNotSupportedError        = "wasm_not_supported"
	WasmNotInitializedError      = "wasm_not_initialized"
	KeysGenerationError          = "generate_keys"
	WalletMarshalError           = "wallet_marshal"
	InvalidReferencePathError    = "invalid_reference_path"
	InvalidListPathError         = "invalid_list_path"
	FileNotFoundError            = "file_not_found"
	InvalidInputLengthError      = "invalid_input_length"
	InvalidRefPathError          = "invalid_ref_path"
	WriteMarkerValidationError   = "write_marker_validation_failed"
	SdkNotInitializedError       = "sdk_not_initialized"
	InvalidPathError             = "invalid_path"
	InvalidNameError             = "invalid_name"
	AuthTicketDecodeError        = "auth_ticket_decode_error"
	ConsensusNotReachedError     = "consensus_not_reached"
	ConsensusFailedError         = "consensus_failed"
)

var (
	ErrUnknown                 = zerrors.New(UnknownCodeError, "")
	ErrTooManyRequests         = zerrors.New(TooManyRequestsError, "")
	ErrParse                   = zerrors.New(ParseError, "")
	ErrValidate                = zerrors.New(ValidateError, "")
	ErrTransactionVerification = zerrors.New(TransactionVerificationError, "")
	ErrTransactionSend         = zerrors.New(TransactionSendError, "")
	ErrLeafExist               = zerrors.New(LeafExistError, "")
	ErrRecoverKeys             = zerrors.New(RecoverKeysError, "")
	ErrWasmNotSupported        = zerrors.New(WasmNotSupportedError, "")
	ErrWasmNotInitialized      = zerrors.New(WasmNotInitializedError, "")
	ErrKeysGeneration          = zerrors.New(KeysGenerationError, "")
	ErrWalletMarshal           = zerrors.New(WalletMarshalError, "")
	ErrInvalidReferencePath    = zerrors.New(InvalidReferencePathError, "")
	ErrInvalidListPath         = zerrors.New(InvalidListPathError, "")
	ErrFileNotFound            = zerrors.New(FileNotFoundError, "")
	ErrInvalidInputLength      = zerrors.New(InvalidInputLengthError, "")
	ErrInvalidRefPath          = zerrors.New(InvalidRefPathError, "")
	ErrWriteMarkerValidation   = zerrors.New(WriteMarkerValidationError, "")
	ErrSdkNotInitialized       = zerrors.New(SdkNotInitializedError, "")
	ErrInvalidPath             = zerrors.New(InvalidPathError, "")
	ErrInvalidName             = zerrors.New(InvalidNameError, "")
	ErrAuthTicketDecode        = zerrors.New(AuthTicketDecodeError, "")
	ErrConsensusNotReached     = zerrors.New(ConsensusNotReachedError, "")
	ErrConsensusFailed         = zerrors.New(ConsensusFailedError, "")
)
