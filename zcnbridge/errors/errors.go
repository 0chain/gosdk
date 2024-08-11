// Error struct and functions.
package errors

import (
	"errors"
	"fmt"
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
	// ErrWrapper implements error wrapper interface.
	ErrWrapper struct {
		code string
		text string
		wrap error
	}
)

// Error implements error interface.
func (e *ErrWrapper) Error() string {
	return e.code + delim + e.text
}

// Unwrap implements error unwrap interface.
func (e *ErrWrapper) Unwrap() error {
	return e.wrap
}

// Wrap implements error wrapper interface.
func (e *ErrWrapper) Wrap(err error) *ErrWrapper {
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

// Is checks if error is equal to target error.
// 		- err: error to check
// 		- target: target error to compare
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// New returns constructed error wrapper interface.
func New(code, text string) *ErrWrapper {
	return &ErrWrapper{code: code, text: text}
}

// Wrap wraps given error into a new error with format.
func Wrap(code, text string, err error) *ErrWrapper {
	wrapper := &ErrWrapper{code: code, text: text}
	if err != nil && !errors.Is(wrapper, err) {
		wrapper.wrap = err
		wrapper.text += delim + err.Error()
	}

	return wrapper
}

func (err *Error) Error() string {
	return fmt.Sprintf("%s: %s", err.Code, err.Msg)
}

// NewError create a new error instance given a code and a message.
//   - code: error code
//   - msg: error message
func NewError(code string, msg string) *Error {
	return &Error{Code: code, Msg: msg}
}

/*NewErrorf - create a new error with format */
func NewErrorf(code string, format string, args ...interface{}) *Error {
	return &Error{Code: code, Msg: fmt.Sprintf(format, args...)}
}
