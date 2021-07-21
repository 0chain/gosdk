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

func (w *withError) pprint() string {
	retError := ""
	if w.current != nil {
		switch c := w.current.(type) {
		case *Error:
			retError += c.pprint()
		case *withError:
			retError += c.pprint()
		default:
			retError += c.Error()
		}
	}
	if w.previous != nil {
		// retError += "\n" + w.previous.Error()
		switch p := w.previous.(type) {
		case *Error:
			retError += ": " + p.pprint()
		case *withError:
			retError += ": " + p.pprint()
		default:
			retError += ": " + p.Error()
		}
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
	return newWithLevel(3, args...)
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
		return invalidUsage()
	}

	return &currentError
}

func invalidUsage() *Error {
	return &Error{
		Code:     "incorrect_usage",
		Msg:      "you should at least pass message to properly wrap the current error!",
		Location: getErrorLocation(4),
	}
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

/*
Newf - creates a new error
*/
func Newf(code string, format string, args ...interface{}) *Error {
	return newWithLevel(3, code, fmt.Sprintf(format, args...))
}

func getErrorLocation(level int) string {
	_, file, line, _ := runtime.Caller(level)
	return fmt.Sprintf("%s:%d", file, line)
}

/* Is - tells whether actual error is targer error
where, actual error can be either Error/withError
if actual error is wrapped error then if any internal error
matches the target error then function results in true
*/
func Is(actual error, target *Error) bool {
	actualError := isError(actual)
	if actualError != nil {
		if actualError.Code == "" && target.Code == "" {
			return actualError.Msg == target.Msg
		} else {
			return actualError.Code == target.Code
		}

	} else {
		actualWithError := isWithError(actual)
		if actualWithError != nil {
			if actualWithError.current != nil {
				return Is(actualWithError.current, target)
			} else {
				return Is(actualWithError.previous, target)
			}
		} else {
			return false
		}
	}
}

// isError - parses the error into Error
func isError(err error) *Error {
	t, ok := err.(*Error)
	if ok {
		return t
	}
	return nil
}

// isError - parses the error into withError
func isWithError(err error) *withError {
	t, ok := err.(*withError)
	if ok {
		return t
	}
	return nil
}

// PPrint - pretty print
func PPrint(err error) string {
	if err == nil {
		return ""
	}
	switch e := err.(type) {
	case *Error:
		return e.pprint()
	case *withError:
		return e.pprint()
	default:
		return e.Error()
	}
}

func Register(args ...string) func() *Error {
	return func() *Error {
		return newWithLevel(3, args...)
	}
}

func Unmarshal(data string) error {
	errors := strings.Split(data, "\n")
	var finalError error

	for i := len(errors) - 1; i >= 0; i-- {
		e := errors[i]
		var err Error
		se := strings.Split(e, " ")

		switch len(se) {
		case 2:
			err.Location = se[0]
			err.Msg = se[1]
		case 3:
			err.Location = se[0]
			if isCode(se[1]) {
				err.Code = extractCode(se[1])
				err.Msg = se[2]
			} else {
				err.Msg = strings.Join(se[1:], " ")
			}
		default:
			continue
		}
		finalError = Wrap(finalError, &err)
	}
	return finalError
}

func isCode(code string) bool {
	// ascii code for ":" is 58
	return code[len(code)-1] == 58
}

func extractCode(code string) string {
	return strings.TrimRight(code, ":")
}
