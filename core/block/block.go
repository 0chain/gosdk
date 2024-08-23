// Provides the data structures and methods to work with the block data structure.
// The block data structure is the core data structure in the 0chain protocol.
// It is used to store the transactions and the state of the system at a given point in time.
// The block data structure is used to create the blockchain, which is a chain of blocks that are linked together using the hash of the previous block.
package block

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/core/node"
	"github.com/0chain/gosdk/core/util"
	"net/http"

	"github.com/0chain/errors"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/transaction"
)

const GET_BLOCK_INFO = `/v1/block/get?`

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

func GetBlockByRound(h *node.NodeHolder, ctx context.Context, numSharders int, round int64) (b *Block, err error) {

	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	numSharders = len(h.Healthy()) // overwrite, use all
	h.QueryFromShardersContext(ctx, numSharders,
		fmt.Sprintf("%sround=%d&content=full,header", GET_BLOCK_INFO, round),
		result)

	var (
		maxConsensus   int
		roundConsensus = make(map[string]int)
	)

	type respObj struct {
		Block  *Block  `json:"block"`
		Header *Header `json:"header"`
	}

	for i := 0; i < numSharders; i++ {
		var rsp = <-result
		if rsp == nil {
			logger.Log.Error("nil response")
			continue
		}
		logger.Log.Debug(rsp.Url, rsp.Status)

		if rsp.StatusCode != http.StatusOK {
			logger.Log.Error(rsp.Body)
			continue
		}

		var respo respObj
		if err = json.Unmarshal([]byte(rsp.Body), &respo); err != nil {
			logger.Log.Error("block parse error: ", err)
			err = nil
			continue
		}

		if respo.Block == nil {
			logger.Log.Debug(rsp.Url, "no block in response:", rsp.Body)
			continue
		}

		if respo.Header == nil {
			logger.Log.Debug(rsp.Url, "no block header in response:", rsp.Body)
			continue
		}

		if respo.Header.Hash != string(respo.Block.Hash) {
			logger.Log.Debug(rsp.Url, "header and block hash mismatch:", rsp.Body)
			continue
		}

		b = respo.Block
		b.Header = respo.Header

		var h = encryption.FastHash([]byte(b.Hash))
		if roundConsensus[h]++; roundConsensus[h] > maxConsensus {
			maxConsensus = roundConsensus[h]
		}
	}

	if maxConsensus == 0 {
		return nil, errors.New("", "round info not found")
	}

	return
}
