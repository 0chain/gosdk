package sdk

import "sync"

type Consensus struct {
	mu              *sync.RWMutex
	consensus       int // Total successful and valid response from blobbers
	consensusThresh int // Consensus threshold percentage
	fullconsensus   int // Total number of blobbers in allocation
}

// Done increase consensus by 1
func (c *Consensus) Done() {
	c.mu.Lock()
	c.consensus++
	c.mu.Unlock()
}

// Reset reset consensus to 0
func (c *Consensus) Reset() {
	c.mu.Lock()
	c.consensus = 0
	c.mu.Unlock()
}

func (c *Consensus) Init(threshConsensus, fullConsensus int) {
	c.mu.Lock()
	c.consensusThresh = threshConsensus
	c.fullconsensus = fullConsensus
	c.mu.Unlock()
}

func (c *Consensus) getConsensus() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.consensus
}

func (c *Consensus) isConsensusOk() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.consensus >= c.consensusThresh
}
