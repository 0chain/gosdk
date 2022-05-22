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

func (conf *Configuration) CreateEthClient() (*ethclient.Client, error) {
	client, err := ethclient.Dial(conf.EthereumNodeURL)
	if err != nil {
		Logger.Error(err)
	}
	return client, err
}

func (conf *Configuration) createSignedTransactionFromKeyStore() *bind.TransactOpts {
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

func (conf *Configuration) createSignedTransactionFromKeyStoreWithGasPrice(gasLimitUnits uint64) *bind.TransactOpts {
	client, err := conf.CreateEthClient()
	if err != nil {
		Logger.Fatal(errors.Wrap(err, "failed to create ethereum client"))
	}

	nonce, err := client.PendingNonceAt(context.Background(), common.HexToAddress(conf.WalletAddress))
	if err != nil {
		Logger.Fatal(err)
	}

	gasPriceWei, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		Logger.Fatal(err)
	}

	opts := conf.createSignedTransactionFromKeyStore()

	opts.Nonce = big.NewInt(int64(nonce)) // (nil = use pending state), look at bind.CallOpts{Pending: true}
	opts.GasLimit = gasLimitUnits         // in units  (0 = estimate)
	opts.GasPrice = gasPriceWei           // wei (nil = gas price oracle)

	return opts
}
