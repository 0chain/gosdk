package sdk

import "sync"

type Consensus struct {
	sync.RWMutex
	consensus       float32 // Total successful and valid response from blobbers
	consensusThresh float32 // Consensus threshold percentage
	fullconsensus   float32 // Total number of blobbers in allocation
	// Total successful and valid responses required from blobbers. Usually its a.DataShards + 1 but if number of parity-shards is 0 then
	// it is a.DataShards. Current implements adds 10 percent as additional percentage to consensusThreshold
	consensusRequiredForOk float32
}

// Done increase consensus by 1
func (req *Consensus) Done() {
	req.Lock()
	defer req.Unlock()
	req.consensus++

}

// Reset reset consensus to 0
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
	return (req.consensus * 100) / req.fullconsensus
}

func (req *Consensus) getConsensusRequiredForOk() float32 {
	req.RLock()
	defer req.RUnlock()

	//TODO This if block can be removed if consensus issue is fixed/considered in chunked upload
	if req.consensusRequiredForOk == 0 {
		return req.consensusThresh + 10
	}
	return (req.consensusRequiredForOk)
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
