// Package constants provides constants
// naming  is to use MixedCaps or mixedCaps rather than underscores to write multiword names. https://golang.org/doc/effective_go#mixed-caps
package constants

import "errors"

var (
	// ErrUnableHash failed to hash with unknown exception
	ErrUnableHash = errors.New("unable to hash")

	// ErrUnableWriteFile failed to write bytes to file
	ErrUnableWriteFile = errors.New("unable to write file")

	// ErrNotImplemented feature/method is not implemented yet
	ErrNotImplemented = errors.New("Not Implemented")
)
