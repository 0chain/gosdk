package errors

import (
	"reflect"
	"testing"
)

const (
	testCode = "test_code"
	testText = "test text"
	wrapCode = "wrap_code"
	wrapText = "wrap text"
)

func Test_errWrapper_Error(t *testing.T) {
	t.Parallel()

	tests := [1]struct {
		name string
		err  error
		want string
	}{
		{
			name: "OK",
			err:  Wrap(wrapCode, wrapText, New(testCode, testText)),
			want: wrapCode + delim + wrapText + delim + testCode + delim + testText,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := test.err.Error(); got != test.want {
				t.Errorf("Error() got: %v | want: %v", got, test.want)
			}
		})
	}
}

func Test_errWrapper_Unwrap(t *testing.T) {
	t.Parallel()

	err := New(testCode, testText)

	tests := [1]struct {
		name    string
		wrapper *errWrapper
		want    error
	}{
		{
			name:    "OK",
			wrapper: Wrap(wrapCode, wrapText, err),
			want:    err,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := test.wrapper.Unwrap(); !reflect.DeepEqual(got, test.want) {
				t.Errorf("Unwrap() got: %#v | want: %#v", got, test.want)
			}
		})
	}
}

func Test_errWrapper_Wrap(t *testing.T) {
	t.Parallel()

	err := New(testCode, testText)

	tests := [1]struct {
		name    string
		error   error
		wrapper *errWrapper
		want    *errWrapper
	}{
		{
			name:    "OK",
			error:   New(testCode, testText),
			wrapper: New(wrapCode, wrapText),
			want:    &errWrapper{code: wrapCode, text: wrapText + delim + err.Error(), wrap: err},
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := test.wrapper.Wrap(test.error); !reflect.DeepEqual(got, test.want) {
				t.Errorf("Wrap() got: %#v | want: %#v", got, test.want)
			}
		})
	}
}

func Test_errAny(t *testing.T) {
	t.Parallel()

	testErr := New(testCode, testText)
	wrapErr := Wrap(wrapCode, wrapText, testErr)

	tests := [2]struct {
		name    string
		list    []error
		wrapErr error
		want    bool
	}{
		{
			name:    "TRUE",
			list:    []error{testErr},
			wrapErr: wrapErr,
			want:    true,
		},
		{
			name:    "FALSE",
			list:    []error{testErr},
			wrapErr: Wrap(wrapCode, wrapText, New(testCode, testText)),
			want:    false,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := Any(test.wrapErr, test.list...); got != test.want {
				t.Errorf("errIs() got: %v | want: %v", got, test.want)
			}
		})
	}
}

func Test_errIs(t *testing.T) {
	t.Parallel()

	testErr := New(testCode, testText)
	wrapErr := Wrap(wrapCode, wrapText, testErr)

	tests := [2]struct {
		name    string
		testErr error
		wrapErr error
		want    bool
	}{
		{
			name:    "TRUE",
			testErr: testErr,
			wrapErr: wrapErr,
			want:    true,
		},
		{
			name:    "FALSE",
			testErr: testErr,
			wrapErr: Wrap(wrapCode, wrapText, New(testCode, testText)),
			want:    false,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := Is(test.wrapErr, test.testErr); got != test.want {
				t.Errorf("errIs() got: %v | want: %v", got, test.want)
			}
		})
	}
}

func Test_errNew(t *testing.T) {
	t.Parallel()

	tests := [1]struct {
		name string
		code string
		text string
		want *errWrapper
	}{
		{
			name: "Equal",
			code: testCode,
			text: testText,
			want: &errWrapper{code: testCode, text: testText},
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := New(test.code, test.text); !reflect.DeepEqual(got, test.want) {
				t.Errorf("errNew() got: %#v | want: %#v", got, test.want)
			}
		})
	}
}

func Test_errWrap(t *testing.T) {
	t.Parallel()

	tests := [2]struct {
		name string
		code string
		text string
		wrap error
		want string
	}{
		{
			name: "OK",
			code: wrapCode,
			text: wrapText,
			wrap: New(testCode, testText),
			want: wrapCode + delim + wrapText + delim + testCode + delim + testText,
		},
		{
			name: "nil_Wrap_OK",
			code: wrapCode,
			text: wrapText,
			wrap: nil,
			want: wrapCode + delim + wrapText,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := Wrap(test.code, test.text, test.wrap).Error(); got != test.want {
				t.Errorf("errWrap() got: %#v | want: %#v", got, test.want)
			}
		})
	}
}
