// DEPRECATED: This package is deprecated and will be removed in a future release.
package chain

import (
	"github.com/0chain/gosdk/zmagmacore/time"
)

// Chain represents data structure that holds the chain data.
type Chain struct {
	ID           string
	Version      string
	CreationDate time.Timestamp
	OwnerID      string
	BlockWorker  string
}

// serverChain is the chain object of the chain the server is responsible for.
var serverChain = new(Chain)

// SetServerChain sets the server chain object to package variable serverChain.
func SetServerChain(c *Chain) {
	serverChain = c
}

// GetServerChain returns the chain object for the server chain.
func GetServerChain() *Chain {
	return serverChain
}

// NewChain creates a new Chain.
func NewChain(id, OwnerID, blockWorker string) *Chain {
	chain := Provider()
	chain.ID = id
	chain.OwnerID = OwnerID
	chain.BlockWorker = blockWorker
	return chain
}

// Provider returns entity for chain object.
func Provider() *Chain {
	c := &Chain{}
	c.Version = "1.0"
	c.CreationDate = time.Now()
	return c
}
