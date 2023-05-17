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
		return nil, err
	}
	return client, err
}

func (app *Znft) createSignedTransactionFromKeyStore(ctx context.Context) (*bind.TransactOpts, error) {
	var (
		signerAddress = common.HexToAddress(app.cfg.WalletAddress)
		password      = app.cfg.VaultPassword
		value         = app.cfg.Value
	)

	client, err := CreateEthClient(app.cfg.EthereumNodeURL)
	if err != nil {
		err := errors.Wrap(err, "failed to create ethereum client")
		Logger.Fatal(err)
		return nil, err
	}

	keyDir := path.Join(app.cfg.Homedir, WalletDir)
	ks := keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)
	signer := accounts.Account{
		Address: signerAddress,
	}
	signerAcc, err := ks.Find(signer)
	if err != nil {
		err := errors.Wrapf(err, "signer: %s", signerAddress.Hex())
		Logger.Fatal(err)
		return nil, err
	}

	chainID, err := client.ChainID(ctx)
	if err != nil {
		err := errors.Wrap(err, "failed to get chain ID")
		Logger.Fatal(err)
		return nil, err
	}

	err = ks.TimedUnlock(signer, password, time.Second*2)
	if err != nil {
		Logger.Fatal(err)
		return nil, err
	}

	opts, err := bind.NewKeyStoreTransactorWithChainID(ks, signerAcc, chainID)
	if err != nil {
		Logger.Fatal(err)
		return nil, err
	}

	valueWei := new(big.Int).Mul(big.NewInt(value), big.NewInt(params.Wei))

	opts.Value = valueWei // in wei (= no funds)

	return opts, nil
}

func (app *Znft) createSignedTransactionFromKeyStoreWithGasPrice(ctx context.Context, gasLimitUnits uint64) (*bind.TransactOpts, error) { //nolint
	client, err := CreateEthClient(app.cfg.EthereumNodeURL)
	if err != nil {
		err := errors.Wrap(err, "failed to create ethereum client")
		Logger.Fatal(err)
		return nil, err
	}

	nonce, err := client.PendingNonceAt(ctx, common.HexToAddress(app.cfg.WalletAddress))
	if err != nil {
		Logger.Fatal(err)
		return nil, err
	}

	gasPriceWei, err := client.SuggestGasPrice(ctx)
	if err != nil {
		Logger.Fatal(err)
		return nil, err
	}

	opts, err := app.createSignedTransactionFromKeyStore(ctx)
	if err != nil {
		Logger.Fatal(err)
		return nil, err
	}

	opts.Nonce = big.NewInt(int64(nonce)) // (nil = use pending state), look at bind.CallOpts{Pending: true}
	opts.GasLimit = gasLimitUnits         // in units  (0 = estimate)
	opts.GasPrice = gasPriceWei           // wei (nil = gas price oracle)

	return opts, nil
}
