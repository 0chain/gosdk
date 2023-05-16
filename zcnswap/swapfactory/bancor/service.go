package bancor

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"time"

	"github.com/0chain/errors"
	l "github.com/0chain/gosdk/zboxcore/logger"
	contractErc20 "github.com/0chain/gosdk/zcnbridge/ethereum/erc20"
	"github.com/0chain/gosdk/zcnswap/config"
	"github.com/0chain/gosdk/zcnswap/contracts"
	"github.com/0chain/gosdk/zcnswap/swapfactory"
	"github.com/0chain/gosdk/zcnswap/swapfactory/erc20"
	ethutils "github.com/0chain/gosdk/zcnswap/utils"
	cmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

type ISwapService interface {
	SwapWithConversionPath(pair *swapfactory.ExchangePair, from string) (*types.Transaction, error)
	Approve(spender, fromToken, fromAddress cmn.Address, amountIn *big.Int) error
	IncreaseAllowance(spender, fromToken, fromAddress cmn.Address, addedValue *big.Int) error
	PackConvert(path *[]cmn.Address, amount, minReturn *big.Int, affiliate cmn.Address) ([]byte, error)
	EstimateRate(from, to string, amount *big.Int) (*swapfactory.ExchangePair, error)
}

type swapService struct {
	Client     *ethclient.Client
	PrivateKey *ecdsa.PrivateKey
}

// func SwapWithConversionPath(path *[]cmn.Address, from string, amountIn, amountOut *big.Int, privateKey *ecdsa.PrivateKey, client *ethclient.Client) (*types.Transaction, error) {
func (s *swapService) SwapWithConversionPath(pair *swapfactory.ExchangePair, from string) (*types.Transaction, error) {
	var affiliateAccount = cmn.HexToAddress("0x0000000000000000000000000000000000000000")

	fromAddress := cmn.HexToAddress(from)
	spender := cmn.HexToAddress(config.Configuration.BancorAddress)
	fromToken := cmn.HexToAddress(config.Configuration.UsdcTokenAddress)

	bancorModule, err := contracts.NewIBancorNetwork(spender, s.Client)
	if err != nil {
		return nil, err
	}

	balance, err := erc20.TokenBalance(fromAddress, fromToken, s.Client)
	if err != nil {
		return nil, err
	}

	if balance.Cmp(pair.AmountIn) != 1 {
		return nil, errors.New("500", "Not enough balance")
	}

	tokenInfo, err := erc20.TokenInfo(fromToken, s.Client)
	if err != nil {
		return nil, err
	}

	// allowance
	aCheckAllowanceResult, err := erc20.CheckAllowance(fromToken,
		spender,
		fromAddress,
		pair.AmountIn,
		tokenInfo.TokenSymbol == "ETH",
		s.Client)
	if err != nil {
		l.Logger.Error("Allowance error", zap.Error(err))
		return nil, err
	}

	convertPayload, err := s.PackConvert(pair.ConversionPath, pair.AmountIn, pair.AmountOut, affiliateAccount)
	if err != nil {
		l.Logger.Error("error in PackConvert", zap.Error(err))
		return nil, err
	}

	if !aCheckAllowanceResult.IsSatisfied {
		err := s.Approve(spender, fromToken, fromAddress, pair.AmountIn)
		if err != nil {
			l.Logger.Error("Approve error", zap.Error(err))
			return nil, err
		}
	}

	value := big.NewInt(0)
	opts, err := ethutils.NewSignedTransaction(convertPayload, from, spender.Hex(), value, s.PrivateKey, s.Client)
	if err != nil {
		l.Logger.Error("Signed transaction errorr", zap.Error(err))
		return nil, err
	}

	// ropsten test network deployed old contract, hence using old method
	//return bancorModule.ConvertByPath2(opts, *path, amountIn, amountOut, affiliateAccount)
	return bancorModule.ConvertByPath(opts, *pair.ConversionPath, pair.AmountIn, pair.AmountOut, affiliateAccount, affiliateAccount, big.NewInt(0))
}

