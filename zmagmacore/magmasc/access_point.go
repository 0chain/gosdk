package magmasc

import (
	"encoding/json"
	"time"

	"github.com/0chain/errors"
	"github.com/magma/augmented-networks/accounting/protos"

	"github.com/0chain/gosdk/zmagmacore/config"
	"github.com/0chain/gosdk/zmagmacore/node"
	ctime "github.com/0chain/gosdk/zmagmacore/time"
)

// AccessPoint represents access point node stored in blockchain.
type AccessPoint struct {
	ID            string `json:"id"`
	Terms         Terms  `json:"terms,omitempty"`
	MinStake      int64  `json:"min_stake,omitempty"`
	ProviderExtID string `json:"provider_ext_id"`
}

// NewAccessPointFromCfg creates AccessPoint from config.AccessPoint
func NewAccessPointFromCfg(cfg *config.AccessPoint) *AccessPoint {
	return &AccessPoint{
		ID: node.ID(),
		Terms: Terms{
			Price:           cfg.Terms.Price,
			PriceAutoUpdate: cfg.Terms.PriceAutoUpdate,
			MinCost:         cfg.Terms.MinCost,
			Volume:          cfg.Terms.Volume,
			QoS: &protos.QoS{
				DownloadMbps: cfg.Terms.QoS.DownloadMbps,
				UploadMbps:   cfg.Terms.QoS.UploadMbps,
			},
			QoSAutoUpdate: &QoSAutoUpdate{
				DownloadMbps: cfg.Terms.QoSAutoUpdate.DownloadMbps,
				UploadMbps:   cfg.Terms.QoSAutoUpdate.DownloadMbps,
			},
			ProlongDuration: cfg.Terms.ProlongDuration,
			ExpiredAt:       ctime.Timestamp(time.Now().Add(cfg.Terms.ExpiredAt).Unix()),
		},
		ProviderExtID: cfg.ProviderExtID,
	}
}

// Decode implements util.Serializable interface.
func (m *AccessPoint) Decode(blob []byte) error {
	var accessPoint AccessPoint
	if err := json.Unmarshal(blob, &accessPoint); err != nil {
		return ErrDecodeData.Wrap(err)
	}
	if err := accessPoint.Validate(); err != nil {
		return err
	}

	m.ID = accessPoint.ID
	m.MinStake = accessPoint.MinStake
	m.ProviderExtID = accessPoint.ProviderExtID
	m.Terms = accessPoint.Terms

	return nil
}

// Encode implements util.Serializable interface.
func (m *AccessPoint) Encode() []byte {
	blob, _ := json.Marshal(m)
	return blob
}

// GetType returns node type.
func (m *AccessPoint) GetType() string {
	return AccessPointType
}

// Validate checks the AccessPoint for correctness.
// If it is not return errInvalidAccessPoint.
func (m *AccessPoint) Validate() (err error) {
	switch { // is invalid
	case m.ID == "":
		err = errors.New(ErrCodeBadRequest, "accessPoint external id is required")

	case m.ProviderExtID == "":
		err = errors.New(ErrCodeBadRequest, "accessPoint provider external id is required")

	default:
		return nil // is valid
	}

	return ErrInvalidAccessPoint.Wrap(err)
}
