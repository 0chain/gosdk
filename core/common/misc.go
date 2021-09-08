package common

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/0chain/errors"
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

// WhoPays for file downloading.
type WhoPays int

// possible variants
const (
	WhoPaysOwner    WhoPays = iota // 0, file owner pays
	WhoPays3rdParty                // 1, 3rd party user pays
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
	return errors.New("validate_error", fmt.Sprintf("unknown WhoPays value: %d", int(wp)))
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
		err = errors.New("parse_error", fmt.Sprintf("empty or unknown 'who_pays' value: %q", val))
	}
	return
}

/* Balance */

// minimum token unit (sas)
const tokenUnit = 1e10

// reParseToken is a regexp to parse string representation of token
var reParseToken = regexp.MustCompile(`^((?:\d*\.)?\d+)\s+(SAS|sas|uZCN|uzcn|mZCN|mzcn|ZCN|zcn)$`)

// Balance represents 0chain native token
type Balance int64

func (b Balance) ToToken() float64 {
	return float64(b) / tokenUnit
}

// String implements fmt.Stringer interface.
func (b Balance) String() string {
	return b.AutoFormat()
}

func (b Balance) Format(unit BalanceUnit) string {
	v := float64(b)
	switch unit {
	case SAS:
		return fmt.Sprintf("%d %v", b, unit)
	case UZCN:
		v /= 1e4
	case MZCN:
		v /= 1e7
	case ZCN:
		v /= 1e10
	}
	return fmt.Sprintf("%.3f %v", v, unit)
}

func (b Balance) AutoFormat() string {
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
func ToBalance(token float64) Balance {
	return Balance(token * tokenUnit)
}

func FormatBalance(b Balance, unit BalanceUnit) string {
	return b.Format(unit)
}

func AutoFormatBalance(b Balance) string {
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
		return errors.New("", "undefined balance unit: "+s)
	}
	return nil
}
