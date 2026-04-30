//go:build mobile
// +build mobile

package zcn

import (
	"strconv"

	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zcncore"
)

// WritePoolLock locks given number of tokes for given duration in read pool.
// ## Inputs
//   - allocID: allocation id
//   - tokens:  sas tokens
//   - fee: sas tokens
func WritePoolLock(allocID string, tokens, fee string) (string, error) {
	tokensUint, err := strconv.ParseUint(tokens, 10, 64)

	if err != nil {
		return "", err
	}

	feeUint, err := strconv.ParseUint(fee, 10, 64)

	if err != nil {
		return "", err
	}

	hash, _, err := sdk.WritePoolLock(
		allocID,
		tokensUint,
		feeUint,
	)

	return hash, err
}

// ConvertToValueMobile converts ZCN tokens to SAS tokens (mobile wrapper)
// ## Inputs
//   - token: ZCN tokens as float64
//
// ## Outputs
//   - SAS tokens as uint64
func ConvertToValueMobile(token float64) uint64 {
	return zcncore.ConvertToValue(token)
}
