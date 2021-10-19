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
	ErrNotImplemented = errors.New("not implemented")

	// ErrInvalidOperation failed to invoke a method
	ErrInvalidOperation = errors.New("invalid operation")

	// ErrBadRequest bad request
	ErrBadRequest = errors.New("bad request")

	// ErrUnknown unknown exception
	ErrUnknown = errors.New("unknown")

	// ErrBadDatabaseOperation unknown exception for db
	ErrBadDatabaseOperation = errors.New("bad db")

	// ErrInternal an unknown internal server error
	ErrInternal = errors.New("internal")

	// ErrEntityNotFound entity can't found in db
	ErrEntityNotFound = errors.New("entity not found")
)
