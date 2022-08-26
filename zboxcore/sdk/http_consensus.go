package sdk

import (
	"fmt"
	"net/http"
)

type HttpConsensus struct {
	consensus *Consensus
	winResult string
	results   map[string]int
}

func NewHttpConsensus(threshConsensus, fullConsensus, consensusOK float32) *HttpConsensus {
	c := &HttpConsensus{
		consensus: &Consensus{},
		results:   make(map[string]int),
	}

	fmt.Println(threshConsensus, fullConsensus, consensusOK)
	c.consensus.Init(threshConsensus, fullConsensus, consensusOK)

	return c
}

func (c *HttpConsensus) GetConsensusResult() (string, bool) {

	fmt.Println(c.consensus.consensus, c.consensus.fullconsensus, c.consensus.consensusThresh, c.consensus.getConsensusRate(), c.consensus.getConsensusRequiredForOk())
	if c.consensus.isConsensusOk() {
		return c.winResult, true
	}

	return "", false
}

func (c *HttpConsensus) Add(req *http.Request, resp *http.Response, respBody []byte) {
	c.consensus.Lock()
	defer c.consensus.Unlock()

	s := string(respBody)

	n, ok := c.results[s]
	if !ok {
		c.results[s] = 1
		n = 1
	} else {
		n++
	}

	if n > c.results[c.winResult] {
		c.winResult = s
		c.consensus.consensus = float32(n)

		fmt.Println("consensus: ", n)
	} else {
		fmt.Println("consensus: ", c.results[c.winResult])
	}
}
