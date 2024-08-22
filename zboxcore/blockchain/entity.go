package blockchain

import (
	"sync/atomic"

	"github.com/0chain/gosdk/core/common"
)

// StakePoolSettings information.
type StakePoolSettings struct {
	DelegateWallet string         `json:"delegate_wallet"`
	MinStake       common.Balance `json:"min_stake"`
	MaxStake       common.Balance `json:"max_stake"`
	NumDelegates   int            `json:"num_delegates"`
	ServiceCharge  float64        `json:"service_charge"`
}

// UpdateStakePoolSettings information.
type UpdateStakePoolSettings struct {
	DelegateWallet *string         `json:"delegate_wallet,omitempty"`
	MinStake       *common.Balance `json:"min_stake,omitempty"`
	MaxStake       *common.Balance `json:"max_stake,omitempty"`
	NumDelegates   *int            `json:"num_delegates,omitempty"`
	ServiceCharge  *float64        `json:"service_charge,omitempty"`
}

type ValidationNode struct {
	ID                string            `json:"id"`
	BaseURL           string            `json:"url"`
	StakePoolSettings StakePoolSettings `json:"stake_pool_settings"`
}

type UpdateValidationNode struct {
	ID                string                   `json:"id"`
	BaseURL           *string                  `json:"url"`
	StakePoolSettings *UpdateStakePoolSettings `json:"stake_pool_settings"`
}

type StorageNode struct {
	ID      string `json:"id"`
	Baseurl string `json:"url"`

	skip uint64 `json:"-"` // skip on error
}

func (sn *StorageNode) SetSkip(t bool) {
	var val uint64
	if t {
		val = 1
	}
	atomic.StoreUint64(&sn.skip, val)
}

func (sn *StorageNode) IsSkip() bool {
	return atomic.LoadUint64(&sn.skip) > 0
}
