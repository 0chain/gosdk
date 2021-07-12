package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type newErrorTestCase struct {
	about           string
	args            []string
	expectedCode    string
	expectedMessage string
}

func getNewErrorTestCase() []newErrorTestCase {
	return []newErrorTestCase{
		{
			about:           "creating an error with code and msg.",
			args:            []string{"500", "This is a very big error! Beware of it!"},
			expectedCode:    "500",
			expectedMessage: "This is a very big error! Beware of it!",
		},
		{
			about:           "creating an error with empty code and msg.",
			args:            []string{"", "This is a very big error! Beware of it!"},
			expectedCode:    "",
			expectedMessage: "This is a very big error! Beware of it!",
		},
		{
			about:           "creating an error with code and empty msg.",
			args:            []string{"401", ""},
			expectedCode:    "401",
			expectedMessage: "",
		},
		{
			about:           "creating an error with just msg.",
			args:            []string{"This is a short error!"},
			expectedCode:    "",
			expectedMessage: "This is a short error!",
		},
		{
			about:           "creating an error with not allowed parameters",
			args:            []string{"code", "message", "third"},
			expectedCode:    "incorrect_usage",
			expectedMessage: "you should at least pass message to create a proper error!",
		},
		{
			about:           "creating an error with not allowed parameters",
			args:            []string{"code", "message", "third", "fourth"},
			expectedCode:    "incorrect_usage",
			expectedMessage: "you should at least pass message to create a proper error!",
		},
		{
			about:           "creating an error with empty parameters",
			args:            []string{},
			expectedCode:    "incorrect_usage",
			expectedMessage: "you should at least pass message to create a proper error!",
		},
	}
}

func TestNew(t *testing.T) {
	for _, tc := range getNewErrorTestCase() {
		t.Run(tc.about, func(t *testing.T) {
			err := New(tc.args...)

			require.Equal(t, tc.expectedCode, err.Code)
			require.Equal(t, tc.expectedMessage, err.Msg)
		})
	}
}

func TestError(t *testing.T) {
	for _, tc := range getNewErrorTestCase() {
		t.Run(tc.about, func(t *testing.T) {
			err := New(tc.args...)

			require.Contains(t, err.Error(), tc.expectedMessage)
		})
	}
}

type wrapTopTestCase struct {
	about              string
	testCase           []interface{}
	expectedTopMessage string
}

func getWrapTopTestCases() []wrapTopTestCase {
	return []wrapTopTestCase{
		{
			about: "wrapping all errors",
			testCase: []interface{}{
				New("500", "This is a very big error! Beware of it!"),
				New("", "This is a very big error! Beware of it!"),
				New("401", ""),
				New("This is a short error!"),
				New("code", "message", "third"),
				New("code", "message", "third", "fourth"),
				New(),
				errors.New("error created from err package"),
				fmt.Errorf("%s", "error created from fmt package"),
				nil,
			},
			expectedTopMessage: "incorrect_usage: you should at least pass message to properly wrap the current error!",
		},
		{
			about: "wrapping all messages",
			testCase: []interface{}{
				"This is a very \"big\" error! Beware of it!",
				"This is a very 'big' error! Beware of it!",
				"This is a short error!",
				"",
			},
			expectedTopMessage: "incorrect_usage: you should at least pass message to properly wrap the current error!",
		},
		{
			about: "wrapping errors and messages",
			testCase: []interface{}{
				New("500", "This is a very big error! Beware of it!"),
				"This is a very \"big\" error! Beware of it!",
				New("401", ""),
				"This is a very 'big' error! Beware of it!",
				New("This is a short error!"),
				"",
				nil,
				New("code", "message", "third"),
				"This is a short error!",
				New("code", "message", "third", "fourth"),
				New(),
				New("", "This is a very big error! Beware of it!"),
			},
			expectedTopMessage: "This is a very big error! Beware of it!",
		},
	}
}

func TestWrap(t *testing.T) {
	for _, gtc := range getWrapTopTestCases() {
		t.Run(gtc.about, func(t *testing.T) {
			var wrappedError error
			for _, tc := range gtc.testCase {
				wrappedError = Wrap(wrappedError, tc)
			}
			require.Equal(t, len(gtc.testCase), len(strings.Split(wrappedError.Error(), "\n")))
		})
	}
}

func TestTop(t *testing.T) {
	for _, gtc := range getWrapTopTestCases() {
		t.Run(gtc.about, func(t *testing.T) {
			var wrappedError error
			for _, tc := range gtc.testCase {
				wrappedError = Wrap(wrappedError, tc)
			}
			require.Equal(t, gtc.expectedTopMessage, Top(wrappedError))
		})
	}
}

type newErrorfTestCase struct {
	about           string
	code            string
	format          string
	args            []interface{}
	expectedCode    string
	expectedMessage string
}

func getNewErrorfTestCase() []newErrorfTestCase {
	return []newErrorfTestCase{
		{
			about:           "creating an error with code and msg with integer arg.",
			code:            "integer_error",
			format:          "This error has a integer: %d",
			args:            []interface{}{500},
			expectedCode:    "integer_error",
			expectedMessage: "This error has a integer: 500",
		},
		{
			about:           "creating an error with code and msg with string arg.",
			code:            "string_error",
			format:          "This error has a string: %s",
			args:            []interface{}{"500"},
			expectedCode:    "string_error",
			expectedMessage: "This error has a string: 500",
		},
		{
			about:           "creating an error with empty code and empty msg with string arg.",
			code:            "",
			format:          "This error has empty code with a string: %s",
			args:            []interface{}{"401"},
			expectedCode:    "",
			expectedMessage: "This error has empty code with a string: 401",
		},
		{
			about:           "creating an error with just msg.",
			code:            "",
			format:          "This is a short error!",
			args:            []interface{}{},
			expectedCode:    "",
			expectedMessage: "This is a short error!",
		},
		// {
		// 	about:           "creating an error with 0args and args are expected",
		// 	code:            "incorrect_usage",
		// 	format:          "This is a short error! %d %s",
		// 	args:            []interface{}{},
		// 	expectedCode:    "incorrect_usage",
		// 	expectedMessage: "This is a short error!",
		// },
		// {
		// 	about:           "creating an error with not allowed parameters",
		// 	args:            []string{"code", "message", "third", "fourth"},
		// 	expectedCode:    "incorrect_usage",
		// 	expectedMessage: "you should at least pass message to create a proper error!",
		// },
		// {
		// 	about:           "creating an error with empty parameters",
		// 	args:            []string{},
		// 	expectedCode:    "incorrect_usage",
		// 	expectedMessage: "you should at least pass message to create a proper error!",
		// },
	}
}

func TestNewf(t *testing.T) {
	for _, tc := range getNewErrorfTestCase() {
		t.Run(tc.about, func(t *testing.T) {
			err := Newf(tc.code, tc.format, tc.args...)

			require.Equal(t, tc.expectedCode, err.Code)
			require.Equal(t, tc.expectedMessage, err.Msg)
		})
	}
}
