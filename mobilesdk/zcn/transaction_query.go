//go:build mobile
// +build mobile

package zcn

import (
	transaction_query_mobile "github.com/0chain/gosdk/zcncore"
)

// GetUserLockedTotal get total token user locked
// # Inputs
//   - clientID wallet id
func GetUserLockedTotal(clientID string) (int64, error) {
	return transaction_query_mobile.GetUserLockedTotal(clientID)
}
