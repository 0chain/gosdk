package errors

import (
	"fmt"
	"runtime"
	"strings"
)

/*Error type for a new application error */
type Error struct {
	Code     string `json:"code,omitempty"`
	Msg      string `json:"msg"`
	Location string `json:"location"`
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
	return newWithLevel(3, args...)
}

/*
Newf - creates a new error
*/
func Newf(code string, format string, args ...interface{}) *Error {
	return newWithLevel(3, code, fmt.Sprintf(format, args...))
}

func Register(args ...string) func() *Error {
	return func() *Error {
		return newWithLevel(3, args...)
	}
}

func (err *Error) Error() string {
	if strings.TrimSpace(err.Code) == "" {
		return fmt.Sprintf("%s %s", err.Location, strings.TrimSpace(err.Msg))
	}
	return fmt.Sprintf("%s %s: %s", err.Location, strings.TrimSpace(err.Code), strings.TrimSpace(err.Msg))
}

func (err *Error) top() string {
	if strings.TrimSpace(err.Code) == "" {
		return err.Msg
	}
	return fmt.Sprintf("%s: %s", strings.TrimSpace(err.Code), strings.TrimSpace(err.Msg))
}

func (err *Error) pprint() string {
	if strings.TrimSpace(err.Code) == "" {
		return err.Msg
	}
	return fmt.Sprintf("%s: %s", strings.TrimSpace(err.Code), strings.TrimSpace(err.Msg))
}

func newWithLevel(level int, args ...string) *Error {
	currentError := Error{
		Location: getErrorLocation(level),
	}

	switch len(args) {
	case 1:
		currentError.Msg = args[0]
	case 2:
		if isInvalidCode(args[0]) {
			return invalidCode(args[0])
		}
		currentError.Code, currentError.Msg = args[0], args[1]
	default:
		return invalidUsage(args...)
	}

	return &currentError
}

func getErrorLocation(level int) string {
	_, file, line, _ := runtime.Caller(level)
	return fmt.Sprintf("%s:%d", file, line)
}

func invalidUsage(args ...string) *Error {
	return &Error{
		Code:     "incorrect_usage",
		Msg:      fmt.Sprintf("max allowed parameters is 2 i.e code, msg. parameters sent - %d", len(args)),
		Location: getErrorLocation(4),
	}
}

func isCode(code string) bool {
	// ascii code for ":" is 58
	return code[len(code)-1] == 58
}

func invalidCode(code string) *Error {
	return &Error{
		Code:     "incorrect_code",
		Msg:      "code should not have spaces. use '" + strings.ToLower(strings.ReplaceAll(code, " ", "_")) + "' instead of '" + code + "'",
		Location: getErrorLocation(4),
	}
}

func isInvalidCode(code string) bool {
	return len(strings.Split(code, " ")) != 1
}
