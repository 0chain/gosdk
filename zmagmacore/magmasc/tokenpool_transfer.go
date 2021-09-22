package magmasc

import (
	"encoding/json"

	"github.com/0chain/gosdk/core/util"
)

type (
	// TokenPoolTransfer stores info about token transfers from pool to pool.
	TokenPoolTransfer struct {
		TxnHash    string `json:"txn_hash,omitempty"`
		FromPool   string `json:"from_pool,omitempty"`
		ToPool     string `json:"to_pool,omitempty"`
		Value      int64  `json:"value,omitempty"`
		FromClient string `json:"from_client,omitempty"`
		ToClient   string `json:"to_client,omitempty"`
	}
)

var (
	// Make sure TokenPoolTransfer implements Serializable interface.
	_ util.Serializable = (*TokenPoolTransfer)(nil)
)

// Decode implements util.Serializable interface.
func (m *TokenPoolTransfer) Decode(blob []byte) error {
	return json.Unmarshal(blob, m)
}

// Encode implements util.Serializable interface.
func (m *TokenPoolTransfer) Encode() []byte {
	blob, _ := json.Marshal(m)
	return blob
}
