//go:build mobile
// +build mobile

package zcn

import (
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

// WritePoolLock locks given number of tokes for given duration in read pool.
// ## Inputs
//   - allocID: allocation id
//   - tokens:  sas tokens
//   - fee: sas tokens
func WritePoolLock(allocID string, tokens, fee string) (string, error) {
	t, err := util.ParseCoinStr(tokens)
	if err != nil {
		return "", err
	}

	f, err := util.ParseCoinStr(fee)
	hash, _, err := sdk.WritePoolLock(allocID, t, f)

	return hash, err
}
