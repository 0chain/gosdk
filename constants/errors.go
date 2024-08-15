// Package constants provides constants.
// The convention of naming is to use MixedCaps or mixedCaps rather than underscores to write multiword names. https://golang.org/doc/effective_go#mixed-caps
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

	// ErrNotLockedWritMarker failed to lock WriteMarker
	ErrNotLockedWritMarker = errors.New("failed to lock WriteMarker")

	// ErrNotUnlockedWritMarker failed to unlock WriteMarker
	ErrNotUnlockedWritMarker = errors.New("failed to unlock WriteMarker")

	// ErrInvalidHashnode invalid hashnode
	ErrInvalidHashnode = errors.New("invalid hashnode")

	// ErrBadRequest bad request
	ErrBadRequest = errors.New("bad request")

	// ErrNotFound ref not found
	ErrNotFound = errors.New("ref not found")

	// ErrFileOptionNotPermitted requested operation is not allowed on this allocation (file_options)
	ErrFileOptionNotPermitted = errors.New("this options for this file is not permitted for this allocation")
)
