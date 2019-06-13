package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zcncore"
)

var cfgFile string
var walletFile string

var sharders []string
var miners []string
var clientConfig string

var rootCmd = &cobra.Command{
	Use:   "zbox",
	Short: "0Box is a decentralized storage application written on the 0Chain platform",
	Long: `0Box is a decentralized storage application written on the 0Chain platform.
			Complete documentation is available at https://0chain.net`,
}

var clientWallet *zcncrypto.Wallet

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.zcn/nodes.yaml)")
	rootCmd.PersistentFlags().StringVar(&walletFile, "wallet", "", "wallet file (default is $HOME/.zcn/wallet.txt)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	nodeConfig := viper.New()
	var configDir string
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	configDir = home + "/.zcn"
	// Search config in home directory with name ".cobra" (without extension).
	nodeConfig.AddConfigPath(configDir)
	if &cfgFile != nil && len(cfgFile) > 0 {
		nodeConfig.SetConfigName(cfgFile)
	} else {
		nodeConfig.SetConfigName("nodes")
	}

	if err := nodeConfig.ReadInConfig(); err != nil {
		fmt.Println("Can't read config:", err)
		os.Exit(1)
	}
	sharders = nodeConfig.GetStringSlice("sharders")
	miners = nodeConfig.GetStringSlice("miners")
	signScheme := nodeConfig.GetString("signature_scheme")
	numKeys := nodeConfig.GetInt("num_of_keys")
	chainID := nodeConfig.GetString("chain_id")

	//TODO: move the private key storage to the keychain or secure storage
	var walletFilePath string
	if &walletFile != nil && len(walletFile) > 0 {
		walletFilePath = configDir + "/" + walletFile
	} else {
		walletFilePath = configDir + "/wallet.txt"
	}
	zcncore.InitZCNSDK(miners, sharders, signScheme)
	//set the log file
	zcncore.SetLogFile("cmdlog.log", true)
	sdk.SetLogFile("cmdlog.log", true)
	if _, err := os.Stat(walletFilePath); os.IsNotExist(err) {
		wg := &sync.WaitGroup{}
		statusBar := &ZCNStatus{wg: wg}
		wg.Add(1)
		err = zcncore.CreateWallet(numKeys, statusBar)
		if err == nil {
			wg.Wait()
		} else {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		if len(statusBar.walletString) == 0 || !statusBar.success {
			fmt.Println("Error creating the wallet." + statusBar.errMsg)
			os.Exit(1)
		}
		fmt.Println("ZCN wallet created")
		clientConfig = string(statusBar.walletString)
		file, err := os.Create(walletFilePath)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		defer file.Close()
		fmt.Fprintf(file, clientConfig)
	} else {
		f, err := os.Open(walletFilePath)
		if err != nil {
			fmt.Println("Error opening the wallet", err)
			os.Exit(1)
		}
		clientBytes, err := ioutil.ReadAll(f)
		if err != nil {
			fmt.Println("Error reading the wallet", err)
			os.Exit(1)
		}
		clientConfig = string(clientBytes)
	}
	//minerjson, _ := json.Marshal(miners)
	//sharderjson, _ := json.Marshal(sharders)
	wallet := &zcncrypto.Wallet{}
	err = json.Unmarshal([]byte(clientConfig), wallet)
	clientWallet = wallet
	if err != nil {
		fmt.Println("Invalid wallet at path:" + walletFilePath)
		os.Exit(1)
	}
	//init the storage sdk with the known miners, sharders and client wallet info
	err = sdk.InitStorageSDK(clientConfig, miners, sharders, chainID, signScheme)
	if err != nil {
		fmt.Println("Error in sdk init", err)
		os.Exit(1)
	}
}
