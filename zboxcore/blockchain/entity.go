// Methods and types for blockchain entities and interactions.
package blockchain

import (
	"sync/atomic"

	"github.com/0chain/gosdk/zboxcore/marker"
)

// StakePoolSettings information.
type StakePoolSettings struct {
	DelegateWallet string  `json:"delegate_wallet"`
	NumDelegates   int     `json:"num_delegates"`
	ServiceCharge  float64 `json:"service_charge"`
}

// UpdateStakePoolSettings represent stake pool information of a provider node.
type UpdateStakePoolSettings struct {
	DelegateWallet *string  `json:"delegate_wallet,omitempty"`
	NumDelegates   *int     `json:"num_delegates,omitempty"`
	ServiceCharge  *float64 `json:"service_charge,omitempty"`
}

// ValidationNode represents a validation node (miner)
type ValidationNode struct {
	ID                string            `json:"id"`
	BaseURL           string            `json:"url"`
	StakePoolSettings StakePoolSettings `json:"stake_pool_settings"`
}

// UpdateValidationNode represents a validation node (miner) update
type UpdateValidationNode struct {
	ID                string                   `json:"id"`
	BaseURL           *string                  `json:"url"`
	StakePoolSettings *UpdateStakePoolSettings `json:"stake_pool_settings"`
}

// StorageNode represents a storage node (blobber)
type StorageNode struct {
	ID             string `json:"id"`
	Baseurl        string `json:"url"`
	AllocationRoot string `json:"-"`
	LatestWM       *marker.WriteMarker

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
