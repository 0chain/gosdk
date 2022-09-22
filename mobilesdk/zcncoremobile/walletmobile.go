package zcncoremobile

import (
	"encoding/json"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/version"
	models "github.com/0chain/gosdk/mobilesdk/zbox"
	"github.com/0chain/gosdk/mobilesdk/zcncore"
	"github.com/0chain/gosdk/mobilesdk/zcncrypto"
)

// CreateWallet creates the wallet for to configure signature scheme.
// It also registers the wallet again to blockchain.
func CreateWallet(statusCb models.WalletCallback) error {
	if len(zcncore.GetConfig().GetChain().Miners) < 1 || len(zcncore.GetConfig().GetChain().Sharders) < 1 {
		return errors.New("", "SDK not initialized")
	}
	go func() {
		sigScheme := zcncrypto.NewSignatureScheme(zcncore.GetConfig().GetChain().SignatureScheme)
		wallet, err := sigScheme.GenerateKeys()
		if err != nil {
			statusCb.OnWalletCreateComplete(zcncore.StatusError, "", err.Error())
			return
		}
		err = zcncore.RegisterToMiners(wallet, statusCb)
		if err != nil {
			statusCb.OnWalletCreateComplete(zcncore.StatusError, "", err.Error())
			return
		}
	}()
	return nil
}

// IsMnemonicValid is an utility function to check the mnemonic valid
func IsMnemonicValid(mnemonic string) bool {
	return zcncrypto.IsMnemonicValid(mnemonic)
}

// RecoverWallet recovers the previously generated wallet using the mnemonic.
// It also registers the wallet again to block chain.
func RecoverWallet(mnemonic string, statusCb models.WalletCallback) error {
	if !zcncrypto.IsMnemonicValid(mnemonic) {
		return errors.New("", "Invalid mnemonic")
	}
	go func() {
		sigScheme := zcncrypto.NewSignatureScheme(zcncore.GetConfig().GetChain().SignatureScheme)
		wallet, err := sigScheme.RecoverKeys(mnemonic)
		if err != nil {
			statusCb.OnWalletCreateComplete(zcncore.StatusError, "", err.Error())
			return
		}
		err = zcncore.RegisterToMiners(wallet, statusCb)
		if err != nil {
			statusCb.OnWalletCreateComplete(zcncore.StatusError, "", err.Error())
			return
		}

	}()
	return nil
}

// ConvertToValue converts ZCN tokens to value
func ConvertToValue(token float64) int64 {
	return int64(token * float64(zcncore.TOKEN_UNIT))
}

// ConvertToToken converts the value to ZCN tokens
func ConvertToToken(value int64) float64 {
	return float64(value) / float64(zcncore.TOKEN_UNIT)
}

// SetWalletInfo should be set before any transaction or client specific APIs
// splitKeyWallet parameter is valid only if SignatureScheme is "BLS0Chain"
func SetWalletInfo(w string, splitKeyWallet bool) error {
	var wallet zcncrypto.Wallet
	localCfg := zcncore.GetConfig()
	err := json.Unmarshal([]byte(w), &wallet)
	if err == nil {
		localCfg.SetWallet(wallet)
		if localCfg.GetChain().SignatureScheme == "bls0chain" {
			localCfg.SetIsSplitWallet(splitKeyWallet)
		}
		localCfg.SetIsValidWallet(true)
	}

	return err
}

// GetBalance retreives wallet balance from sharders
func GetBalance(cb models.GetBalanceCallback) error {
	err := zcncore.CheckConfig()
	if err != nil {
		return err
	}
	go func() {
		value, info, err := zcncore.GetBalanceFromSharders(zcncore.GetConfig().GetWallet().ClientID)
		if err != nil {
			zcncore.Logger.Error(err)
			cb.OnBalanceAvailable(zcncore.StatusError, 0, info)
			return
		}
		cb.OnBalanceAvailable(zcncore.StatusSuccess, value, info)
	}()
	return nil
}

// GetVersion - returns version string
func GetVersion() string {
	return version.VERSIONSTR
}

// SetLogLevel set the log level.
// lvl - 0 disabled; higher number (upto 4) more verbosity
func SetLogLevel(lvl int) {
	zcncore.Logger.SetLevel(lvl)
}

func Init(chainConfigJSON string) error {
	return zcncore.Init(chainConfigJSON)
}
