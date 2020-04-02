package common

import (
	"fmt"
	"strconv"
)

const tokenUnit = 10000000000.0

// Balance represents an amount of tokens.
type Balance int64

// ToToken converts the Balance to ZCN tokens.
func (b Balance) ToToken() float64 {
	return float64(b) / tokenUnit
}

// ToBalance converts ZCN tokens to Balance.
func ToBalance(tok float64) Balance {
	return Balance(tok * tokenUnit)
}

// String implements fmt.Stringer interface.
func (b Balance) String() string {
	return strconv.FormatFloat(b.ToToken(), 'f', -1, 64)
}

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
