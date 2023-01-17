package util

import (
	"fmt"
	"strconv"
)

func ParseCoinStr(vs string) (uint64, error) {
	if vs == "" {
		return 0, nil
	}

	v, err := strconv.ParseUint(vs, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid token value: %v, err: %v", vs, err)
	}

	return v, nil
}
