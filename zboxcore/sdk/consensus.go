package sdk

import "sync"

type Consensus struct {
	sync.RWMutex
	consensus       float32
	consensusThresh float32
	fullconsensus   float32
}

// Done increase 1
func (req *Consensus) Done() {
	req.Lock()
	defer req.Unlock()
	req.consensus++

}

// Reset reset consensus with 0
func (req *Consensus) Reset() {
	req.Lock()
	defer req.Unlock()
	req.consensus = 0
}

func (req *Consensus) getConsensus() float32 {
	req.RLock()
	defer req.RUnlock()
	return req.consensus
}

func (req *Consensus) getConsensusRate() float32 {
	req.RLock()
	defer req.RUnlock()
	// if req.isRepair {
	// 	return (req.consensus * 100) / float32(bits.OnesCount32(req.uploadMask))
	// } else {
	return (req.consensus * 100) / req.fullconsensus
	//}
}

func (req *Consensus) getConsensusRequiredForOk() float32 {
	req.RLock()
	defer req.RUnlock()

	return (req.consensusThresh + additionalSuccessRate)
}

func (req *Consensus) isConsensusOk() bool {
	req.RLock()
	defer req.RUnlock()

	return (req.getConsensusRate() >= req.getConsensusRequiredForOk())
}

func (req *Consensus) isConsensusMin() bool {
	req.RLock()
	defer req.RUnlock()

	return (req.getConsensusRate() >= req.consensusThresh)
}
