// Provides functions and data structures to interact with the system nodes in the context of the blockchain network.
package client

import (
	"sort"
	"sync"
	"time"
)

const statSize = 20
const defaultTimeout = 5 * time.Second

type NodeHolder struct {
	consensus int
	guard     sync.Mutex
	stats     map[string]*NodeStruct
	nodes     []string
}

type NodeStruct struct {
	id     string
	weight int64
	stats  []int
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

func (h *NodeHolder) Success(id string) {
	h.guard.Lock()
	defer h.guard.Unlock()
	h.adjustNode(id, 1)
}

func (h *NodeHolder) Fail(id string) {
	h.guard.Lock()
	defer h.guard.Unlock()
	h.adjustNode(id, -1)
}

func (h *NodeHolder) adjustNode(id string, res int) {
	n := NewNode(id)
	nodes := h.nodes
	if node, ok := h.stats[id]; ok {
		for i, v := range nodes {
			if v == id {
				nodes = append(nodes[:i], nodes[i+1:]...)
				break
			}
		}

		sourceStats := node.stats
		sourceStats = append(sourceStats, res)
		if len(sourceStats) > statSize {
			sourceStats = sourceStats[1:]
		}
		node.stats = sourceStats

		w := int64(0)
		for i, s := range sourceStats {
			w += int64(i+1) * int64(s)
		}
		node.weight = w

		n = node
	}

	i := sort.Search(len(nodes), func(i int) bool {
		return h.stats[nodes[i]].weight < n.weight
	})
	h.nodes = append(nodes[:i], append([]string{n.id}, nodes[i:]...)...)
}

func (h *NodeHolder) Healthy() (res []string) {
	h.guard.Lock()
	defer h.guard.Unlock()

	return h.nodes[:h.consensus]
}

func (h *NodeHolder) All() (res []string) {
	h.guard.Lock()
	defer h.guard.Unlock()

	return h.nodes
}
