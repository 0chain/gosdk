package node

import (
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zcnbridge/wallet"
)

// Node represent self node type.
type Node struct {
	wallet         *wallet.Wallet
	ethereumWallet *wallet.EthereumWallet
	startTime      common.Timestamp
	nonce          int64
}

var (
	self = &Node{}
)

// Start writes to self node current time, sets wallet, external id and url
// Should be used only once while application is starting.
func Start(wallet *wallet.Wallet, ethWallet *wallet.EthereumWallet) {
	self = &Node{
		wallet:         wallet,
		startTime:      common.Now(),
		ethereumWallet: ethWallet,
	}
}

// GetEthereumWallet returns ethereum wallet string
func GetEthereumWallet() *wallet.EthereumWallet {
	return self.ethereumWallet
}

// GetWalletString returns marshaled to JSON string nodes wallet.
func GetWalletString() (string, error) {
	return self.wallet.StringJSON()
}

// ID returns id of Node.
func ID() string {
	return self.wallet.ID()
}

// PublicKey returns id of Node.
func PublicKey() string {
	return self.wallet.PublicKey()
}

// StartTime returns time when Node is started.
func StartTime() common.Timestamp {
	return self.startTime
}

func IncrementNonce() int64 {
	self.nonce++
	return self.nonce
}
