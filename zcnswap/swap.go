package zcnswap

import (
	"github.com/0chain/errors"
	l "github.com/0chain/gosdk/zboxcore/logger"
	contracts "github.com/0chain/gosdk/zcnswap/ethereum/bancor"
	"github.com/0chain/gosdk/zcnswap/swapfactory"
	"github.com/0chain/gosdk/zcnswap/swapfactory/erc20"
	ethutils "github.com/0chain/gosdk/zcnswap/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
	"math/big"
)

//func (s *swapService) PackConvert(path *[]cmn.Address, amount, minReturn *big.Int, affiliate cmn.Address) ([]byte, error) {
//	abi, err := contracts.IBancorNetworkMetaData.GetAbi()
//	if err != nil {
//		return nil, errors.New("500", "No ABI")
//	}
//
//	// ropsten test network deployed old contract, hence using old method
//	//pack, err := abi.Pack("convertByPath2", path, amount, minReturn, affiliate)
//	pack, err := abi.Pack("convertByPath", path, amount, minReturn, affiliate, affiliate, big.NewInt(0))
//	if err != nil {
//		return nil, errors.New("500", "Unable to pack")
//	}
//
//	return pack, nil
//}

func (s *SwapClient) Swap(swapAmount int64, sourceTokenAddress, targetTokenAddress string) (string, error) {
	fromAddress := common.HexToAddress(ethAccount.SourceAddr.Hex())
	amount := new(big.Int).SetInt64(swapAmount)

	// checking for available funds
	balance, err := erc20.TokenBalance(fromAddress,
		common.HexToAddress(targetTokenAddress),
		s.ethereumClient)
	if err != nil {
		return "", err
	}
	if balance.Cmp(amount) == -1 {
		return "", errors.New("500", "Not enough balance")
	}

	//bancorService := bancor.NewSwapService(client, zcncore.GetClientWalletKey())

	pair, err := s.EstimateRate(sourceTokenAddress, targetTokenAddress, amount)
	if err != nil {
		return "", err
	}

	spender := common.HexToAddress(s.BancorAddress)
	fromToken := common.HexToAddress(s.UsdcTokenAddress)

	bancorModule, err := contracts.NewIBancorNetwork(spender, s.ethereumClient)
	if err != nil {
		return "", err
	}

	balance, err := erc20.TokenBalance(fromAddress, fromToken, s.ethereumClient)
	if err != nil {
		return "", err
	}

	if balance.Cmp(pair.AmountIn) != 1 {
		return "", errors.New("500", "Not enough balance")
	}

	tokenInfo, err := erc20.TokenInfo(fromToken, s.ethereumClient)
	if err != nil {
		return "", err
	}

	// allowance
	aCheckAllowanceResult, err := erc20.CheckAllowance(fromToken,
		spender,
		fromAddress,
		pair.AmountIn,
		tokenInfo.TokenSymbol == "ETH",
		s.ethereumClient)
	if err != nil {
		l.Logger.Error("Allowance error", zap.Error(err))
		return "", err
	}

	affiliateAccount := common.HexToAddress(AFFILIATE_ACCOUNT)

	convertPayload, err := s.PackConvert(pair.ConversionPath, pair.AmountIn, pair.AmountOut, affiliateAccount)
	if err != nil {
		l.Logger.Error("error in PackConvert", zap.Error(err))
		return "", err
	}

	if !aCheckAllowanceResult.IsSatisfied {
		err := s.Approve(spender, fromToken, fromAddress, pair.AmountIn)
		if err != nil {
			l.Logger.Error("Approve error", zap.Error(err))
			return "", err
		}
	}

	value := big.NewInt(0)
	opts, err := ethutils.NewSignedTransaction(convertPayload, from, spender.Hex(), value, s.PrivateKey, s.ethereumClient)
	if err != nil {
		l.Logger.Error("Signed transaction errorr", zap.Error(err))
		return "", err
	}

	signedTx, err := bancorModule.ConvertByPath(opts, *pair.ConversionPath, pair.AmountIn, pair.AmountOut, affiliateAccount, affiliateAccount, big.NewInt(0))
	if err != nil {
		return "", err
	}

	return signedTx.Hash().String(), nil
}

func (s *SwapClient) SwapWithConversionPath(pair *swapfactory.ExchangePair, from string) (*types.Transaction, error) {
	//var affiliateAccount = cmn.HexToAddress("0x0000000000000000000000000000000000000000")
	//
	//fromAddress := cmn.HexToAddress(from)
	//spender := cmn.HexToAddress(config.Configuration.BancorAddress)
	//fromToken := cmn.HexToAddress(config.Configuration.UsdcTokenAddress)
	//
	//bancorModule, err := contracts.NewIBancorNetwork(spender, s.Client)
	//if err != nil {
	//	return nil, err
	//}
	//
	//balance, err := erc20.TokenBalance(fromAddress, fromToken, s.Client)
	//if err != nil {
	//	return nil, err
	//}
	//
	//if balance.Cmp(pair.AmountIn) != 1 {
	//	return nil, errors.New("500", "Not enough balance")
	//}
	//
	//tokenInfo, err := erc20.TokenInfo(fromToken, s.Client)
	//if err != nil {
	//	return nil, err
	//}
	//
	//// allowance
	//aCheckAllowanceResult, err := erc20.CheckAllowance(fromToken,
	//	spender,
	//	fromAddress,
	//	pair.AmountIn,
	//	tokenInfo.TokenSymbol == "ETH",
	//	s.Client)
	//if err != nil {
	//	l.Logger.Error("Allowance error", zap.Error(err))
	//	return nil, err
	//}
	//
	//convertPayload, err := s.PackConvert(pair.ConversionPath, pair.AmountIn, pair.AmountOut, affiliateAccount)
	//if err != nil {
	//	l.Logger.Error("error in PackConvert", zap.Error(err))
	//	return nil, err
	//}
	//
	//if !aCheckAllowanceResult.IsSatisfied {
	//	err := s.Approve(spender, fromToken, fromAddress, pair.AmountIn)
	//	if err != nil {
	//		l.Logger.Error("Approve error", zap.Error(err))
	//		return nil, err
	//	}
	//}
	//
	//value := big.NewInt(0)
	//opts, err := ethutils.NewSignedTransaction(convertPayload, from, spender.Hex(), value, s.PrivateKey, s.Client)
	//if err != nil {
	//	l.Logger.Error("Signed transaction errorr", zap.Error(err))
	//	return nil, err
	//}
	//
	//// ropsten test network deployed old contract, hence using old method
	////return bancorModule.ConvertByPath2(opts, *path, amountIn, amountOut, affiliateAccount)
	//return bancorModule.ConvertByPath(opts, *pair.ConversionPath, pair.AmountIn, pair.AmountOut, affiliateAccount, affiliateAccount, big.NewInt(0))
}

func (s *SwapClient) GetAllowance(tokenAddr, spender, userAddr common.Address, client *ethclient.Client) (*big.Int, error) {
	erc20Module, err := NewERC20(tokenAddr, client)
	if err != nil {
		return nil, err
	}

	return erc20Module.Allowance(&bind.CallOpts{}, userAddr, spender)
}
