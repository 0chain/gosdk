package common

import (
	"errors"
	"math"
)

func TryAddInt(a, b int) (int, error) {
	if a > 0 && b > 0 {
		if math.MaxInt-a < b {
			return 0, errors.New("math: integer overflow")
		}

	}

	if a < 0 && b < 0 {
		if math.MinInt-a > b {
			return 0, errors.New("math: integer underflow")
		}

	}

	return a + b, nil
}

func MustAddInt(a, b int) int {
	i, err := TryAddInt(a, b)
	if err != nil {
		panic(err.Error())
	}

	return i
}
