// Package constants provides constants.the convention of naming is to use MixedCaps or mixedCaps rather than underscores to write multiword names. https://golang.org/doc/effective_go#mixed-caps
package constants

import "errors"

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
