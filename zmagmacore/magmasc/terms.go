package magmasc

import (
	"encoding/json"

	magma "github.com/magma/augmented-networks/accounting/protos"

	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zmagmacore/errors"
	"github.com/0chain/gosdk/zmagmacore/time"
)

type (
	// Terms represents a provider and service terms.
	Terms struct {
		Price           float32        `json:"price"`                       // tokens per Megabyte
		PriceAutoUpdate float32        `json:"price_auto_update,omitempty"` // price change on auto update
		MinCost         float32        `json:"min_cost"`                    // minimal cost for a session
		Volume          int64          `json:"volume"`                      // bytes per a session
		QoS             *magma.QoS     `json:"qos"`                         // quality of service guarantee
		QoSAutoUpdate   *QoSAutoUpdate `json:"qos_auto_update,omitempty"`   // qos change on auto update
		ProlongDuration time.Duration  `json:"prolong_duration,omitempty"`  // duration in seconds to prolong the terms
		ExpiredAt       time.Timestamp `json:"expired_at,omitempty"`        // timestamp till a session valid
	}

	// QoSAutoUpdate represents data of qos terms on auto update.
	QoSAutoUpdate struct {
		DownloadMbps float32 `json:"download_mbps"`
		UploadMbps   float32 `json:"upload_mbps"`
	}
)

var (
	// Make sure Terms implements Serializable interface.
	_ util.Serializable = (*Terms)(nil)
)

// NewTerms returns a new constructed provider terms.
func NewTerms() *Terms {
	return &Terms{QoS: &magma.QoS{}}
}

// Decode implements util.Serializable interface.
func (m *Terms) Decode(blob []byte) error {
	var terms Terms
	if err := json.Unmarshal(blob, &terms); err != nil {
		return ErrDecodeData.Wrap(err)
	}
	if err := terms.Validate(); err != nil {
		return err
	}

	m.Price = terms.Price
	m.PriceAutoUpdate = terms.PriceAutoUpdate
	m.MinCost = terms.MinCost
	m.Volume = terms.Volume
	m.QoS.UploadMbps = terms.QoS.UploadMbps
	m.QoS.DownloadMbps = terms.QoS.DownloadMbps
	m.QoSAutoUpdate = terms.QoSAutoUpdate
	m.ProlongDuration = terms.ProlongDuration
	m.ExpiredAt = terms.ExpiredAt

	return nil
}

// Decrease makes automatically Decrease provider terms by config.
func (m *Terms) Decrease() *Terms {
	m.Volume = 0 // the volume of terms must be zeroed

	if m.ProlongDuration != 0 {
		m.ExpiredAt += time.Timestamp(m.ProlongDuration) // prolong expire of terms
	}

	if m.PriceAutoUpdate != 0 && m.Price > m.PriceAutoUpdate {
		m.Price -= m.PriceAutoUpdate // down the price
	}

	if m.QoSAutoUpdate != nil {
		if m.QoSAutoUpdate.UploadMbps != 0 {
			m.QoS.UploadMbps += m.QoSAutoUpdate.UploadMbps // up the qos of upload mbps
		}

		if m.QoSAutoUpdate.DownloadMbps != 0 {
			m.QoS.DownloadMbps += m.QoSAutoUpdate.DownloadMbps // up the qos of download mbps
		}
	}

	return m
}

// Encode implements util.Serializable interface.
func (m *Terms) Encode() []byte {
	blob, _ := json.Marshal(m)
	return blob
}

// Expired returns if terms already expired.
func (m *Terms) Expired() bool {
	return m.ExpiredAt < time.Now()+TermsExpiredDuration
}

// GetAmount returns calculated amount value of provider terms.
func (m *Terms) GetAmount() (amount int64) {
	price := m.GetPrice()
	if price > 0 {
		amount = price * m.GetVolume()
		if minCost := m.GetMinCost(); amount < minCost {
			amount = minCost
		}
	}

	return amount
}

// GetMinCost returns calculated min cost value of provider terms.
func (m *Terms) GetMinCost() (cost int64) {
	if m.MinCost > 0 {
		cost = int64(m.MinCost * billion)
	}

	return cost
}

// GetPrice returns calculated price value of provider terms.
// NOTE: the price value will be represented in token units per megabyte.
func (m *Terms) GetPrice() (price int64) {
	if m.Price > 0 {
		price = int64(m.Price * billion)
	}

	return price
}

// GetVolume returns value of the provider terms volume.
// If the Volume is empty it will be calculated by the provider terms.
func (m *Terms) GetVolume() int64 {
	if m.Volume == 0 {
		mbps := (m.QoS.UploadMbps + m.QoS.DownloadMbps) / octet // megabytes per second
		duration := float32(m.ExpiredAt - time.Now())           // duration in seconds
		// rounded of bytes per second multiplied by duration in seconds
		m.Volume = int64(mbps * duration)
	}

	return m.Volume
}

// Increase makes automatically Increase provider terms by config.
func (m *Terms) Increase() *Terms {
	m.Volume = 0 // the volume of terms must be zeroed

	if m.ProlongDuration != 0 {
		m.ExpiredAt += time.Timestamp(m.ProlongDuration) // prolong expire of terms
	}

	if m.PriceAutoUpdate != 0 {
		m.Price += m.PriceAutoUpdate // up the price
	}

	if m.QoSAutoUpdate != nil {
		if m.QoSAutoUpdate.UploadMbps != 0 && m.QoS.UploadMbps > m.QoSAutoUpdate.UploadMbps {
			m.QoS.UploadMbps -= m.QoSAutoUpdate.UploadMbps // down the qos of upload mbps
		}

		if m.QoSAutoUpdate.DownloadMbps != 0 && m.QoS.DownloadMbps > m.QoSAutoUpdate.DownloadMbps {
			m.QoS.DownloadMbps -= m.QoSAutoUpdate.DownloadMbps // down the qos of download mbps
		}
	}

	return m
}

// Validate checks Terms for correctness.
// If it is not return errInvalidTerms.
func (m *Terms) Validate() (err error) {
	switch { // is invalid
	case m.QoS == nil:
		err = errors.New(ErrCodeBadRequest, "invalid terms qos")

	case m.QoS.UploadMbps <= 0:
		err = errors.New(ErrCodeBadRequest, "invalid terms qos upload mbps")

	case m.QoS.DownloadMbps <= 0:
		err = errors.New(ErrCodeBadRequest, "invalid terms qos download mbps")

	case m.Expired():
		now := time.NowTime().Add(TermsExpiredDuration).Format(time.RFC3339)
		err = errors.New(ErrCodeBadRequest, "expired at must be after "+now)

	default:
		return nil // is valid
	}

	return ErrInvalidTerms.Wrap(err)
}
