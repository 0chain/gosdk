package zcnbridge

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/0chain/gosdk/core/common"

	"github.com/0chain/gosdk/core/logger"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/zcnbridge/chain"
	"github.com/0chain/gosdk/zcnbridge/log"
	"github.com/0chain/gosdk/zcnbridge/wallet"
	"github.com/0chain/gosdk/zcncore"
	"go.uber.org/zap"
)

// SetupWallets Sets up the wallet and node
// Wallet setup reads keys from keyfile and registers in the 0chain
func (b *Bridge) SetupWallets(cfg chain.Config) {
	err := b.SetupSDK(cfg)
	if err != nil {
		log.Logger.Fatal("failed to setup ZCNSDK", zap.Error(err))
	}

	walletConfig, err := SetupZCNWallet(cfg)
	if err != nil {
		log.Logger.Fatal("failed to setup wallet", zap.Error(err))
	}

	ethWalletConfig, err := b.SetupEthereumWallet()
	if err != nil {
		log.Logger.Fatal("failed to setup ethereum wallet", zap.Error(err))
	}

	b.Instance.startTime = common.Now()
	b.Instance.wallet = walletConfig
	b.Instance.ethereumWallet = ethWalletConfig
}

func SetupZCNWallet(cfg chain.Config) (*wallet.Wallet, error) {
	var (
		err  error
		home string
	)

	home, err = os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	file := filepath.Join(home, ".zcn", cfg.WalletFile())
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil, errors.Wrap(err, "error opening the wallet "+file)
	}

	f, _ := os.Open(file)
	clientBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrap(err, "error reading the wallet")
	}

	clientConfig := string(clientBytes)

	w, err := wallet.AssignWallet(clientConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to assign the wallet")
	}

	err = w.RegisterToMiners()
	if err != nil {
		return nil, errors.Wrap(err, "failed to register to miners")
	}

	return w, nil
}

// SetupSDK runs zcncore.SetLogFile, zcncore.SetLogLevel and zcncore.InitZCNSDK using provided Config.
// If an error occurs during execution, the program terminates with code 2 and the error will be written in os.Stderr.
// setupZCNSDK should be used only once while application is starting.
func (b *Bridge) SetupSDK(cfg chain.Config) error {
	var logName = cfg.LogDir() + "/zsdk.log"
	zcncore.SetLogFile(logName, false)
	zcncore.SetLogLevel(logLevelFromStr(cfg.LogLvl()))
	serverChain := chain.GetServerChain()
	err := zcncore.InitZCNSDK(
		cfg.BlockWorker(),
		cfg.SignatureScheme(),
		zcncore.WithChainID(serverChain.ID),
		zcncore.WithMinSubmit(serverChain.MinSubmit),
		zcncore.WithMinConfirmation(serverChain.MinCfm),
		zcncore.WithEthereumNode(b.EthereumNodeURL),
	)

	return err
}

// logLevelFromStr converts string log level to gosdk logger level int value.
func logLevelFromStr(level string) int {
	switch level {
	case "none":
		return logger.NONE
	case "fatal":
		return logger.FATAL
	case "error":
		return logger.ERROR
	case "info":
		return logger.INFO
	case "debug":
		return logger.DEBUG

	default:
		return -1
	}
}
