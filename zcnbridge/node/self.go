package node

import (
	"time"

	"github.com/0chain/gosdk/zcnbridge/wallet"
)

// Node represent self node type.
type Node struct {
	wallet    *wallet.Wallet
	startTime time.Time
}

var (
	self = &Node{}
)

// Start writes to self node current time, sets wallet, external id and url
//
// Start should be used only once while application is starting.
func Start(wallet *wallet.Wallet) {
	self = &Node{
		wallet:    wallet,
		startTime: time.Now(),
	}
}

// GetWalletString returns marshaled to JSON string nodes wallet.
func GetWalletString() (string, error) {
	return self.wallet.StringJSON()
}

func SetWallet(wall *wallet.Wallet) {
	self.wallet = wall
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
func StartTime() time.Time {
	return self.startTime
}
