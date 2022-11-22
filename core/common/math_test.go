package common

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMathAddInt(t *testing.T) {
	a := math.MaxInt
	b := 1

	_, err := TryAddInt(a, b)
	require.NotNil(t, err)

	tests := []struct {
		name    string
		a       int
		b       int
		result  int
		wantErr bool
	}{
		{
			name:    "greater than MaxInt must be overflow",
			a:       math.MaxInt,
			b:       1,
			wantErr: true,
		},
		{
			name:   "equal to MaxInt must not be overflow",
			a:      math.MaxInt - 1,
			b:      1,
			result: math.MaxInt,
		},
		{
			name:   "less than MaxInt must not be overflow",
			a:      math.MaxInt - 2,
			b:      1,
			result: math.MaxInt - 1,
		},
		{
			name:   "greater than MinInt must not be underflow",
			a:      math.MinInt,
			b:      1,
			result: math.MinInt + 1,
		},
		{
			name:   "equal to MinInt should not be underflow",
			a:      math.MinInt + 1,
			b:      -1,
			result: math.MinInt,
		},
		{
			name:    "less than MinInt must be underflow",
			a:       math.MinInt,
			b:       -1,
			wantErr: true,
		},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {

			i, err := TryAddInt(test.a, test.b)
			if test.wantErr {
				require.NotNil(t, err)
			}

			require.Equal(t, test.result, i)

		})

	}

}
