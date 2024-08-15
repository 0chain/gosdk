package zboxutil

import (
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUint128(t *testing.T) {
	type args struct {
		x uint64
	}
	tests := []struct {
		name string
		args args
		want Uint128
	}{
		{
			name: "OK",
			args: args{x: 5},
			want: Uint128{
				Low: 5,
			},
		},
		{
			name: "0_OK",
			args: args{x: 0},
			want: Uint128{
				Low: 0,
			},
		},
		{
			name: "Max_OK",
			args: args{x: math.MaxUint64},
			want: Uint128{
				Low: math.MaxUint64,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewUint128(tt.args.x); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUint128() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint128_Add(t *testing.T) {
	tests := []struct {
		x         Uint128
		y         Uint128
		want      Uint128
		wantPanic bool
	}{
		{},
		{
			x: NewUint128(^uint64(0)),
			y: NewUint128(1),
			want: Uint128{
				High: 1,
				Low:  0,
			},
		},
		{
			x: NewUint128(^uint64(0)),
			y: NewUint128(^uint64(0)),
			want: Uint128{
				High: 1,
				Low:  ^uint64(0) - 1,
			},
		},
		{
			x: Uint128{
				Low:  ^uint64(0),
				High: ^uint64(0),
			},
			y:         NewUint128(1),
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		func() {
			defer func() {
				pan := recover()
				assert.True(t, tt.wantPanic != (pan == nil))
			}()

			if got := tt.x.Add(tt.y); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Add() = %v, want %v", got, tt.want)
			}
		}()
	}
}

func TestUint128_Add64(t *testing.T) {
	testData := []struct {
		num       Uint128
		expected  Uint128
		add       uint64
		wantPanic bool
	}{
		{Uint128{0, 0}, Uint128{0, 1}, 1, false},
		{Uint128{18446744073709551615, 18446744073709551614}, Uint128{18446744073709551615, 18446744073709551615}, 1, false},
		{Uint128{0, 18446744073709551615}, Uint128{1, 0}, 1, false},
		{Uint128{18446744073709551615, 0}, Uint128{18446744073709551615, 1}, 1, false},
		{Uint128{0, 18446744073709551615}, Uint128{1, 24}, 25, false},
		{Uint128{^uint64(0), ^uint64(0)}, Uint128{}, 1, true},
	}

	for _, test := range testData {
		func() {
			defer func() {
				pan := recover()
				assert.True(t, test.wantPanic != (pan == nil))
			}()
			res := test.num.Add64(test.add)
			if res != test.expected {
				t.Errorf("expected: %v + %d = %v but got %v", test.num, test.add, test.expected, res)
			}
		}()
	}
}

func TestUint128_Equals(t *testing.T) {
	testData := []struct {
		u1       Uint128
		u2       Uint128
		expected bool
	}{
		{Uint128{0, 0}, Uint128{0, 1}, false},
		{Uint128{1, 0}, Uint128{0, 1}, false},
		{Uint128{18446744073709551615, 18446744073709551614}, Uint128{18446744073709551615, 18446744073709551615}, false},
		{Uint128{0, 1}, Uint128{0, 1}, true},
		{Uint128{0, 0}, Uint128{0, 0}, true},
		{Uint128{314, 0}, Uint128{314, 0}, true},
		{Uint128{18446744073709551615, 18446744073709551615}, Uint128{18446744073709551615, 18446744073709551615}, true},
	}

	for _, test := range testData {
		if actual := test.u1.Equals(test.u2); actual != test.expected {
			t.Errorf("expected: %v.Equals(%v) expected %v but got %v", test.u1, test.u2, test.expected, actual)
		}
	}
}

func TestUint128_And(t *testing.T) {
	u1 := Uint128{14799720563850130797, 11152134164166830811}
	u2 := Uint128{10868624793753271583, 6542293553298186666}

	expected := Uint128{9529907221165552909, 1927615693132931210}
	if !(u1.And(u2)).Equals(expected) {
		t.Errorf("incorrect AND computation: %v & %v != %v", u1, u2, expected)
	}
}

func TestUint128_Or64(t *testing.T) {
	u1 := Uint128{14799720563850130797, 11152134164166830811}
	u2 := Uint128{10868624793753271583, 6542293553298186666}

	expected := Uint128{16138438136437849471, 15766812024332086267}
	if !(u1.Or(u2)).Equals(expected) {
		t.Errorf("incorrect OR computation: %v | %v != %v", u1, u2, expected)
	}
}

func TestUint128_Xor(t *testing.T) {
	u1 := Uint128{14799720563850130797, 11152134164166830811}
	u2 := Uint128{10868624793753271583, 6542293553298186666}

	expected := Uint128{6608530915272296562, 13839196331199155057}
	if !(u1.Xor(u2)).Equals(expected) {
		t.Errorf("incorrect XOR computation: %v ^ %v != %v", u1, u2, expected)
	}
}

func TestUint128_Sub(t *testing.T) {
	testData := []struct {
		num      Uint128
		expected Uint128
		sub      Uint128
	}{
		{Uint128{0, 1}, Uint128{0, 0}, NewUint128(1)},
		{Uint128{18446744073709551615, 18446744073709551615}, Uint128{18446744073709551615, 18446744073709551614}, NewUint128(1)},
		{Uint128{0, 18446744073709551615}, Uint128{0, 18446744073709551614}, NewUint128(1)},
		{Uint128{18446744073709551615, 0}, Uint128{18446744073709551614, 18446744073709551615}, NewUint128(1)},
		{Uint128{18446744073709551615, 0}, Uint128{18446744073709551614, 18446744073709551591}, NewUint128(25)},
		{Uint128{0, 0}, Uint128{18446744073709551615, 18446744073709551615}, Uint128{0, 1}},
	}

	for _, test := range testData {

		res := test.num.Sub(test.sub)
		if res != test.expected {
			t.Errorf("expected: %v - %d = %v but got %v", test.num, test.sub, test.expected, res)
		}
	}
}

func TestUint128_Sub64(t *testing.T) {
	testData := []struct {
		num      Uint128
		expected Uint128
		sub      uint64
	}{
		{Uint128{0, 1}, Uint128{0, 0}, 1},
		{Uint128{18446744073709551615, 18446744073709551615}, Uint128{18446744073709551615, 18446744073709551614}, 1},
		{Uint128{0, 18446744073709551615}, Uint128{0, 18446744073709551614}, 1},
		{Uint128{18446744073709551615, 0}, Uint128{18446744073709551614, 18446744073709551615}, 1},
		{Uint128{18446744073709551615, 0}, Uint128{18446744073709551614, 18446744073709551591}, 25},
		{Uint128{0, 0}, Uint128{18446744073709551615, 18446744073709551615}, 1},
	}

	for _, test := range testData {
		res := test.num.Sub64(test.sub)
		if res != test.expected {
			t.Errorf("expected: %v - %d = %v but got %v", test.num, test.sub, test.expected, res)
		}
	}
}

func TestUint128_Equals64(t *testing.T) {
	testData := []struct {
		u1       Uint128
		u2       uint64
		expected bool
	}{
		{Uint128{0, 0}, 1, false},
		{Uint128{1, 0}, 1, false},
		{Uint128{18446744073709551615, 18446744073709551614}, 18446744073709551615, false},
		{Uint128{0, 1}, 1, true},
		{Uint128{0, 0}, 0, true},
	}

	for _, test := range testData {
		if actual := test.u1.Equals64(test.u2); actual != test.expected {
			t.Errorf("expected: %v.Equals64(%v) expected %v but got %v", test.u1, test.u2, test.expected, actual)
		}
	}
}

func TestUint128_And64(t *testing.T) {
	var (
		u1        = Uint128{14799720563850130797, 11152134164166830811}
		u2 uint64 = 6542293553298186666
	)

	expected := Uint128{0, 1927615693132931210}
	if !u1.And64(u2).Equals(expected) {
		t.Errorf("incorrect AND computation: %v & %v != %v", u1, u2, expected)
	}
}

func TestUint128_Lsh(t *testing.T) {
	testData := []struct {
		u1       Uint128
		u2       uint64
		expected Uint128
	}{
		{Uint128{0, 1}, 5, Uint128{0, 32}},
		{Uint128{0, 1}, 63, Uint128{0, 1 << 63}},
		{Uint128{0, 1}, 64, Uint128{1, 0}},
		{Uint128{0, 1}, 127, Uint128{1 << 63, 0}},
	}

	for _, test := range testData {
		if actual := test.u1.Lsh(test.u2); actual != test.expected {
			t.Errorf("expected: %v.Lsn(%v) expected %v but got %v", test.u1, test.u2, test.expected, actual)
		}
	}
}

func TestUint128_Not(t *testing.T) {
	testData := []struct {
		u1       Uint128
		expected Uint128
	}{
		{Uint128{0, 0}, Uint128{^uint64(0), ^uint64(0)}},
		{Uint128{0, 1}, Uint128{^uint64(0), ^uint64(1)}},
		{Uint128{1, 0}, Uint128{^uint64(1), ^uint64(0)}},
	}

	for _, test := range testData {
		if actual := test.u1.Not(); actual != test.expected {
			t.Errorf("expected: %v.Not() expected %v but got %v", test.u1, test.expected, actual)
		}
	}
}

func TestUint128_CountOnes(t *testing.T) {
	testData := []struct {
		u1       Uint128
		expected int
	}{
		{Uint128{0, ^uint64(0)}, 64},
		{Uint128{^uint64(0), 1}, 65},
		{Uint128{^uint64(0), ^uint64(0)}, 128},
		{Uint128{0, 0}, 0},
	}

	for _, test := range testData {
		if actual := test.u1.CountOnes(); actual != test.expected {
			t.Errorf("expected: %v.CountOnes() expected %v but got %v", test.u1, test.expected, actual)
		}
	}
}

func TestUint128_TrailingZeros(t *testing.T) {
	testData := []struct {
		u1       Uint128
		expected int
	}{
		{Uint128{0, ^uint64(0)}, 0},
		{Uint128{0, 1 << 63}, 63},
		{Uint128{1 << 63, 0}, 127},
		{Uint128{0, 0}, 128},
	}

	for _, test := range testData {
		if actual := test.u1.TrailingZeros(); actual != test.expected {
			t.Errorf("expected: %v.TrailingZeros() expected %v but got %v", test.u1, test.expected, actual)
		}
	}
}
