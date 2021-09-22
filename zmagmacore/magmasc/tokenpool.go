package magmasc

import (
	"encoding/json"

	"github.com/0chain/gosdk/core/util"
)

type (
	// TokenPool represents token pool implementation.
	TokenPool struct {
		ID        string              `json:"id"`
		Balance   int64               `json:"balance"`
		HolderID  string              `json:"holder_id"`
		PayerID   string              `json:"payer_id"`
		PayeeID   string              `json:"payee_id"`
		Transfers []TokenPoolTransfer `json:"transfers,omitempty"`
	}
)

var (
	// Make sure tokenPool implements Serializable interface.
	_ util.Serializable = (*TokenPool)(nil)
)

// Decode implements util.Serializable interface.
func (m *TokenPool) Decode(blob []byte) error {
	var pool TokenPool
	if err := json.Unmarshal(blob, &pool); err != nil {
		return errDecodeData.Wrap(err)
	}

	m.ID = pool.ID
	m.Balance = pool.Balance
	m.HolderID = pool.HolderID
	m.PayerID = pool.PayerID
	m.PayeeID = pool.PayeeID
	m.Transfers = pool.Transfers

	return nil
}

// Encode implements util.Serializable interface.
func (m *TokenPool) Encode() []byte {
	blob, _ := json.Marshal(m)
	return blob
}
