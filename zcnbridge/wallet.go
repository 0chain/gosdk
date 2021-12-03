package zcnbridge

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/0chain/gosdk/core/logger"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/zcnbridge/chain"
	"github.com/0chain/gosdk/zcnbridge/log"
	"github.com/0chain/gosdk/zcnbridge/wallet"
	"github.com/0chain/gosdk/zcncore"
	"go.uber.org/zap"
)

func (b *Bridge) SetupChain() {
	initChain()
}

func (b *Bridge) RestoreChain() {
	restoreChain()
}

func (b *Bridge) SetupSDK(cfg chain.Config) {
	err := b.initSDK(cfg)
	if err != nil {
		log.Logger.Fatal("failed to setup ZCNSDK", zap.Error(err))
	}
}

// SetupWallet Sets up the wallet and node
// Wallet setup reads keys from keyfile and registers in the 0chain
func (b *Bridge) SetupWallet() {
	walletConfig, err := initZCNWallet()
	if err != nil {
		log.Logger.Fatal("failed to setup wallet", zap.Error(err))
	}
	b.Instance.wallet = walletConfig
}

func (b *Bridge) SetupEthereumWallet() {
	var ethWalletConfig, err = b.CreateEthereumWallet()
	if err != nil {
		log.Logger.Fatal("failed to setup ethereum wallet", zap.Error(err))
	} else {
		log.Logger.Info("created ethereum wallet", zap.Error(err))
	}

	b.Instance.ethereumWallet = ethWalletConfig
}

func initZCNWallet() (*wallet.Wallet, error) {
	file := filepath.Join(getConfigDir(), "wallet.json")
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

// initSDK runs zcncore.SetLogFile, zcncore.SetLogLevel and zcncore.InitZCNSDK using provided Config.
// If an error occurs during execution, the program terminates with code 2 and the error will be written in os.Stderr.
// setupZCNSDK should be used only once while application is starting.
func (b *Bridge) initSDK(cfg chain.Config) error {
	var logName = cfg.LogDir() + "/zsdk.log"
	zcncore.SetLogFile(logName, false)
	zcncore.SetLogLevel(logLevelFromStr(cfg.LogLvl()))
	serverChain := chain.GetServerChain()
	err := zcncore.InitZCNSDK(
		serverChain.BlockWorker,
		serverChain.SignatureScheme,
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
