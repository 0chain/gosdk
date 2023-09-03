package erc20

import (
	"github.com/0chain/gosdk/zcnbridge/ethereum/erc20"
	"math/big"

	cmn "github.com/0chain/errors"
	"github.com/ethereum/go-ethereum/common"
)

// PackApprove - pack approve payload from ABI contract
func PackApprove(spender common.Address, amount *big.Int) ([]byte, error) {
	abi, err := erc20.ERC20MetaData.GetAbi()
	if err != nil {
		return nil, cmn.New("500", "No ABI")
	}

	pack, err := abi.Pack("approve", spender, amount)
	if err != nil {
		return nil, cmn.New("500", "Unable to pack")
	}

	return pack, nil
}
