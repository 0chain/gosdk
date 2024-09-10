package client

import "sync"

func (h *NodeHolder) Healthy() (res []string) {
	h.guard.Lock()
	defer h.guard.Unlock()

	return h.nodes[:h.consensus]
}

type NodeHolder struct {
	consensus int
	guard     sync.Mutex
	stats     map[string]*NodeStruct
	nodes     []string
}

func NewHolder(nodes []string, consensus int) *NodeHolder {
	if len(nodes) < consensus {
		panic("consensus is not correct")
	}
	holder := NodeHolder{consensus: consensus, stats: make(map[string]*NodeStruct)}

	for _, n := range nodes {
		holder.nodes = append(holder.nodes, n)
		holder.stats[n] = NewNode(n)
	}
	return &holder
}
func NewNode(id string) *NodeStruct {
	return &NodeStruct{
		id:     id,
		weight: 1,
		stats:  []int{1},
	}
}

type NodeStruct struct {
	id     string
	weight int64
	stats  []int
}
