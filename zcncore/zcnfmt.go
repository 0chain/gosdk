package zcncore

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

const (
	SAS Unit = iota
	UZCN
	MZCN
	ZCN
)

var (
	ErrUndefinedFormat = errors.New("undefined format")

	// regexp to parse zcn
	reParseZCN = regexp.MustCompile(`(\d+)\s+(SAS|sas|uZCN|uzcn|mZCN|mzcn|ZCN|zcn)`)
)

type Unit byte

func (unit Unit) String() string {
	switch unit {
	case SAS:
		return "SAS"
	case MZCN:
		return "mZCN"
	case UZCN:
		return "uZCN"
	}
	return "ZCN"
}

func (unit *Unit) Parse(s string) error {
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
		return ErrUndefinedFormat
	}
	return nil
}

func ParseUnit(s string) (Unit, error) {
	var unit Unit
	err := unit.Parse(s)
	if err != nil {
		return unit, err
	}
	return unit, nil
}

// Balance is a type used for 0chain native token
type Balance int64

func (token Balance) Format(unit Unit) string {
	switch unit {
	case SAS:
		return fmt.Sprintf("%d SAS", token)
	case UZCN:
		return fmt.Sprintf("%.3f uZCN", float64(token)/1e3)
	case MZCN:
		return fmt.Sprintf("%.3f mZCN", float64(token)/1e6)
	}
	return fmt.Sprintf("%.3f ZCN", float64(token)/1e9)
}

func (token Balance) AutoFormat() string {
	switch {
	case token/1e9 > 0:
		return token.Format(ZCN)
	case token/1e6 > 0:
		return token.Format(MZCN)
	case token/1e3 > 0:
		return token.Format(UZCN)
	}
	return token.Format(SAS)
}

func Format(token Balance, unit Unit) string {
	return token.Format(unit)
}

func AutoFormat(token Balance) string {
	return token.AutoFormat()
}

func Parse(zcnstr string) (Balance, error) {

	matches := reParseZCN.FindAllStringSubmatch(zcnstr, 1)

	if len(matches) != 1 {
		return 0, fmt.Errorf("invalid input: %s", zcnstr)
	}

	token, err := strconv.ParseInt(matches[0][0], 10, 64)
	if err != nil {
		return 0, err
	}

	switch matches[0][1] {
	case "uZCN", "uzcn":
		token *= 10e3
	case "mZCN", "mzcn":
		token *= 10e6
	case "ZCN", "zcn":
		token *= 10e9
	}

	return Balance(token), nil
}
