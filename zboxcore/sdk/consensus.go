package sdk

type Consensus struct {
	consensus       float32
	consensusThresh float32
	fullconsensus   float32
}

func (req *Consensus) getConsensusRate() float32 {
	// if req.isRepair {
	// 	return (req.consensus * 100) / float32(bits.OnesCount32(req.uploadMask))
	// } else {
	return (req.consensus * 100) / req.fullconsensus
	//}
}

func (req *Consensus) getConsensusRequiredForOk() float32 {
	return (req.consensusThresh + additionalSuccessRate)
}

func (req *Consensus) isConsensusOk() bool {
	return (req.getConsensusRate() >= req.getConsensusRequiredForOk())
}

func (req *Consensus) isConsensusMin() bool {
	return (req.getConsensusRate() >= req.consensusThresh)
}
