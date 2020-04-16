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
	RoundRandomSeed       int64  `json:"round_random_seed,omitempy"`
	MerkleTreeRoot        string `json:"merkle_tree_root,omitempty"`
	StateHash             string `json:"state_hash,omitempty"`
	ReceiptMerkleTreeRoot string `json:"receipt_merkle_tree_root,omitempty"`
	NumTxns               int64  `json:"num_txns,omitempty"`
}

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
