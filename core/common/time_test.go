package common

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseTime(t *testing.T) {

	now := time.Date(2022, 3, 22, 0, 0, 0, 0, time.Now().UTC().Location())
	null := time.Now()

	tests := []struct {
		name         string
		input        string
		exceptedTime time.Time
		exceptedErr  error
	}{
		{
			name:         "duration",
			input:        "+1h30m",
			exceptedTime: now.Add(90 * time.Minute),
		},
		{
			name:         "seconds",
			input:        "+30",
			exceptedTime: now.Add(30 * time.Second),
		},
		{
			name:         "Unix timestamp",
			input:        "1647858200",
			exceptedTime: time.Unix(1647858200, 0),
		},
		{
			name:         "YYYY-MM-dd HH:mm:ss",
			input:        "1647858200",
			exceptedTime: time.Unix(1647858200, 0),
		},
		{
			name:         "empty input",
			input:        "",
			exceptedTime: null,
			exceptedErr:  ErrInvalidTime,
		},
		{
			name:         "invaid input",
			input:        "s",
			exceptedTime: null,
			exceptedErr:  ErrInvalidTime,
		},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {

			v, err := ParseTime(now, test.input)

			if test.exceptedTime.Equal(null) {
				require.Nil(t, v)
			} else {
				require.NotNil(t, v)
				require.True(t, v.Equal(test.exceptedTime))
			}

			if test.exceptedErr == nil {
				require.Nil(t, err)
			} else {
				require.True(t, errors.Is(err, test.exceptedErr))
			}
		})
	}

}
