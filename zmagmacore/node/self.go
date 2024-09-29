// DEPRECATED: This package is deprecated and will be removed in a future release.
package node

import (
	"net/url"
	"strconv"
	"time"

	"github.com/0chain/gosdk/zmagmacore/wallet"
)

// Node represent self node type.
type Node struct {
	url       string
	wallet    *wallet.Wallet
	extID     string
	startTime time.Time
}

var (
	self = &Node{}
)

// Start writes to self node current time, sets wallet, external id and url
//
// Start should be used only once while application is starting.
func Start(host string, port int, extID string, wallet *wallet.Wallet) {
	self = &Node{
		url:       makeHostURL(host, port),
		wallet:    wallet,
		extID:     extID,
		startTime: time.Now(),
	}
}

func makeHostURL(host string, port int) string {
	if host == "" {
		host = "localhost"
	}

	uri := url.URL{
		Scheme: "http",
		Host:   host + ":" + strconv.Itoa(port),
	}

	return uri.String()
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

func ExtID() string {
	return self.extID
}

// PublicKey returns id of Node.
func PublicKey() string {
	return self.wallet.PublicKey()
}

// PrivateKey returns id of Node.
func PrivateKey() string {
	return self.wallet.PrivateKey()
}

// StartTime returns time when Node is started.
func StartTime() time.Time {
	return self.startTime
}
