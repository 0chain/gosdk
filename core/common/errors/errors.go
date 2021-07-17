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

func (err *Error) top() string {
	if err.Code == "" {
		return err.Msg
	}
	return fmt.Sprintf("%s: %s", err.Code, err.Msg)
}

// Top since errors can be wrapped and stacked,
// it's necessary to get the top level error for tests and validations
func Top(err error) string {
	switch t := err.(type) {
	case *Error:
		return t.top()
	case *withError:
		switch ct := t.current.(type) {
		case *Error:
			return ct.top()
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

func invalidWrap() error {
	return &Error{
		Code:     "incorrect_usage",
		Msg:      "you should at least pass message to properly wrap the current error!",
		Location: getErrorLocation(3),
	}
}

// Wrap wrap the previous error with current error/ message
func Wrap(previous error, current interface{}) error {

	var currentError error
	switch c := current.(type) {
	case error:
		if c == nil {
			currentError = invalidWrap()
		} else {
			currentError = c
		}
	case string:
		if c == "" {
			currentError = invalidWrap()
		} else {
			currentError = &Error{
				Msg:      c,
				Location: getErrorLocation(2),
			}
		}
	default:
		currentError = invalidWrap()
	}

	return &withError{
		previous: previous,
		current:  currentError,
	}
}

/*
New - create a new error
two arguments can be passed!
1. code
2. message
if only one argument is passed its considered as message
if two arguments are passed then
	first argument is considered for code and
	second argument is considered for message
*/
func New(args ...string) *Error {
	currentError := Error{
		Location: getErrorLocation(2),
	}

	switch len(args) {
	case 1:
		currentError.Msg = args[0]
	case 2:
		currentError.Code = args[0]
		currentError.Msg = args[1]
	default:
		currentError.Code = "incorrect_usage"
		currentError.Msg = "you should at least pass message to create a proper error!"
	}

	return &currentError
}

func getErrorLocation(level int) string {
	_, file, line, _ := runtime.Caller(level)
	return fmt.Sprintf("%s:%d", file, line)
}

// /*InvalidRequest - create error messages that are needed when validating request input */
// func InvalidRequest(msg string) error {
// 	return New("invalid_request", fmt.Sprintf("Invalid request (%v)", msg))
// }
