// +build js,wasm

package jsbridge

import "errors"

var (
	// ErrMismatchedInputLength the length of input are mismatched
	ErrMismatchedInputLength = errors.New("binder: mismatched input length")

	// ErrBinderNotImplemented the type binder is not implemented yet
	ErrBinderNotImplemented = errors.New("binder: not impelmented")
)
