package zcncore

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"regexp"
	"sync"

	"github.com/0chain/gosdk/core/tokenrate"
	"github.com/0chain/gosdk/core/zcncrypto"
	hdwallet "github.com/0chain/gosdk/zcncore/ethhdwallet"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"golang.org/x/crypto/sha3"
)

// TODO change to real wallets
const walletAddr = "0xb9EF770B6A5e12E45983C5D80545258aA38F3B78"
const tokenAddress = "0x28b149020d2152179873ec60bed6bf7cd705775d"

var once sync.Once

var ethClient *ethclient.Client

var getEthClient = func() (*ethclient.Client, error) {
	var err error

	once.Do(func() {
		if len(_config.chain.EthNode) == 0 {
			err = fmt.Errorf("eth node SDK not initialized")
			return
		}

		logging.Info("requesting from ", _config.chain.EthNode)
		ethClient, err = ethclient.Dial(_config.chain.EthNode)
	})

	return ethClient, err
}

// TokensToEth - converting wei to eth tokens
func TokensToEth(tokens int64) float64 {
	fbalance := new(big.Float)
	fbalance.SetString(fmt.Sprint(tokens))
	ethValue, _ := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18))).Float64()
	return ethValue
}

// TokensToEth - converting eth tokens to wei
func EthToTokens(tokens float64) int64 {
	return int64(tokens * float64(params.Ether))
}

func GTokensToEth(tokens int64) float64 {
	return float64(tokens) / float64(params.GWei)
}

func GEthToTokens(gwei float64) int64 {
	return int64(gwei * float64(params.GWei))
}

// GetWalletAddrFromEthMnemonic - wallet ETH address from mnemoninnc
func GetWalletAddrFromEthMnemonic(mnemonic string) (string, error) {
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return "", err
	}

	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, false)
	if err != nil {
		return "", err
	}

	privKey, err := wallet.PrivateKeyHex(account)
	if err != nil {
		return "", err
	}

	type ethWalletinfo struct {
		ID         string `json:"ID"`
		PrivateKey string `json:"PrivateKey"`
	}

	res, err := json.Marshal(ethWalletinfo{ID: account.Address.Hex(), PrivateKey: privKey})
	return string(res), err
}

// GetEthBalance - getting back balance for ETH wallet
func GetEthBalance(ethAddr string, cb GetBalanceCallback) error {
	go func() {
		value, err := getBalanceFromEthNode(ethAddr)
		if err != nil {
			logging.Error(err)
			cb.OnBalanceAvailable(StatusError, 0, err.Error())
			return
		}
		cb.OnBalanceAvailable(StatusSuccess, value, "")
	}()
	return nil
}

// IsValidEthAddress - multiple checks for valid ETH address
func IsValidEthAddress(ethAddr string) (bool, error) {
	client, err := getEthClient()
	if err != nil {
		return false, err
	}

	return isValidEthAddress(ethAddr, client)
}

// IsEthTransactionVerified checks if the transaction - given its hash - is verified on the ethereum network
//   - txHash: transaction hash
func IsEthTransactionVerified(txHash string) (bool, error) {
	client, err := getEthClient()
	if err != nil {
		return false, err
	}

	var (
		tx      *types.Transaction
		pending bool
	)

	tx, pending, err = client.TransactionByHash(context.Background(), common.HexToHash(txHash))
	if err != nil {
		return false, err
	}
	return tx != nil && !pending, nil
}

func isValidEthAddress(ethAddr string, client *ethclient.Client) (bool, error) {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	if !re.MatchString(ethAddr) {
		return false, fmt.Errorf("regex error")
	}

	address := common.HexToAddress(ethAddr)
	balance, err := client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}

	isContract := balance.Int64() > 0
	return isContract, nil
}

// CreateWalletFromEthMnemonic - creating new wallet from Eth mnemonics
func CreateWalletFromEthMnemonic(mnemonic, password string, statusCb WalletCallback) error {
	if len(_config.chain.Miners) < 1 || len(_config.chain.Sharders) < 1 {
		return fmt.Errorf("SDK not initialized")
	}
	go func() {
		sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
		_, err := sigScheme.GenerateKeysWithEth(mnemonic, password)
		if err != nil {
			statusCb.OnWalletCreateComplete(StatusError, "", err.Error())
			return
		}
	}()
	return nil
}

// CheckEthHashStatus - checking the status of ETH transaction
// possible values 0 or 1
func CheckEthHashStatus(hash string) int {
	txHash := common.HexToHash(hash)

	var client *ethclient.Client
	var err error
	if client, err = getEthClient(); err != nil {
		return -1
	}

	tx, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		return -1
	}
	return int(tx.Status)
}

// ConvertZcnTokenToETH - converting Zcn tokens to Eth
//   - f: ZCN tokens amount
func ConvertZcnTokenToETH(f float64) (float64, error) {
	ethRate, err := tokenrate.GetUSD(context.TODO(), "eth")
	if err != nil {
		return 0, err
	}
	return f * ethRate, nil
}

// SuggestEthGasPrice - return back suggested price for gas
func SuggestEthGasPrice() (int64, error) {
	var client *ethclient.Client
	var err error
	if client, err = getEthClient(); err != nil {
		return 0, err
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return 0, err
	}

	return gasPrice.Int64(), nil
}

// TransferEthTokens - transfer ETH tokens to multisign wallet
func TransferEthTokens(fromPrivKey string, amountTokens, gasPrice int64) (string, error) {
	var client *ethclient.Client
	var err error
	if client, err = getEthClient(); err != nil {
		return "", err
	}

	privateKey, err := crypto.HexToECDSA(fromPrivKey)
	if err != nil {
		return "", err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", err
	}

	toAddress := common.HexToAddress(walletAddr)
	tokenAddress := common.HexToAddress(tokenAddress)

	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]
	fmt.Println(hexutil.Encode(methodID)) // 0xa9059cbb

	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAddress)) // 0x0000000000000000000000004592d8f8d7b001e72cb26a73e4fa1806a51ac79d

	amount := new(big.Int)
	amount.SetInt64(amountTokens)

	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAmount)) // 0x00000000000000000000000000000000000000000000003635c9adc5dea00000

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		To:   &tokenAddress,
		Data: data,
	})
	if err != nil {
		log.Fatal(err)
	}

	txData := &types.LegacyTx{
		Nonce:    nonce,
		GasPrice: big.NewInt(gasPrice),
		Gas:      gasLimit,
		To:       &tokenAddress,
		Value:    amount,
		Data:     data,
	}
	tx := types.NewTx(txData)

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return "", err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return "", err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}

	return signedTx.Hash().Hex(), nil
}

func getBalanceFromEthNode(ethAddr string) (int64, error) {
	if client, err := getEthClient(); err == nil {
		account := common.HexToAddress(ethAddr)
		logging.Info("for eth address", account)
		balance, err := client.BalanceAt(context.Background(), account, nil)
		if err != nil {
			return 0, err
		}

		logging.Info("balance", balance.String())

		return balance.Int64(), nil
	} else {
		return 0, err
	}
}
