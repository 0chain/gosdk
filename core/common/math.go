package common

import "math"

func MustAddInt(a, b int) int {

	if a > 0 && b > 0 {

		if math.MaxInt-a > b {
			panic("math: integer overflow")
		}

		return a + b
	}

	if a < 0 && b < 0 {
		if math.MaxInt-(-a) > -b {
			panic("math: integer underflow")
		}
	}

	return a + b
}
