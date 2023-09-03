package swapfactory

import (
	cmn "github.com/ethereum/go-ethereum/common"
	"math/big"
)

type ExchangePair struct {
	ContractName   string         `json:"contract_name,omitempty"`
	AmountIn       *big.Int       `json:"amount_in,omitempty"`
	AmountOut      *big.Int       `json:"amount_out,omitempty"`
	ExchangeRatio  *big.Int       `json:"exchange_ratio,omitempty"`
	TxFee          *big.Int       `json:"tx_fee,omitempty"`
	SupportSwap    bool           `json:"support_swap,omitempty"`
	ConversionPath *[]cmn.Address `json:"coversion_path,omitempty"`
}
