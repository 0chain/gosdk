package zboxutil

import "math/bits"

type Uint128 struct {
	high uint64
	low  uint64
}

func NewUint128(x uint64) Uint128 {
	return Uint128{low: x}
}

// Add returns x+y.
func (x Uint128) Add(y Uint128) Uint128 {
	low, carry := bits.Add64(x.low, y.low, 0)
	high, carry := bits.Add64(x.high, y.high, carry)
	if carry != 0 {
		panic("overflow")
	}
	return Uint128{high, low}
}

// Add64 returns x+y.
func (x Uint128) Add64(y uint64) Uint128 {
	low, carry := bits.Add64(x.low, y, 0)
	high, carry := bits.Add64(x.high, 0, carry)
	if carry != 0 {
		panic("overflow")
	}
	return Uint128{high, low}
}

// Sub returns x-y.
func (x Uint128) Sub(y Uint128) Uint128 {
	low, borrow := bits.Sub64(x.low, y.low, 0)
	high, _ := bits.Sub64(x.high, y.high, borrow)
	return Uint128{high, low}
}

// Sub64 returns x-y.
func (x Uint128) Sub64(y uint64) Uint128 {
	low, borrow := bits.Sub64(x.low, y, 0)
	high, _ := bits.Sub64(x.high, 0, borrow)
	return Uint128{high, low}
}

// Equals returns true if x == y.
func (x Uint128) Equals(y Uint128) bool {
	return x == y
}

// Equals64 returns true if x == y.
func (x Uint128) Equals64(y uint64) bool {
	return x.low == y && x.high == 0
}

// And returns x&y.
func (x Uint128) And(v Uint128) Uint128 {
	return Uint128{x.high & v.high, x.low & v.low}
}

// And64 returns x&y.
func (x Uint128) And64(y uint64) Uint128 {
	return Uint128{low: x.low & y}
}

// Lsh returns x<<y.
func (x Uint128) Lsh(y uint64) Uint128 {
	z := Uint128{}
	if y > 64 {
		z.low = 0
		z.high = x.low << (y - 64)
	} else {
		z.low = x.low << y
		z.high = x.high<<y | x.low>>(64-y)
	}

	return z
}

// Not returns ^x.
func (x Uint128) Not() Uint128 {
	return Uint128{^x.high, ^x.low}
}

// Or returns x|y.
func (x Uint128) Or(v Uint128) Uint128 {
	return Uint128{x.high | v.high, x.low | v.low}
}

// Xor returns x^y.
func (x Uint128) Xor(v Uint128) Uint128 {
	return Uint128{x.high ^ v.high, x.low ^ v.low}
}

// CountOnes return num of 1 bits in x.
func (x Uint128) CountOnes() int {
	return bits.OnesCount64(x.low) + bits.OnesCount64(x.high)
}

// TrailingZeros returns the number of trailing zero bits in x; the result is 128 for x == 0.
func (x Uint128) TrailingZeros() int {
	if x.low > 0 {
		return bits.TrailingZeros64(x.low)
	}
	return 64 + bits.TrailingZeros64(x.high)
}
