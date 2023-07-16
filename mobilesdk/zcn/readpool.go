//go:build mobile
// +build mobile

package zcn

import (
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

// ReadPoolLock locks given number of tokes for given duration in read pool.
// ## Inputs
//   - tokens:  sas tokens
//   - fee: sas tokens
func ReadPoolLock(tokens, fee string) (string, error) {
	t, err := util.ParseCoinStr(tokens)
	if err != nil {
		return "", err
	}

	f, err := util.ParseCoinStr(fee)
	if err != nil {
		return "", err
	}

	hash, _, err := sdk.ReadPoolLock(t, f)
	return hash, err
}

// ReadPoolUnLock unlocks all the tokens in the readpool associated with the current wallet.
// ## Inputs
//   - fee: sas tokens
func ReadPoolUnLock(fee string) (string, error) {
	f, err := util.ParseCoinStr(fee)
	if err != nil {
		return "", err
	}

	hash, _, err := sdk.ReadPoolUnlock(f)
	return hash, err
}
