package errors

import (
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
