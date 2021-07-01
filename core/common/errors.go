package common

import (
	"fmt"
	"runtime"
)

/*Error type for a new application error */
type Error struct {
	Code     string `json:"code,omitempty"`
	Msg      string `json:"msg"`
	Location string `json:"location"`
}

func (err *Error) Error() string {
	if err.Code == "" {
		return fmt.Sprintf("%s %s", err.Location, err.Msg)
	}
	return fmt.Sprintf("%s %s: %s", err.Location, err.Code, err.Msg)
}

type withError struct {
	previous error
	current  error
}

func (w *withError) Error() string {
	return w.current.Error() + "\n" + w.previous.Error()
}

// WrapWithError wrap the previous error with current error
func WrapWithError(previous error, current error) error {
	if current == nil {
		return previous
	}

	return &withError{
		previous: previous,
		current:  current,
	}
}

// WrapWithError wrap the previous error with current error message
func WrapWithMessage(previous error, msg string) error {
	return &withError{
		previous: previous,
		current: &Error{
			Msg:      msg,
			Location: getErrorLocation(),
		},
	}
}

/*NewError - create a new error */
func NewError(code string, msg string) *Error {
	return &Error{
		Code:     code,
		Msg:      msg,
		Location: getErrorLocation(),
	}
}

func getErrorLocation() string {
	_, file, line, _ := runtime.Caller(2)
	return fmt.Sprintf("%s:%d", file, line)
}

/*InvalidRequest - create error messages that are needed when validating request input */
func InvalidRequest(msg string) error {
	return NewError("invalid_request", fmt.Sprintf("Invalid request (%v)", msg))
}
