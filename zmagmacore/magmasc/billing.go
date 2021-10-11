package magmasc

import (
	"encoding/json"

	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zmagmacore/errors"
	"github.com/0chain/gosdk/zmagmacore/magmasc/pb"
	"github.com/0chain/gosdk/zmagmacore/time"
)

type (
	// Billing represents all info about data usage.
	Billing struct {
		Amount      int64          `json:"amount"`
		DataMarker  *DataMarker    `json:"data_marker"`
		CompletedAt time.Timestamp `json:"completed_at,omitempty"`
	}
)

var (
	// Make sure Billing implements Serializable interface.
	_ util.Serializable = (*Billing)(nil)
)

// CalcAmount calculates and sets the billing Amount value by given price.
// NOTE: the cost value must be represented in token units per megabyte.
func (m *Billing) CalcAmount(terms Terms) {
	price := float64(terms.GetPrice())
	if price > 0 && m.DataMarker != nil && m.DataMarker.DataUsage != nil {
		// data usage summary in megabytes
		mbps := float64(m.DataMarker.DataUsage.DownloadBytes+m.DataMarker.DataUsage.UploadBytes) / million
		m.Amount = int64(mbps * price) // rounded amount of megabytes multiplied by price
	}
	if minCost := terms.GetMinCost(); m.Amount < minCost {
		m.Amount = minCost
	}
}

// Decode implements util.Serializable interface.
func (m *Billing) Decode(blob []byte) error {
	var bill Billing
	if err := json.Unmarshal(blob, &bill); err != nil {
		return ErrDecodeData.Wrap(err)
	}

	m.Amount = bill.Amount
	m.DataMarker = bill.DataMarker
	m.CompletedAt = bill.CompletedAt

	return nil
}

// Encode implements util.Serializable interface.
func (m *Billing) Encode() []byte {
	blob, _ := json.Marshal(m)
	return blob
}

// Validate checks given data usage is correctness for the billing.
func (m *Billing) Validate(dataUsage *pb.DataUsage) (err error) {
	isBillDataUsageNil := m.DataMarker == nil || m.DataMarker.DataUsage == nil

	switch {
	case dataUsage == nil:
		err = errors.New(ErrCodeBadRequest, "data usage required")

	case !isBillDataUsageNil && m.DataMarker.DataUsage.SessionTime > dataUsage.SessionTime:
		err = errors.New(ErrCodeBadRequest, "invalid session time")

	case !isBillDataUsageNil && m.DataMarker.DataUsage.UploadBytes > dataUsage.UploadBytes:
		err = errors.New(ErrCodeBadRequest, "invalid upload bytes")

	case !isBillDataUsageNil && m.DataMarker.DataUsage.DownloadBytes > dataUsage.DownloadBytes:
		err = errors.New(ErrCodeBadRequest, "invalid download bytes")

	default:
		return nil // is valid - everything is ok
	}

	return ErrInvalidDataUsage.Wrap(err)
}
