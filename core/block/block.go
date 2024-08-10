// Provides the data structures and methods to work with the block data structure.
// The block data structure is the core data structure in the 0chain protocol.
// It is used to store the transactions and the state of the system at a given point in time.
// The block data structure is used to create the blockchain, which is a chain of blocks that are linked together using the hash of the previous block.
package block

import (
	"fmt"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/transaction"
)

type Key []byte

type Header struct {
	Version               string `json:"version,omitempty"`
	CreationDate          int64  `json:"creation_date,omitempty"`
	Hash                  string `json:"hash,omitempty"`
	MinerID               string `json:"miner_id,omitempty"`
	Round                 int64  `json:"round,omitempty"`
	RoundRandomSeed       int64  `json:"round_random_seed,omitempty"`
	MerkleTreeRoot        string `json:"merkle_tree_root,omitempty"`
	StateHash             string `json:"state_hash,omitempty"`
	ReceiptMerkleTreeRoot string `json:"receipt_merkle_tree_root,omitempty"`
	NumTxns               int64  `json:"num_txns,omitempty"`
}

// IsBlockExtends - check if the block extends the previous block
//   - prevHash is the hash of the previous block
func (h *Header) IsBlockExtends(prevHash string) bool {
	var data = fmt.Sprintf("%s:%s:%d:%d:%d:%s:%s", h.MinerID, prevHash,
		h.CreationDate, h.Round, h.RoundRandomSeed, h.MerkleTreeRoot,
		h.ReceiptMerkleTreeRoot)
	return encryption.Hash(data) == h.Hash
}

/*Block - data structure that holds the block data */
type Block struct {
	Header *Header `json:"-"`

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

type ChainStats struct {
	BlockSize            int     `json:"block_size"`
	Count                int     `json:"count"`
	CurrentRound         int     `json:"current_round"`
	Delta                int     `json:"delta"`
	LatestFinalizedRound int     `json:"latest_finalized_round"`
	Max                  float64 `json:"max"`
	Mean                 float64 `json:"mean"`
	Min                  float64 `json:"min"`
	Percentile50         float64 `json:"percentile_50"`
	Percentile90         float64 `json:"percentile_90"`
	Percentile95         float64 `json:"percentile_95"`
	Percentile99         float64 `json:"percentile_99"`
	Rate15Min            float64 `json:"rate_15_min"`
	Rate1Min             float64 `json:"rate_1_min"`
	Rate5Min             float64 `json:"rate_5_min"`
	RateMean             float64 `json:"rate_mean"`
	StdDev               float64 `json:"std_dev"`
	TotalTxns            int     `json:"total_txns"`
}

type FeeStats struct {
	MaxFees  common.Balance `json:"max_fees"`
	MinFees  common.Balance `json:"min_fees"`
	MeanFees common.Balance `json:"mean_fees"`
}
