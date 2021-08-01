package errors

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

/*Error type for a new application error */
type Error struct {
	Code string `json:"code,omitempty"`
	Msg  string `json:"msg"`
}

func (err *Error) Error() string {
	if err.Code == "" {
		return err.Msg
	}
	return fmt.Sprintf("%s: %s", err.Code, err.Msg)
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
	currentError := Error{}

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

func Newf(code string, format string, args ...interface{}) *Error {
	return New(code, fmt.Sprintf(format, args...))
}

func Is(err error, target *Error) bool {
	if err == nil {
		return false
	}
	actualError := isError(err)
	if actualError != nil {
		if actualError.Code == "" && target.Code == "" {
			return actualError.Msg == target.Msg
		} else {
			return actualError.Code == target.Code
		}
	} else {
		return is(err, target)
	}
}

func is(err error, target *Error) bool {
	return strings.Contains(strings.Trim(strings.Split(err.Error(), " ")[0], ":"), target.Code)
}

func isError(err error) *Error {
	t, ok := err.(*Error)
	if ok {
		return t
	}
	return nil
}

func Wrap(err error, message string) error {
	if err == nil {
		err = errors.New("")
	}
	return errors.Wrap(err, message)
}
