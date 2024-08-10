package zcnbridge

import (
	"fmt"
	"math/big"
	"time"

	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zcncore"
	"github.com/pkg/errors"
)

// AlchemyGasEstimationRequest describes request used for Alchemy enhanced JSON-RPC API.
type AlchemyGasEstimationRequest struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value string `json:"value"`
	Data  string `json:"data"`
}

// GasEstimationRequest describes request used for Alchemy enhanced JSON-RPC API.
type GasEstimationRequest struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value string `json:"value"`
}

// GasPriceEstimationResult represents result of the gas price estimation operation execution.
type GasPriceEstimationResult struct {
	Value float64 `json:"value"`
}

// BancorTokenDetails describes Bancor ZCN zcntoken pool details
type BancorTokenDetails struct {
	Data struct {
		Rate struct {
			ETH  string `json:"eth"`
			BNT  string `json:"bnt"`
			USDC string `json:"usd"`
			EURC string `json:"eur"`
		}
	} `json:"data"`
}

func GetTransactionStatus(hash string) (int, error) {
	_, err := zcncore.GetEthClient()
	if err != nil {
		return -1, err
	}

	return zcncore.CheckEthHashStatus(hash), nil
}

// ConfirmEthereumTransaction confirms Ethereum transaction by hash.
//   - hash is the transaction hash to confirm.
//   - times is the number of times to try confirming the transaction.
//   - duration is the duration to wait between each confirmation attempt.
func ConfirmEthereumTransaction(hash string, times int, duration time.Duration) (int, error) {
	var (
		res = 0
		err error
	)

	if hash == "" {
		return -1, errors.New("transaction hash should not be empty")
	}

	for i := 0; i < times; i++ {
		res, err = GetTransactionStatus(hash)
		if err != nil {
			Logger.Info(fmt.Sprintf("confirmation of Ethereum transaction %s [ERROR]", hash))
			return -1, err
		}
		if res == 1 {
			Logger.Info(fmt.Sprintf("confirmation of Ethereum transaction %s [OK]", hash))
			return res, nil
		}
		if res == 0 {
			Logger.Info(fmt.Sprintf("confirmation of Ethereum transaction %s [FAILED]", hash))
			return res, nil
		}
		Logger.Info(fmt.Sprintf("Try confirming Ethereum transaction %s # %d", hash, i))
		sys.Sleep(duration)
	}

	Logger.Info(fmt.Sprintf("Verification of transaction %s is still pending after %d efforts, try checking it later", hash, times))

	return res, nil
}

func addPercents(gasLimitUnits uint64, percents int) *big.Int {
	gasLimitBig := big.NewInt(int64(gasLimitUnits))
	factorBig := big.NewInt(int64(percents))
	deltaBig := gasLimitBig.Div(gasLimitBig, factorBig)

	origin := big.NewInt(int64(gasLimitUnits))
	gasLimitBig = origin.Add(origin, deltaBig)

	return gasLimitBig
}

// ConvertIntToHex converts given int value to hex string.
func ConvertIntToHex(value int64) string {
	return fmt.Sprintf("%#x", value)
}
