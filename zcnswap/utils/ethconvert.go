package utils

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// ErrInvalidUnit is returned when the given unit is not supported.
var ErrInvalidUnit = fmt.Errorf("invalid unit")

func init() {
	// set the decimal precision to 18
	decimal.DivisionPrecision = 18
}

// Convert converts the given amount from the given unit to the given unit.
func Convert(input decimal.Decimal, inputUnit string, outputUnit string) (decimal.Decimal, error) {
	wei, err := ToWei(input, inputUnit)
	if err != nil {
		return decimal.Zero, err
	}
	return FromWei(wei, outputUnit)
}

// ToWei converts the given amount to Wei.
func ToWei(amount decimal.Decimal, unit string) (decimal.Decimal, error) {
	if unit == Wei {
		return amount.RoundDown(0), nil
	}
	value, err := getUnitValue(unit)
	if err != nil {
		return decimal.Zero, err
	}
	return amount.Mul(value).RoundDown(0), nil
}

// FromWei converts the given amount from Wei to the given unit.
func FromWei(amount decimal.Decimal, unit string) (decimal.Decimal, error) {
	if unit == Wei {
		return amount.RoundDown(0), nil
	}
	value, err := getUnitValue(unit)
	if err != nil {
		return decimal.Zero, err
	}
	return amount.Div(value), nil
}

func getUnitValue(unit string) (decimal.Decimal, error) {
	value, ok := units[unit]
	if !ok {
		return decimal.Zero, ErrInvalidUnit
	}
	return decimal.NewFromString(value)
}
