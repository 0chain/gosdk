package znft

import (
	"context"
	"math/big"
	"path"
	"time"

	. "github.com/0chain/gosdk/zcncore"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/pkg/errors"
)

func (conf *Configuration) CreateEthClient() (*ethclient.Client, error) {
	client, err := ethclient.Dial(conf.EthereumNodeURL)
	if err != nil {
		Logger.Error(err)
	}
	return client, err
}

func (conf *Configuration) CreateSignedTransactionFromKeyStore(ctx context.Context, gasLimitUnits uint64) *bind.TransactOpts {
	var (
		signerAddress = common.HexToAddress(conf.WalletAddress)
		password      = conf.VaultPassword
		value         = conf.Value
	)

	client, err := conf.CreateEthClient()
	if err != nil {
		Logger.Fatal(errors.Wrap(err, "failed to create ethereum client"))
	}

	keyDir := path.Join(conf.Homedir, WalletDir)
	ks := keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)
	signer := accounts.Account{
		Address: signerAddress,
	}
	signerAcc, err := ks.Find(signer)
	if err != nil {
		Logger.Fatal(errors.Wrapf(err, "signer: %s", signerAddress.Hex()))
	}

	chainID, err := client.ChainID(ctx)
	if err != nil {
		Logger.Fatal(errors.Wrap(err, "failed to get chain ID"))
	}

	nonce, err := client.PendingNonceAt(ctx, signerAddress)
	if err != nil {
		Logger.Fatal(err)
	}

	gasPriceWei, err := client.SuggestGasPrice(ctx)
	if err != nil {
		Logger.Fatal(err)
	}

	err = ks.TimedUnlock(signer, password, time.Second*2)
	if err != nil {
		Logger.Fatal(err)
	}

	opts, err := bind.NewKeyStoreTransactorWithChainID(ks, signerAcc, chainID)
	if err != nil {
		Logger.Fatal(err)
	}

	valueWei := new(big.Int).Mul(big.NewInt(value), big.NewInt(params.Wei))

	opts.Nonce = big.NewInt(int64(nonce))
	opts.Value = valueWei         // in wei
	opts.GasLimit = gasLimitUnits // in units
	opts.GasPrice = gasPriceWei   // wei

	return opts
}
