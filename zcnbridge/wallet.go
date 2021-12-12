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

func (b *BridgeClient) RestoreChain() {
	restoreChain()
}

func (b *BridgeClient) SetupZCNSDK(logDir, logLevel string) {
	err := b.initSDK(logDir, logLevel)
	if err != nil {
		log.Logger.Fatal("failed to setup ZCNSDK", zap.Error(err))
	}
}

// SetupZCNWallet Sets up the zcnWallet and node
// Wallet setup reads keys from keyfile and registers in the 0chain
func (b *BridgeClient) SetupZCNWallet(filename string) {
	walletConfig, err := initZCNWallet(filename)
	if err != nil {
		log.Logger.Fatal("failed to setup zcn wallet", zap.Error(err))
	}
	b.Instance.zcnWallet = walletConfig
	chain.GetServerChain().ID = walletConfig.ID()
}

func (b *BridgeClient) SetupEthereumWallet() {
	clientEthereumWallet, err := b.CreateEthereumWallet()
	if err != nil {
		log.Logger.Fatal("failed to setup client ethereum zcnWallet", zap.Error(err))
	} else {
		log.Logger.Info("created client ethereum zcnWallet", zap.Error(err))
	}

	b.Instance.ethWallet = clientEthereumWallet
}

func (b *BridgeOwner) SetupEthereumWallet() {
	ownerEthereumWallet, err := b.CreateEthereumWallet()
	if err != nil {
		log.Logger.Fatal("failed to setup owner ethereum zcnWallet", zap.Error(err))
	} else {
		log.Logger.Info("created owner ethereum zcnWallet", zap.Error(err))
	}

	b.Instance.ethWallet = ownerEthereumWallet
}

func initZCNWallet(filename string) (*wallet.Wallet, error) {
	file := filepath.Join(GetConfigDir(), filename)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil, errors.Wrap(err, "error opening the zcnWallet "+file)
	}

	f, _ := os.Open(file)
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	clientBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrap(err, "error reading the zcnWallet")
	}

	clientConfig := string(clientBytes)

	w, err := wallet.AssignWallet(clientConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to assign the zcnWallet")
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
func (b *BridgeClient) initSDK(logDir, logLevel string) error {
	var logName = logDir + "/zsdk.log"
	zcncore.SetLogFile(logName, false)
	zcncore.SetLogLevel(logLevelFromStr(logLevel))
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
