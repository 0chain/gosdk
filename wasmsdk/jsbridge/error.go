//go:build js && wasm
// +build js,wasm

package jsbridge

import "errors"

var (
	// ErrMismatchedInputLength the length of input are mismatched
	ErrMismatchedInputLength = errors.New("binder: mismatched input length")

	// ErrMismatchedOutputLength the length of output are mismatched
	ErrMismatchedOutputLength = errors.New("binder: mismatched output length")

	// ErrBinderNotImplemented the type binder is not implemented yet
	ErrBinderNotImplemented = errors.New("binder: not impelmented")

	// ErrFuncNotSupported the type function is not supported yet
	ErrFuncNotSupported = errors.New("func: not supported")

	// ErrIsNotFunc bind works with func only
	ErrIsNotFunc = errors.New("func: bind works with func only")
)
