package zcnbridge

import (
	"fmt"
	"io/ioutil"
	"path"

	"gopkg.in/yaml.v3"
)

// CreateInitialOwnerConfig create initial config for the bridge owner using argument,
// filename - where file is saved (default: HOME/.zcn),
// ethereumAddress - client Ethereum address (should be also registered in the local key storage),
// bridgeAddress - bridge contract address,
// wzcnAddress - token contract address,
// authorizersAddress - authorizersAddress contract address,
// ethereumNodeURL - Ethereum node url (usually, Infura/Alchemy),
// gas - default gas value for Ethereum transaction,
// value - is a default value for Ethereum transaction, default is 0,
func CreateInitialOwnerConfig(
	filename, ethereumAddress, bridgeAddress, wzcnAddress, authorizersAddress, ethereumNodeURL, password string,
	gas, value int64,
) {
	type BridgeOwnerYaml struct {
		// KeyStorage unlock storage
		Password string `yaml:"Password"`
		// User's address
		EthereumAddress string `yaml:"EthereumAddress"`
		// Address of Ethereum bridge contract
		BridgeAddress string `yaml:"BridgeAddress"`
		// Address of WZCN token (Example: https://ropsten.etherscan.io/token/0x930E1BE76461587969Cb7eB9BFe61166b1E70244)
		WzcnAddress string `yaml:"WzcnAddress"`
		// Address of Authorizers smart contract
		AuthorizersAddress string `yaml:"AuthorizersAddress"`
		// URL of ethereum RPC node (infura or alchemy)
		EthereumNodeURL string `yaml:"EthereumNodeURL"`
		// Gas limit to execute ethereum transaction
		GasLimit int64 `yaml:"GasLimit"`
		// Value to execute ZCN smart contracts in wei
		Value int64 `yaml:"Value"`
	}

	type Bridge struct {
		Owner *BridgeOwnerYaml
	}

	cfg := &BridgeOwnerYaml{
		Password:           password,
		EthereumAddress:    ethereumAddress,
		BridgeAddress:      bridgeAddress,
		WzcnAddress:        wzcnAddress,
		AuthorizersAddress: authorizersAddress,
		EthereumNodeURL:    ethereumNodeURL,
		GasLimit:           gas,
		Value:              value,
	}

	bridge := Bridge{Owner: cfg}

	data, err := yaml.Marshal(bridge)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = ioutil.WriteFile(path.Join(GetConfigDir(), filename), data, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("config written to " + filename)
}

// CreateInitialClientConfig create initial config for the bridge client using argument,
// filename - where file is saved (default: HOME/.zcn),
// ethereumAddress - client Ethereum address (should be also registered in the local key storage),
// bridgeAddress - bridge contract address,
// wzcnAddress - token contract address,
// ethereumNodeURL - Ethereum node url (usually, Infura/Alchemy),
// gas - default gas value for Ethereum transaction,
// value - is a default value for Ethereum transaction, default is 0,
// consensus - is a consensus to find for burn/mint tickets
func CreateInitialClientConfig(
	filename, ethereumAddress, bridgeAddress, wzcnAddress, ethereumNodeURL, password string,
	gas, value int64,
	consensus float64,
) {
	type BridgeClientYaml struct {
		// KeyStorage unlock storage
		Password string `yaml:"Password"`
		// User's address
		EthereumAddress string `yaml:"EthereumAddress"`
		// Address of Ethereum bridge contract
		BridgeAddress string `yaml:"BridgeAddress"`
		// Address of WZCN token (Example: https://ropsten.etherscan.io/token/0x930E1BE76461587969Cb7eB9BFe61166b1E70244)
		WzcnAddress string `yaml:"WzcnAddress"`
		// URL of ethereum RPC node (infura or alchemy)
		EthereumNodeURL string `yaml:"EthereumNodeURL"`
		// Gas limit to execute ethereum transaction
		GasLimit int64 `yaml:"GasLimit"`
		// Value to execute ZCN smart contracts in wei
		Value              int64   `yaml:"Value"`
		ConsensusThreshold float64 `yaml:"ConsensusThreshold"`
	}

	type Bridge struct {
		Bridge *BridgeClientYaml
	}

	cfg := &BridgeClientYaml{
		Password:           password,
		EthereumAddress:    ethereumAddress,
		BridgeAddress:      bridgeAddress,
		WzcnAddress:        wzcnAddress,
		EthereumNodeURL:    ethereumNodeURL,
		GasLimit:           gas,
		Value:              value,
		ConsensusThreshold: consensus,
	}

	bridge := Bridge{Bridge: cfg}

	data, err := yaml.Marshal(bridge)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = ioutil.WriteFile(path.Join(GetConfigDir(), filename), data, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("config written to " + filename)
}
