//go:build !mobile
// +build !mobile

package zcncore

import (
	"time"

	"github.com/0chain/gosdk/core/common"
)

// Provider represents the type of provider.
type Provider int

const (
	ProviderMiner Provider = iota + 1
	ProviderSharder
	ProviderBlobber
	ProviderValidator
	ProviderAuthorizer
)

type ConfirmationStatus int

type Miner struct {
	ID         string      `json:"id"`
	N2NHost    string      `json:"n2n_host"`
	Host       string      `json:"host"`
	Port       int         `json:"port"`
	PublicKey  string      `json:"public_key"`
	ShortName  string      `json:"short_name"`
	BuildTag   string      `json:"build_tag"`
	TotalStake int64       `json:"total_stake"`
	Stat       interface{} `json:"stat"`
}

// Node represents a node (miner or sharder) in the network.
type Node struct {
	Miner     Miner `json:"simple_miner"`
	StakePool `json:"stake_pool"`
}

// MinerSCNodes list of nodes registered to the miner smart contract
type MinerSCNodes struct {
	Nodes []Node `json:"Nodes"`
}

type DelegatePool struct {
	Balance      int64  `json:"balance"`
	Reward       int64  `json:"reward"`
	Status       int    `json:"status"`
	RoundCreated int64  `json:"round_created"` // used for cool down
	DelegateID   string `json:"delegate_id"`
}

type StakePool struct {
	Pools    map[string]*DelegatePool `json:"pools"`
	Reward   int64                    `json:"rewards"`
	Settings StakePoolSettings        `json:"settings"`
	Minter   int                      `json:"minter"`
}

type stakePoolRequest struct {
	ProviderType Provider `json:"provider_type,omitempty"`
	ProviderID   string   `json:"provider_id,omitempty"`
}

type MinerSCDelegatePoolInfo struct {
	ID         common.Key     `json:"id"`
	Balance    common.Balance `json:"balance"`
	Reward     common.Balance `json:"reward"`      // uncollected reread
	RewardPaid common.Balance `json:"reward_paid"` // total reward all time
	Status     string         `json:"status"`
}

// MinerSCUserPoolsInfo represents the user stake pools information
type MinerSCUserPoolsInfo struct {
	Pools map[string][]*MinerSCDelegatePoolInfo `json:"pools"`
}

// PriceRange represents a price range allowed by user to filter blobbers.
type PriceRange struct {
	Min common.Balance `json:"min"`
	Max common.Balance `json:"max"`
}

// CreateAllocationRequest is information to create allocation.
type CreateAllocationRequest struct {
	DataShards      int              `json:"data_shards"`
	ParityShards    int              `json:"parity_shards"`
	Size            common.Size      `json:"size"`
	Expiration      common.Timestamp `json:"expiration_date"`
	Owner           string           `json:"owner_id"`
	OwnerPublicKey  string           `json:"owner_public_key"`
	Blobbers        []string         `json:"blobbers"`
	ReadPriceRange  PriceRange       `json:"read_price_range"`
	WritePriceRange PriceRange       `json:"write_price_range"`
}

type StakePoolSettings struct {
	DelegateWallet *string  `json:"delegate_wallet,omitempty"`
	NumDelegates   *int     `json:"num_delegates,omitempty"`
	ServiceCharge  *float64 `json:"service_charge,omitempty"`
}

type Terms struct {
	ReadPrice        common.Balance `json:"read_price"`  // tokens / GB
	WritePrice       common.Balance `json:"write_price"` // tokens / GB `
	MaxOfferDuration time.Duration  `json:"max_offer_duration"`
}

// Blobber represents a blobber node.
type Blobber struct {
	// ID is the blobber ID.
	ID common.Key `json:"id"`
	// BaseURL is the blobber's base URL used to access the blobber
	BaseURL string `json:"url"`
	// Terms of storage service of the blobber (read/write price, max offer duration)
	Terms Terms `json:"terms"`
	// Capacity is the total capacity of the blobber
	Capacity common.Size `json:"capacity"`
	// Used is the capacity of the blobber used to create allocations
	Allocated common.Size `json:"allocated"`
	// LastHealthCheck is the last time the blobber was checked for health
	LastHealthCheck common.Timestamp `json:"last_health_check"`
	// StakePoolSettings is the settings of the blobber's stake pool
	StakePoolSettings StakePoolSettings `json:"stake_pool_settings"`
	// NotAvailable is true if the blobber is not available
	NotAvailable bool `json:"not_available"`
	// IsRestricted is true if the blobber is restricted.
	// Restricted blobbers needs to be authenticated using AuthTickets in order to be used for allocation creation.
	// Check Restricted Blobbers documentation for more details.
	IsRestricted bool `json:"is_restricted"`
}

type Validator struct {
	ID                common.Key        `json:"id"`
	BaseURL           string            `json:"url"`
	StakePoolSettings StakePoolSettings `json:"stake_pool_settings"`
}

// AddAuthorizerPayload represents the payload for adding an authorizer.
type AddAuthorizerPayload struct {
	PublicKey         string                      `json:"public_key"`
	URL               string                      `json:"url"`
	StakePoolSettings AuthorizerStakePoolSettings `json:"stake_pool_settings"` // Used to initially create stake pool
}

// DeleteAuthorizerPayload represents the payload for deleting an authorizer.
type DeleteAuthorizerPayload struct {
	ID string `json:"id"` // authorizer ID
}

// AuthorizerHealthCheckPayload represents the payload for authorizer health check.
type AuthorizerHealthCheckPayload struct {
	ID string `json:"id"` // authorizer ID
}

// AuthorizerStakePoolSettings represents the settings for an authorizer's stake pool.
type AuthorizerStakePoolSettings struct {
	DelegateWallet string  `json:"delegate_wallet"`
	NumDelegates   int     `json:"num_delegates"`
	ServiceCharge  float64 `json:"service_charge"`
}

type AuthorizerConfig struct {
	Fee common.Balance `json:"fee"`
}

// InputMap represents a map of input fields.
type InputMap struct {
	Fields map[string]string `json:"Fields"`
}
