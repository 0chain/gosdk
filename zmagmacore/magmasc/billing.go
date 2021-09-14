package magmasc

import (
	"encoding/json"

	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zmagmacore/errors"
	"github.com/0chain/gosdk/zmagmacore/time"
)

type (
	// Billing represents all info about data usage.
	Billing struct {
		Amount      int64          `json:"amount"`
		DataUsage   DataUsage      `json:"data_usage"`
		CompletedAt time.Timestamp `json:"completed_at,omitempty"`
	}
)

var (
	// Make sure Billing implements Serializable interface.
	_ util.Serializable = (*Billing)(nil)
)

// CalcAmount calculates and sets the billing Amount value by given price.
// NOTE: the cost value must be represented in token units per megabyte.
func (m *Billing) CalcAmount(terms ProviderTerms) {
	price := float64(terms.GetPrice())
	if price > 0 {
		// data usage summary in megabytes
		mbps := float64(m.DataUsage.UploadBytes+m.DataUsage.DownloadBytes) / million
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
		return errDecodeData.Wrap(err)
	}

	m.Amount = bill.Amount
	m.DataUsage = bill.DataUsage
	m.CompletedAt = bill.CompletedAt

	return nil
}

// Encode implements util.Serializable interface.
func (m *Billing) Encode() []byte {
	blob, _ := json.Marshal(m)
	return blob
}

// Validate checks given data usage is correctness for the billing.
func (m *Billing) Validate(dataUsage *DataUsage) (err error) {
	switch {
	case dataUsage == nil:
		err = errors.New(errCodeBadRequest, "data usage required")

	case m.DataUsage.SessionTime > dataUsage.SessionTime:
		err = errors.New(errCodeBadRequest, "invalid session time")

	case m.DataUsage.UploadBytes > dataUsage.UploadBytes:
		err = errors.New(errCodeBadRequest, "invalid upload bytes")

	case m.DataUsage.DownloadBytes > dataUsage.DownloadBytes:
		err = errors.New(errCodeBadRequest, "invalid download bytes")

	default:
		return nil // is valid - everything is ok
	}

	return errInvalidDataUsage.Wrap(err)
}
