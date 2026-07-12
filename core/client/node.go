// Provides functions and data structures to interact with the system nodes in the context of the blockchain network.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/0chain/gosdk/core/util"
)

const statSize = 20

type NodeHolder struct {
	consensus int
	guard     sync.Mutex
	stats     map[string]*NodeStruct
	nodes     []string

	lfbMu       sync.RWMutex
	lfbSharders []string
	lfbExpiry   time.Time
	lfbUpdating atomic.Bool
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

	// Keep h.nodes and h.stats in lockstep. The "new id" path above falls
	// through here without ever registering n in h.stats; that leaves an id in
	// h.nodes with no stats entry, so the comparator below dereferences a nil
	// *NodeStruct and panics — fatal for any process using this SDK (blobber,
	// validator, wasm: the whole Go runtime exits).
	// This happens whenever Fail()/Success() is called with a sharder URL that
	// wasn't in the original holder (e.g. the live URL differs from the
	// configured one after a scheme/host rewrite). Register it, and nil-guard
	// the comparator so any pre-existing divergence can't crash the runtime.
	h.stats[n.id] = n

	i := sort.Search(len(nodes), func(i int) bool {
		s := h.stats[nodes[i]]
		return s != nil && s.weight < n.weight
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

const (
	lfbMaxDrift     = 3                // sharders must be within 3 blocks of the highest LFB
	lfbCacheTTL     = 10 * time.Second // how long before the LFB cache is considered stale
	lfbQueryTimeout = 3 * time.Second
	// Overall cap on collecting LFB responses. refreshLFBCache is synchronous
	// (HealthyByLFB blocks on it), so waiting for a dead/unreachable sharder —
	// e.g. a down mainnet sharder whose TCP connect hangs — stalls every SC read
	// and recurs each cache TTL. Healthy sharders answer in well under a second;
	// stop waiting for stragglers past this deadline and use whoever responded.
	lfbCollectTimeout = 2 * time.Second
)

// HealthyByLFB returns the LFB-filtered sharder list.
// Always blocks on refresh when the cache is stale to prevent returning a list
// that includes sharders which have fallen behind since the last check.
func (h *NodeHolder) HealthyByLFB() []string {
	h.lfbMu.RLock()
	cached := h.lfbSharders
	stale := time.Now().After(h.lfbExpiry)
	h.lfbMu.RUnlock()

	if stale && h.lfbUpdating.CompareAndSwap(false, true) {
		// Always refresh synchronously — background refresh causes a window where
		// stale cached sharders (that have since fallen behind) are still returned.
		h.refreshLFBCache()
	}

	h.lfbMu.RLock()
	cached = h.lfbSharders
	h.lfbMu.RUnlock()

	if len(cached) > 0 {
		return cached
	}
	// No LFB data yet — return only the single highest-weighted sharder to minimize
	// risk of hitting a stale one. Returning all sharders would defeat the purpose.
	h.guard.Lock()
	defer h.guard.Unlock()
	if len(h.nodes) > 0 {
		return h.nodes[:1]
	}
	return nil
}

// refreshLFBCache queries all sharders concurrently for their current round, updates
// NodeHolder weights, and stores the LFB-filtered list in the cache.
func (h *NodeHolder) refreshLFBCache() {
	defer h.lfbUpdating.Store(false)

	allNodes := h.All()
	if len(allNodes) == 0 {
		return
	}

	type lfbResult struct {
		sharder string
		round   int64
		err     error
	}

	results := make(chan lfbResult, len(allNodes))

	for _, sharder := range allNodes {
		go func(s string) {
			ctx, cancel := context.WithTimeout(context.Background(), lfbQueryTimeout)
			defer cancel()

			url := fmt.Sprintf("%s/v1/current-round", s)
			req, err := util.NewHTTPGetRequestContext(ctx, url)
			if err != nil {
				results <- lfbResult{sharder: s, err: err}
				return
			}

			resp, err := req.Get()
			if err != nil {
				results <- lfbResult{sharder: s, err: err}
				return
			}

			if resp.StatusCode != http.StatusOK {
				results <- lfbResult{sharder: s, err: fmt.Errorf("status %d", resp.StatusCode)}
				return
			}

			var round int64
			if err := json.Unmarshal([]byte(resp.Body), &round); err != nil {
				results <- lfbResult{sharder: s, err: err}
				return
			}

			results <- lfbResult{sharder: s, round: round}
		}(sharder)
	}

	type sharderLFB struct {
		sharder string
		round   int64
	}
	var responding []sharderLFB
	maxLFB := int64(0)

	deadline := time.NewTimer(lfbCollectTimeout)
	defer deadline.Stop()
collect:
	for i := 0; i < len(allNodes); i++ {
		select {
		case r := <-results:
			if r.err != nil {
				logging.Debug(fmt.Sprintf("Sharder %s LFB check failed: %s", r.sharder, r.err.Error()))
				h.Fail(r.sharder)
				continue
			}
			responding = append(responding, sharderLFB{sharder: r.sharder, round: r.round})
			h.Success(r.sharder)
			if r.round > maxLFB {
				maxLFB = r.round
			}
		case <-deadline.C:
			// Dead/slow straggler(s) — stop waiting; the goroutines finish into
			// the buffered channel and are harmlessly discarded.
			logging.Debug("LFB refresh deadline hit — proceeding with responders so far")
			break collect
		}
	}

	var workingSharders []string
	for _, s := range responding {
		if maxLFB-s.round <= lfbMaxDrift {
			workingSharders = append(workingSharders, s.sharder)
		} else {
			logging.Debug(fmt.Sprintf("Sharder %s too far behind: LFB %d vs highest %d (drift %d)",
				s.sharder, s.round, maxLFB, maxLFB-s.round))
			h.Fail(s.sharder)
		}
	}

	logging.Debug(fmt.Sprintf("LFB refresh: %d/%d sharders healthy (within %d blocks of LFB %d)",
		len(workingSharders), len(allNodes), lfbMaxDrift, maxLFB))

	h.lfbMu.Lock()
	// Always update the cache — even when workingSharders is empty. Keeping a stale
	// list that includes now-lagging sharders is worse than having an empty cache
	// (which triggers the single-node fallback in HealthyByLFB).
	h.lfbSharders = workingSharders
	h.lfbExpiry = time.Now().Add(lfbCacheTTL)
	h.lfbMu.Unlock()
}
