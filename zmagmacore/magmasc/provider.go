package magmasc

import (
	"encoding/json"
	"time"

	"github.com/magma/augmented-networks/accounting/protos"

	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zmagmacore/config"
	"github.com/0chain/gosdk/zmagmacore/errors"
	"github.com/0chain/gosdk/zmagmacore/node"
	ctime "github.com/0chain/gosdk/zmagmacore/time"
)

type (
	// Provider represents providers node stored in blockchain.
	Provider struct {
		ID    string    `json:"id"`
		ExtID string    `json:"ext_id"`
		Host  string    `json:"host,omitempty"`
		Terms TermsList `json:"terms"`
	}
)

var (
	// Make sure Provider implements Serializable interface.
	_ util.Serializable = (*Provider)(nil)
)

// NewProviderFromCfg creates Consumer from config.Consumer.
func NewProviderFromCfg(cfg *config.Provider) *Provider {
	terms := make(map[string]ProviderTerms, len(cfg.Terms))
	for _, item := range cfg.Terms {
		terms[item.AccessPointID] = ProviderTerms{
			AccessPointID:   item.AccessPointID,
			Price:           item.Price,
			PriceAutoUpdate: item.PriceAutoUpdate,
			MinCost:         item.MinCost,
			Volume:          item.Volume,
			QoS: &protos.QoS{
				DownloadMbps: item.QoS.DownloadMbps,
				UploadMbps:   item.QoS.UploadMbps,
			},
			QoSAutoUpdate: &QoSAutoUpdate{
				DownloadMbps: item.QoSAutoUpdate.DownloadMbps,
				UploadMbps:   item.QoSAutoUpdate.UploadMbps,
			},
			ProlongDuration: time.Duration(item.ProlongDuration),
			ExpiredAt:       ctime.Timestamp(time.Now().Add(time.Duration(item.ExpiredAt) * time.Minute).Unix()),
		}
	}

	return &Provider{
		ID:    node.ID(),
		ExtID: cfg.ExtID,
		Host:  cfg.Host,
		Terms: terms,
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
	m.Terms = provider.Terms

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
