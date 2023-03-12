package erc20

import (
	"github.com/0chain/gosdk/zcnswap/contracts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func GetAllowance(tokenAddr, spender, userAddr common.Address, client *ethclient.Client) (*big.Int, error) {
	erc20Module, err := contracts.NewERC20(tokenAddr, client)
	if err != nil {
		return nil, err
	}

	return erc20Module.Allowance(&bind.CallOpts{}, userAddr, spender)
}

// func (fromToken, data.Bancor, userAddr, amount) (Allowance, AllowanceSatisfied, AllowanceData, error)
type CheckAllowanceResult struct {
	AllowanceAmount *big.Int `json:"allowanceAmount"`
	IsSatisfied     bool     `json:"isSatisfied"`
	AllowanceData   []byte   `json:"allowanceData"`
}

// CheckAllowance Cannot use this to check ETH
func CheckAllowance(fromToken, spender, userAddr common.Address, amount *big.Int, fromIsETH bool, client *ethclient.Client) (*CheckAllowanceResult, error) {
	if fromIsETH {
		return &CheckAllowanceResult{
			AllowanceAmount: amount,
			IsSatisfied:     true,
			AllowanceData:   []byte(""),
		}, nil
	}

	fromTokenAllowance, err := GetAllowance(fromToken, spender, userAddr, client)
	if err != nil {
		return nil, err
	}

	return &CheckAllowanceResult{
		AllowanceAmount: amount,
		IsSatisfied:     fromTokenAllowance.Cmp(amount) >= 0,
		//AllowanceData:   callData,
	}, nil
}
