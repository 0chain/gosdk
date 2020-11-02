package common

import (
	"fmt"
	"strconv"
	"strings"
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

// WhoPays for file downloading.
type WhoPays int

// possible variants
const (
	WhoPays3rdParty WhoPays = iota // 0, 3rd party user pays
	WhoPaysOwner                   // 1, file owner pays
)

// String implements fmt.Stringer interface.
func (wp WhoPays) String() string {
	switch wp {
	case WhoPays3rdParty:
		return "3rd_party"
	case WhoPaysOwner:
		return "owner"
	}
	return fmt.Sprintf("WhoPays(%d)", int(wp))
}

// Validate the WhoPays value.
func (wp WhoPays) Validate() (err error) {
	switch wp {
	case WhoPays3rdParty, WhoPaysOwner:
		return // ok
	}
	return fmt.Errorf("unknown WhoPays value: %d", int(wp))
}

// Parse given string and set the WhoPays by it. Or return parsing error.
// The given string should be as result of the String method (case insensitive).
func (wp *WhoPays) Parse(val string) (err error) {
	switch strings.ToLower(val) {
	case "owner":
		(*wp) = WhoPaysOwner
	case "3rd_party":
		(*wp) = WhoPays3rdParty
	default:
		err = fmt.Errorf("empty or unknown 'who_pays' value: %q", val)
	}
	return
}
