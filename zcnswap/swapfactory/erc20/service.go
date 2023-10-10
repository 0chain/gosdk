package erc20

import (
	"math/big"

	"github.com/0chain/gosdk/zcnbridge/ethereum/erc20"
	"github.com/0chain/gosdk/zcnswap/contracts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"

	cmn "github.com/0chain/errors"
	"github.com/ethereum/go-ethereum/common"
)

var erc20Info = map[common.Address]Info{}

type Info struct {
	TokenAddr   common.Address
	TokenName   string
	TokenSymbol string
	Decimals    uint8
}

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

// TokenBalance - return balance by selected token contract
func TokenBalance(userAddr, tokenAddr common.Address, client *ethclient.Client) (*big.Int, error) {
	erc20Module, err := contracts.NewERC20(tokenAddr, client)
	if err != nil {
		return nil, err
	}
	return erc20Module.BalanceOf(&bind.CallOpts{}, userAddr)
}

func TokenInfo(tokenAddr common.Address, client *ethclient.Client) (Info, error) {
	ret := Info{}
	if out, ok := erc20Info[tokenAddr]; ok {
		return out, nil
	}

	erc20Module, err := contracts.NewERC20(tokenAddr, client)
	if err != nil {
		return ret, err
	}
	decimals, err := erc20Module.Decimals(nil)
	if err != nil {
		return ret, err
	}
	tokenName, err := erc20Module.Symbol(nil)
	if err != nil {
		return ret, err
	}
	tokenSymbol, err := erc20Module.Name(nil)
	if err != nil {
		return ret, err
	}

	ret.TokenAddr = tokenAddr
	ret.TokenName = tokenName
	ret.TokenSymbol = tokenSymbol
	ret.Decimals = decimals

	erc20Info[tokenAddr] = ret
	return ret, nil
}
