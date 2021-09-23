package magmasc

import (
	"encoding/json"

	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zmagmacore/config"
	"github.com/0chain/gosdk/zmagmacore/errors"
	"github.com/0chain/gosdk/zmagmacore/node"
)

type (
	// Provider represents providers node stored in blockchain.
	Provider struct {
		ID       string `json:"id"`
		ExtID    string `json:"ext_id"`
		Host     string `json:"host,omitempty"`
		MinStake int64  `json:"min_stake,omitempty"`
	}
)

var (
	// Make sure Provider implements Serializable interface.
	_ util.Serializable = (*Provider)(nil)
)

// NewProviderFromCfg creates Provider from config.Provider.
func NewProviderFromCfg(cfg *config.Provider) *Provider {
	return &Provider{
		ID:       node.ID(),
		ExtID:    cfg.ExtID,
		Host:     cfg.Host,
		MinStake: cfg.MinStake,
	}
}

// Decode implements util.Serializable interface.
func (m *Provider) Decode(blob []byte) error {
	var provider Provider
	if err := json.Unmarshal(blob, &provider); err != nil {
		return errDecodeData.Wrap(err)
	}
	if err := provider.Validate(); err != nil {
		return err
	}

	m.ID = provider.ID
	m.ExtID = provider.ExtID
	m.Host = provider.Host
	m.MinStake = provider.MinStake

	return nil
}

// Encode implements util.Serializable interface.
func (m *Provider) Encode() []byte {
	blob, _ := json.Marshal(m)
	return blob
}

// GetType returns Provider's type.
func (m *Provider) GetType() string {
	return providerType
}

// Validate checks Provider for correctness.
// If it is not return errInvalidProvider.
func (m *Provider) Validate() (err error) {
	switch { // is invalid
	case m.ExtID == "":
		err = errors.New(errCodeBadRequest, "provider external id is required")

	case m.Host == "":
		err = errors.New(errCodeBadRequest, "provider host is required")

	default:
		return nil // is valid
	}

	return errInvalidProvider.Wrap(err)
}
