package bancor

//type ISwapService interface {
//	SwapWithConversionPath(pair *swapfactory.ExchangePair, from string) (*types.Transaction, error)
//	Approve(spender, fromToken, fromAddress cmn.Address, amountIn *big.Int) error
//	IncreaseAllowance(spender, fromToken, fromAddress cmn.Address, addedValue *big.Int) error
//	PackConvert(path *[]cmn.Address, amount, minReturn *big.Int, affiliate cmn.Address) ([]byte, error)
//	EstimateRate(from, to string, amount *big.Int) (*swapfactory.ExchangePair, error)
//}
//
//type swapService struct {
//	Client     *ethclient.Client
//	PrivateKey *ecdsa.PrivateKey
//}

//func (s *swapService) Approve(spender, fromToken, fromAddress cmn.Address, amountIn *big.Int) error {
//	l.Logger.Info("Approve called", zap.Any("fromToken", fromToken.Hex()))
//
//	ercModule, err := contracts.NewERC20(fromToken, s.Client)
//	if err != nil {
//		return err
//	}
//
//	approvePayload, err := erc20.PackApprove(spender, amountIn)
//	if err != nil {
//		return err
//	}
//
//	value := big.NewInt(0)
//	opts, err := ethutils.NewSignedTransaction(approvePayload, fromAddress.Hex(), spender.Hex(), value, s.PrivateKey, s.Client)
//	if err != nil {
//		return err
//	}
//
//	trans, err := ercModule.Approve(opts, spender, amountIn)
//	if err != nil {
//		return err
//	}
//
//	l.Logger.Info("Approve tx hash", zap.Any("hash", trans.Hash().Hex()))
//
//	res, err := ethutils.ConfirmEthereumTransaction(trans.Hash().Hex(), 60, time.Minute, s.Client)
//	if err != nil {
//		return err
//	}
//	if res == ethutils.STATUS_FAIL {
//		return errors.New("500", "Unable to confirm transaction")
//	}
//
//	return nil
//}
//
//func (s *swapService) IncreaseAllowance(spender, fromToken, fromAddress cmn.Address, addedValue *big.Int) error {
//	l.Logger.Info("IncreaseAllowance called", zap.Any("fromToken", fromToken.Hex()))
//
//	abi, err := contractErc20.ERC20MetaData.GetAbi()
//	if err != nil {
//		return err
//	}
//
//	pack, err := abi.Pack("increaseAllowance", spender, addedValue)
//	if err != nil {
//		return err
//	}
//
//	value := big.NewInt(0)
//	opts, err := ethutils.NewSignedTransaction(pack, fromAddress.Hex(), spender.Hex(), value, s.PrivateKey, s.Client)
//	if err != nil {
//		return err
//	}
//
//	tokenInstance, err := contractErc20.NewERC20(fromToken, s.Client)
//	if err != nil {
//		return err
//	}
//	trans, err := tokenInstance.IncreaseAllowance(opts, spender, addedValue)
//	if err != nil {
//		return err
//	}
//
//	l.Logger.Info("Allowance transaction", zap.Any("hash", trans.Hash().Hex()))
//
//	res, err := ethutils.ConfirmEthereumTransaction(trans.Hash().Hex(), 60, time.Minute, s.Client)
//	if err != nil {
//		return err
//	}
//	if res == ethutils.STATUS_FAIL {
//		return errors.New("500", "Unable to confirm transaction")
//	}
//
//	return nil
//}

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

// EstimateRate get token exchange rate based on from amount
//func (s *swapService) EstimateRate(from, to string, amount *big.Int) (*swapfactory.ExchangePair, error) {
//	fromHex := cmn.HexToAddress(from)
//	toHex := cmn.HexToAddress(to)
//	bancorAddr := cmn.HexToAddress(config.Configuration.BancorAddress)
//
//
//
//	bancorModule, err := contracts.NewIBancorNetwork(bancorAddr, s.Client)
//	if err != nil {
//		return nil, err
//	}
//
//	convertAddrs, err := bancorModule.ConversionPath(nil, fromHex, toHex)
//	if err != nil {
//		return nil, err
//	}
//
//	result, err := bancorModule.RateByPath(nil, convertAddrs, amount)
//	if err != nil {
//		return nil, err
//	}
//
//	gasPriceWei, err := s.Client.SuggestGasPrice(context.Background())
//	if err != nil {
//		return nil, err
//	}
//
//	return &swapfactory.ExchangePair{
//		ContractName:   "Bancor",
//		AmountIn:       amount,
//		AmountOut:      result,
//		TxFee:          gasPriceWei,
//		ConversionPath: &convertAddrs,
//	}, nil
//}

//// NewSwapService - creating repository
//func NewSwapService(client *ethclient.Client, privateKey *ecdsa.PrivateKey) ISwapService {
//	return &swapService{
//		Client:     client,
//		PrivateKey: privateKey,
//	}
//}
