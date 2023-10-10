package swapfactory

import (
	"encoding/json"
	"math/big"
	"strings"

	cmn "github.com/ethereum/go-ethereum/common"

	"github.com/labstack/echo"
)

type EchoContext struct {
	echo.Context
}

type Token struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
	LogoURI  string `json:"logoURI"`
}

type Tokens []Token

func (t Tokens) Len() int { return len(t) }
func (t Tokens) Less(i, j int) bool {
	cmp := strings.Compare(t[i].Symbol, t[j].Symbol)
	return cmp == -1
}
func (t Tokens) Swap(i, j int) { t[i], t[j] = t[j], t[i] }

type Exchange struct {
	Name       string `json:"name"`
	APIAddress string `json:"api_address"`
	Decimals   int    `json:"decimals"`
}

type ExchangeResult struct {
	FromName      string           `json:"from_name,omitempty"`
	ToName        string           `json:"to_name,omitempty"`
	FromAddr      string           `json:"from_addr,omitempty"`
	ToAddr        string           `json:"to_addr,omitempty"`
	ExchangePairs ExchangePairList `json:"exchange_pairs,omitempty"`
}

type ExchangePair struct {
	ContractName   string         `json:"contract_name,omitempty"`
	AmountIn       *big.Int       `json:"amount_in,omitempty"`
	AmountOut      *big.Int       `json:"amount_out,omitempty"`
	ExchangeRatio  *big.Int       `json:"exchange_ratio,omitempty"`
	TxFee          *big.Int       `json:"tx_fee,omitempty"`
	SupportSwap    bool           `json:"support_swap,omitempty"`
	ConversionPath *[]cmn.Address `json:"coversion_path,omitempty"`
}

func (e *ExchangePair) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ContractName  string `json:"contract_name"`
		AmountIn      string `json:"amount_in"`
		AmountOut     string `json:"amount_out"`
		ExchangeRatio string `json:"exchange_ratio"`
		TxFee         string `json:"tx_fee"`
		SupportSwap   bool   `json:"support_swap"`
	}{
		ContractName:  e.ContractName,
		AmountIn:      e.AmountIn.String(),
		AmountOut:     e.AmountOut.String(),
		ExchangeRatio: e.ExchangeRatio.String(),
		TxFee:         e.TxFee.String(),
		SupportSwap:   e.SupportSwap,
	})
}

type ExchangePairList []ExchangePair

func (p ExchangePairList) Len() int { return len(p) }
func (p ExchangePairList) Less(i, j int) bool {

	return p[i].AmountOut.Cmp(p[j].AmountOut) == 1
}
func (p ExchangePairList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

type PoolToken struct {
	Address      string `json:"address"`
	Name         string `json:"name"`
	Symbol       string `json:"symbol"`
	Logo         string `json:"logo,omitempty"`
	Decimals     int    `json:"decimals,omitempty"`
	DenormWeight string `json:"denormWeight,omitempty"`
	Balance      string `json:"balance,omitempty"`
}

type PoolInfo struct {
	Address     string      `json:"address,omitempty"`
	Platform    string      `json:"platform,omitempty"`
	Liquidity   string      `json:"liquidity,omitempty"`
	Reserves    []string    `json:"reserves,omitempty"`
	TokenPrices []string    `json:"tokenprices,omitempty"`
	Volumes     []string    `json:"volumes,omitempty"`
	ReserveUSD  string      `json:"reserveUSD,omitempty"`
	ReserveETH  string      `json:"reserveETH,omitempty"`
	TotalSupply string      `json:"totalSupply,omitempty"`
	VolumeUSD   string      `json:"volumeUSD,omitempty"`
	Tokens      []PoolToken `json:"tokens,omitempty"`
	SwapFee     string      `json:"swapFee,omitempty"`
	TotalWeight string      `json:"totalWeight,omitempty"`
}

type SwapTx struct {
	Data               string `json:"data"`
	TxFee              string `json:"tx_fee"`
	ContractAddr       string `json:"contract_addr"`
	FromTokenAmount    string `json:"from_token_amount"`
	ToTokenAmount      string `json:"to_token_amount"`
	ExchangeRatio      string `json:"exchange_ratio"`
	FromTokenAddr      string `json:"from_token_addr"`
	Allowance          string `json:"allowance"`
	AllowanceSatisfied bool   `json:"allowance_satisfied"`
	AllowanceData      string `json:"allowance_data"`
}
