package znft

import (
	"context"
	"math/big"
	"path"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/pkg/errors"
)

func CreateEthClient(ethereumNodeURL string) (*ethclient.Client, error) {
	client, err := ethclient.Dial(ethereumNodeURL)
	if err != nil {
		Logger.Error(err)
	}
	return client, err
}

func (app *App) createSignedTransactionFromKeyStore() *bind.TransactOpts {
	var (
		signerAddress = common.HexToAddress(app.cfg.WalletAddress)
		password      = app.cfg.VaultPassword
		value         = app.cfg.Value
	)

	client, err := CreateEthClient(app.cfg.EthereumNodeURL)
	if err != nil {
		Logger.Fatal(errors.Wrap(err, "failed to create ethereum client"))
	}

	keyDir := path.Join(app.cfg.Homedir, WalletDir)
	ks := keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)
	signer := accounts.Account{
		Address: signerAddress,
	}
	signerAcc, err := ks.Find(signer)
	if err != nil {
		Logger.Fatal(errors.Wrapf(err, "signer: %s", signerAddress.Hex()))
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		Logger.Fatal(errors.Wrap(err, "failed to get chain ID"))
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

	opts.Value = valueWei // in wei (= no funds)

	return opts
}

func (app *App) createSignedTransactionFromKeyStoreWithGasPrice(gasLimitUnits uint64) *bind.TransactOpts {
	client, err := CreateEthClient(app.cfg.EthereumNodeURL)
	if err != nil {
		Logger.Fatal(errors.Wrap(err, "failed to create ethereum client"))
	}

	nonce, err := client.PendingNonceAt(context.Background(), common.HexToAddress(app.cfg.WalletAddress))
	if err != nil {
		Logger.Fatal(err)
	}

	gasPriceWei, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		Logger.Fatal(err)
	}

	opts := app.createSignedTransactionFromKeyStore()

	opts.Nonce = big.NewInt(int64(nonce)) // (nil = use pending state), look at bind.CallOpts{Pending: true}
	opts.GasLimit = gasLimitUnits         // in units  (0 = estimate)
	opts.GasPrice = gasPriceWei           // wei (nil = gas price oracle)

	return opts
}
