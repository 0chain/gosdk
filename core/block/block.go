package block

import (
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/transaction"
)

type Key []byte

/*Block - data structure that holds the block data */
type Block struct {
	MinerID           common.Key `json:"miner_id"`
	Round             int64      `json:"round"`
	RoundRandomSeed   int64      `json:"round_random_seed"`
	RoundTimeoutCount int        `json:"round_timeout_count"`

	Hash            common.Key `json:"hash"`
	Signature       string     `json:"signature"`
	ChainID         common.Key `json:"chain_id"`
	ChainWeight     float64    `json:"chain_weight"`
	RunningTxnCount int64      `json:"running_txn_count"`

	Version      string           `json:"version"`
	CreationDate common.Timestamp `json:"creation_date"`

	MagicBlockHash string `json:"magic_block_hash"`
	PrevHash       string `json:"prev_hash"`

	ClientStateHash Key                        `json:"state_hash"`
	Txns            []*transaction.Transaction `json:"transactions,omitempty"`

	// muted

	// VerificationTickets []*VerificationTicket `json:"verification_tickets,omitempty"`
	// PrevBlockVerificationTickets []*VerificationTicket `json:"prev_verification_tickets,omitempty"`
}
