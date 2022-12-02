package sdk

import "sync"

type Consensus struct {
	sync.RWMutex
	consensus       int // Total successful and valid response from blobbers
	consensusThresh int // Consensus threshold percentage
	fullconsensus   int // Total number of blobbers in allocation
}

// Done increase consensus by 1
func (c *Consensus) Done() {
	c.Lock()
	c.consensus++
	c.Unlock()
}

// Reset reset consensus to 0
func (c *Consensus) Reset() {
	c.Lock()
	c.consensus = 0
	c.Unlock()
}

func (c *Consensus) Init(threshConsensus, fullConsensus int) {
	c.Lock()
	c.consensusThresh = threshConsensus
	c.fullconsensus = fullConsensus
	c.Unlock()
}

func (c *Consensus) getConsensus() int {
	c.RLock()
	defer c.RUnlock()
	return c.consensus
}

func (c *Consensus) isConsensusOk() bool {
	c.RLock()
	defer c.RUnlock()

	return c.getConsensus() >= c.consensusThresh
}