func (s *swapService) Approve(spender, fromToken, fromAddress cmn.Address, amountIn *big.Int) error {
	l.Logger.Info("Approve called", zap.Any("fromToken", fromToken.Hex()))

	ercModule, err := contracts.NewERC20(fromToken, s.Client)
	if err != nil {
		return err
	}

	approvePayload, err := erc20.PackApprove(spender, amountIn)
	if err != nil {
		return err
	}

	value := big.NewInt(0)
	opts, err := ethutils.NewSignedTransaction(approvePayload, fromAddress.Hex(), spender.Hex(), value, s.PrivateKey, s.Client)
	if err != nil {
		return err
	}

	trans, err := ercModule.Approve(opts, spender, amountIn)
	if err != nil {
		return err
	}

	l.Logger.Info("Approve tx hash", zap.Any("hash", trans.Hash().Hex()))

	res, err := ethutils.ConfirmEthereumTransaction(trans.Hash().Hex(), 60, time.Minute, s.Client)
	if err != nil {
		return err
	}
	if res == ethutils.STATUS_FAIL {
		return errors.New("500", "Unable to confirm transaction")
	}

	return nil
}

func (s *swapService) IncreaseAllowance(spender, fromToken, fromAddress cmn.Address, addedValue *big.Int) error {
	l.Logger.Info("IncreaseAllowance called", zap.Any("fromToken", fromToken.Hex()))

	abi, err := contractErc20.ERC20MetaData.GetAbi()
	if err != nil {
		return err
	}

	pack, err := abi.Pack("increaseAllowance", spender, addedValue)
	if err != nil {
		return err
	}

	value := big.NewInt(0)
	opts, err := ethutils.NewSignedTransaction(pack, fromAddress.Hex(), spender.Hex(), value, s.PrivateKey, s.Client)
	if err != nil {
		return err
	}

	tokenInstance, err := contractErc20.NewERC20(fromToken, s.Client)
	if err != nil {
		return err
	}
	trans, err := tokenInstance.IncreaseAllowance(opts, spender, addedValue)
	if err != nil {
		return err
	}

	l.Logger.Info("Allowance transaction", zap.Any("hash", trans.Hash().Hex()))

	res, err := ethutils.ConfirmEthereumTransaction(trans.Hash().Hex(), 60, time.Minute, s.Client)
	if err != nil {
		return err
	}
	if res == ethutils.STATUS_FAIL {
		return errors.New("500", "Unable to confirm transaction")
	}

	return nil
}

func (s *swapService) PackConvert(path *[]cmn.Address, amount, minReturn *big.Int, affiliate cmn.Address) ([]byte, error) {
	abi, err := contracts.IBancorNetworkMetaData.GetAbi()
	if err != nil {
		return nil, errors.New("500", "No ABI")
	}

	// ropsten test network deployed old contract, hence using old method
	//pack, err := abi.Pack("convertByPath2", path, amount, minReturn, affiliate)
	pack, err := abi.Pack("convertByPath", path, amount, minReturn, affiliate, affiliate, big.NewInt(0))
	if err != nil {
		return nil, errors.New("500", "Unable to pack")
	}

	return pack, nil
}

// EstimateRate get token exchange rate based on from amount
func (s *swapService) EstimateRate(from, to string, amount *big.Int) (*swapfactory.ExchangePair, error) {
	fromHex := cmn.HexToAddress(from)
	toHex := cmn.HexToAddress(to)
	bancorAddr := cmn.HexToAddress(config.Configuration.BancorAddress)

	bancorModule, err := contracts.NewIBancorNetwork(bancorAddr, s.Client)
	if err != nil {
		return nil, err
	}

	convertAddrs, err := bancorModule.ConversionPath(nil, fromHex, toHex)
	if err != nil {
		return nil, err
	}

	result, err := bancorModule.RateByPath(nil, convertAddrs, amount)
	if err != nil {
		return nil, err
	}

	gasPriceWei, err := s.Client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	return &swapfactory.ExchangePair{
		ContractName:   "Bancor",
		AmountIn:       amount,
		AmountOut:      result,
		TxFee:          gasPriceWei,
		ConversionPath: &convertAddrs,
	}, nil
}

// NewSwapService - creating repository
func NewSwapService(client *ethclient.Client, privateKey *ecdsa.PrivateKey) ISwapService {
	return &swapService{
		Client:     client,
		PrivateKey: privateKey,
	}
}
