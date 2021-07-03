package errors

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

func (err *Error) topLevelError() string {
	if err.Code == "" {
		return err.Msg
	}
	return fmt.Sprintf("%s: %s", err.Code, err.Msg)
}

// TopLevelError since errors can be wrapped and stacked,
// it's necessary to get the top level error for tests and validations
func TopLevelError(err error) string {
	switch t := err.(type) {
	case *Error:
		return t.topLevelError()
	case *withError:
		switch ct := t.current.(type) {
		case *Error:
			return ct.topLevelError()
		default:
			return err.Error()
		}
	default:
		return err.Error()
	}
}

type withError struct {
	previous error
	current  error
}

func (w *withError) Error() string {
	var retError string
	if w.current != nil {
		retError = w.current.Error()
	}
	if w.previous != nil {
		retError += "\n" + w.previous.Error()
	}
	return retError
}

// WrapError wrap the previous error with current error/ message
func WrapError(previous error, current interface{}) error {
	if current == nil {
		return previous
	}

	switch c := current.(type) {
	case error:
		return &withError{
			previous: previous,
			current:  c,
		}
	case string:
		return &withError{
			previous: previous,
			current: &Error{
				Msg:      c,
				Location: getErrorLocation(2),
			},
		}
	default:
		return &withError{
			previous: previous,
			current: &Error{
				Code:     "incorrect_usage",
				Msg:      "you should at least pass message to properly wrap the current error!",
				Location: getErrorLocation(1),
			},
		}
	}
}

/*
NewError - create a new error 
two arguments can be passed! 
1. code 
2. message
if only one argument is passed its considered as message
if two arguments are passed then 
	first argument is considered for code and 
	second argument is considered for message
*/
func NewError(args ...string) *Error {
	switch len(args) {
	case 1:
		return &Error{
			Code:     "",
			Msg:      args[0],
			Location: getErrorLocation(2),
		}
	case 2:
		return &Error{
			Code:     args[0],
			Msg:      args[1],
			Location: getErrorLocation(2),
		}
	default:
		return &Error{
			Code:     "incorrect_usage",
			Msg:      "you should at least pass message to create a proper error!",
			Location: getErrorLocation(1),
		}
	}
}

func getErrorLocation(level int) string {
	_, file, line, _ := runtime.Caller(level)
	return fmt.Sprintf("%s:%d", file, line)
}

/*InvalidRequest - create error messages that are needed when validating request input */
func InvalidRequest(msg string) error {
	return NewError("invalid_request", fmt.Sprintf("Invalid request (%v)", msg))
}
