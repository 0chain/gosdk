package zboxutil

import "math/bits"

type Uint128 struct {
	High uint64
	Low  uint64
}

func NewUint128(x uint64) Uint128 {
	return Uint128{Low: x}
}

// Add returns x+y.
func (x Uint128) Add(y Uint128) Uint128 {
	Low, carry := bits.Add64(x.Low, y.Low, 0)
	High, carry := bits.Add64(x.High, y.High, carry)
	if carry != 0 {
		panic("overfLow")
	}
	return Uint128{High, Low}
}

// Add64 returns x+y.
func (x Uint128) Add64(y uint64) Uint128 {
	Low, carry := bits.Add64(x.Low, y, 0)
	High, carry := bits.Add64(x.High, 0, carry)
	if carry != 0 {
		panic("overfLow")
	}
	return Uint128{High, Low}
}

// Sub returns x-y.
func (x Uint128) Sub(y Uint128) Uint128 {
	Low, borrow := bits.Sub64(x.Low, y.Low, 0)
	High, _ := bits.Sub64(x.High, y.High, borrow)
	return Uint128{High, Low}
}

// Sub64 returns x-y.
func (x Uint128) Sub64(y uint64) Uint128 {
	Low, borrow := bits.Sub64(x.Low, y, 0)
	High, _ := bits.Sub64(x.High, 0, borrow)
	return Uint128{High, Low}
}

// Equals returns true if x == y.
func (x Uint128) Equals(y Uint128) bool {
	return x == y
}

// Equals64 returns true if x == y.
func (x Uint128) Equals64(y uint64) bool {
	return x.Low == y && x.High == 0
}

// And returns x&y.
func (x Uint128) And(v Uint128) Uint128 {
	return Uint128{x.High & v.High, x.Low & v.Low}
}

// And64 returns x&y.
func (x Uint128) And64(y uint64) Uint128 {
	return Uint128{Low: x.Low & y}
}

// Lsh returns x<<y.
func (x Uint128) Lsh(y uint64) Uint128 {
	z := Uint128{}
	if y > 64 {
		z.Low = 0
		z.High = x.Low << (y - 64)
	} else {
		z.Low = x.Low << y
		z.High = x.High<<y | x.Low>>(64-y)
	}

	return z
}

// Not returns ^x.
func (x Uint128) Not() Uint128 {
	return Uint128{^x.High, ^x.Low}
}

// Or returns x|y.
func (x Uint128) Or(v Uint128) Uint128 {
	return Uint128{x.High | v.High, x.Low | v.Low}
}

// Xor returns x^y.
func (x Uint128) Xor(v Uint128) Uint128 {
	return Uint128{x.High ^ v.High, x.Low ^ v.Low}
}

// CountOnes return num of 1 bits in x.
func (x Uint128) CountOnes() int {
	return bits.OnesCount64(x.Low) + bits.OnesCount64(x.High)
}

// TrailingZeros returns the number of trailing zero bits in x; the result is 128 for x == 0.
func (x Uint128) TrailingZeros() int {
	if x.Low > 0 {
		return bits.TrailingZeros64(x.Low)
	}
	return 64 + bits.TrailingZeros64(x.High)
}
