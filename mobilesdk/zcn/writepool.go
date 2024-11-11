//go:build mobile
// +build mobile

package zcn

import (
	"github.com/0chain/gosdk/zboxcore/sdk"
)

// WritePoolLock locks given number of tokes for given duration in read pool.
// ## Inputs
//   - allocID: allocation id
//   - tokens:  sas tokens
//   - fee: sas tokens
func WritePoolLock(allocID string, tokens, fee string) (string, error) {
	hash, _, err := sdk.WritePoolLock(allocID, tokens, fee)

	return hash, err
}
