package common

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"

	"github.com/shopspring/decimal"
)

const (
	ZCNExponent = 10
	// TokenUnit represents the minimum token unit (sas)
	TokenUnit = 1e10
)

var (
	// ErrNegativeValue is returned if a float value is a negative number
	ErrNegativeValue = errors.New("negative coin value")
	// ErrTooManyDecimals is returned if a value has more than 10 decimal places
	ErrTooManyDecimals = errors.New("too many decimal places")
	// ErrTooLarge is returned if a value is greater than math.MaxInt64
	ErrTooLarge = errors.New("value is too large")
	// ErrUint64OverflowsFloat64 is returned if when converting a uint64 to a float64 overflow float64
	ErrUint64OverflowsFloat64 = errors.New("uint64 overflows float64")
	// ErrUint64AddOverflow is returned if when adding uint64 values overflow uint64
	ErrUint64AddOverflow = errors.New("uint64 addition overflow")
)

// A Key represents an identifier. It can be a pool ID, client ID, smart
// contract address, etc.
type Key string

// A Size represents a size in bytes.
type Size int64

func byteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

// String implements fmt.Stringer interface
func (s Size) String() string {
	return byteCountIEC(int64(s))
}

/* Balance */

// reParseToken is a regexp to parse string representation of token
var reParseToken = regexp.MustCompile(`^((?:\d*\.)?\d+)\s+(SAS|sas|uZCN|uzcn|mZCN|mzcn|ZCN|zcn)$`)

// Balance represents client's balance in Züs native token fractions (SAS = 10^-10 ZCN).
type Balance uint64

// ToToken converts Balance to ZCN tokens.
func (b Balance) ToToken() (float64, error) {
	if b > math.MaxInt64 {
		return 0.0, ErrTooLarge
	}

	f, _ := decimal.New(int64(b), -ZCNExponent).Float64()
	return f, nil
}

// String implements fmt.Stringer interface.
func (b Balance) String() string {
	if val, err := b.AutoFormat(); err == nil {
		return val
	}
	return ""
}

// Format returns a string representation of the balance with the given unit.
//   - unit is the balance unit.
func (b Balance) Format(unit BalanceUnit) (string, error) {
	v := float64(b)
	if v < 0 {
		return "", ErrUint64OverflowsFloat64
	}
	switch unit {
	case SAS:
		return fmt.Sprintf("%d %v", b, unit), nil
	case UZCN:
		v /= 1e4
	case MZCN:
		v /= 1e7
	case ZCN:
		v /= 1e10
	default:
		return "", fmt.Errorf("undefined balance unit: %d", unit)
	}
	return fmt.Sprintf("%.3f %v", v, unit), nil
}

// AutoFormat returns a string representation of the balance with the most
func (b Balance) AutoFormat() (string, error) {
	switch {
	case b/1e10 > 0:
		return b.Format(ZCN)
	case b/1e7 > 0:
		return b.Format(MZCN)
	case b/1e4 > 0:
		return b.Format(UZCN)
	}
	return b.Format(SAS)
}

// ToBalance converts ZCN tokens to Balance.
//   - token amount of ZCN tokens.
func ToBalance(token float64) (Balance, error) {
	d := decimal.NewFromFloat(token)
	if d.Sign() == -1 {
		return 0, ErrNegativeValue
	}

	// ZCN have a maximum of 10 decimal places
	if d.Exponent() < -ZCNExponent {
		return 0, ErrTooManyDecimals
	}

	// Multiply the coin balance by 1e10 to obtain coin amount
	e := d.Shift(ZCNExponent)

	// Check that there are no decimal places remaining. This error should not
	// occur, because of the earlier check of ZCNExponent()
	if e.Exponent() < 0 {
		return 0, ErrTooManyDecimals
	}

	maxDecimal := decimal.NewFromInt(math.MaxInt64)
	// Values greater than math.MaxInt64 will overflow after conversion to int64
	if e.GreaterThan(maxDecimal) {
		return 0, ErrTooLarge
	}

	return Balance(e.IntPart()), nil
}

// AddBalance adds c and b, returning an error if the values overflow
func AddBalance(c, b Balance) (Balance, error) {
	sum := c + b
	if sum < c || sum < b {
		return 0, ErrUint64AddOverflow
	}
	return sum, nil
}

// FormatBalance returns a string representation of the balance with the given unit.
func FormatBalance(b Balance, unit BalanceUnit) (string, error) {
	return b.Format(unit)
}

// AutoFormatBalance returns a string representation of the balance with the most
func AutoFormatBalance(b Balance) (string, error) {
	return b.AutoFormat()
}

func ParseBalance(str string) (Balance, error) {

	matches := reParseToken.FindAllStringSubmatch(str, -1)

	if len(matches) != 1 || len(matches[0]) != 3 {
		return 0, fmt.Errorf("invalid input: %s", str)
	}

	b, err := strconv.ParseFloat(matches[0][1], 64)
	if err != nil {
		return 0, err
	}

	var unit BalanceUnit

	err = unit.Parse(matches[0][2])
	if err != nil {
		return 0, err
	}

	switch unit {
	case UZCN:
		b *= 1e4
	case MZCN:
		b *= 1e7
	case ZCN:
		b *= 1e10
	}

	return Balance(b), nil
}

const (
	SAS BalanceUnit = iota
	UZCN
	MZCN
	ZCN
)

type BalanceUnit byte

func (unit BalanceUnit) String() string {
	switch unit {
	case SAS:
		return "SAS"
	case MZCN:
		return "mZCN"
	case UZCN:
		return "uZCN"
	case ZCN:
		return "ZCN"
	}
	return ""
}

func (unit *BalanceUnit) Parse(s string) error {
	switch s {
	case "SAS", "sas":
		*unit = SAS
	case "uZCN", "uzcn":
		*unit = UZCN
	case "mZCN", "mzcn":
		*unit = MZCN
	case "ZCN", "zcn":
		*unit = ZCN
	default:
		return errors.New("undefined balance unit: " + s)
	}
	return nil
}

func ParseBalanceStatic(str string) (int64, error) {
	bal, err := ParseBalance(str)
	return int64(bal), err
}

func FormatStatic(amount int64, unit string) (string, error) {
	token := Balance(amount)

	var unitB BalanceUnit
	err := unitB.Parse(unit)
	if err != nil {
		return "", err
	}

	return token.Format(unitB)
}

func AutoFormatStatic(amount int64) (string, error) {
	token := Balance(amount)
	return token.AutoFormat()
}
