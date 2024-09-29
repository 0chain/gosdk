// DEPRECATED: This package is deprecated and will be removed in a future release.
package errors

import (
	"errors"
	"os"
)

const (
	delim = ": "
)

type (
	// Error type for a new application error.
	Error struct {
		Code string `json:"code,omitempty"`
		Msg  string `json:"msg"`
	}
)

type (
	// wrapper implements error wrapper interface.
	errWrapper struct {
		code string
		text string
		wrap error
	}
)

// Error implements error interface.
func (e *errWrapper) Error() string {
	return e.code + delim + e.text
}

// Unwrap implements error unwrap interface.
func (e *errWrapper) Unwrap() error {
	return e.wrap
}

// Wrap implements error wrapper interface.
func (e *errWrapper) Wrap(err error) *errWrapper {
	return Wrap(e.code, e.text, err)
}

// Any reports whether an error in error's chain
// matches to any error provided in list.
func Any(err error, targets ...error) bool {
	for _, target := range targets {
		if errors.Is(err, target) {
			return true
		}
	}

	return false
}

// ExitErr prints error to os.Stderr and call os.Exit with given code.
func ExitErr(text string, err error, code int) {
	text = Wrap("exit", text, err).Error()
	_, _ = os.Stderr.Write([]byte(text))
	os.Exit(code)
}

// ExitMsg prints message to os.Stderr and call os.Exit with given code.
func ExitMsg(text string, code int) {
	text = New("exit", text).Error()
	_, _ = os.Stderr.Write([]byte(text))
	os.Exit(code)
}

// Is wraps function errors.Is from stdlib to avoid import it
// in other places of the magma smart contract (magmasc) package.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// New returns constructed error wrapper interface.
func New(code, text string) *errWrapper {
	return &errWrapper{code: code, text: text}
}

// Wrap wraps given error into a new error with format.
func Wrap(code, text string, err error) *errWrapper {
	wrapper := &errWrapper{code: code, text: text}
	if err != nil && !errors.Is(wrapper, err) {
		wrapper.wrap = err
		wrapper.text += delim + err.Error()
	}

	return wrapper
}
