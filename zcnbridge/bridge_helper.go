package zcnbridge

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/0chain/gosdk/zcnbridge/log"
	"github.com/0chain/gosdk/zcncore"
	"github.com/pkg/errors"
)

func GetTransactionStatus(hash string) (int, error) {
	_, err := zcncore.GetEthClient()
	if err != nil {
		return -1, err
	}

	return zcncore.CheckEthHashStatus(hash), nil
}

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
			return -1, err
		}
		if res == 1 || res == 0 {
			break
		}
		log.Logger.Info(fmt.Sprintf("try # %d", i))
		time.Sleep(duration)
	}
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

func CreateHash(message string) common.Hash {
	data := []byte(message)
	hash := crypto.Keccak256Hash(data)

	return hash
}
