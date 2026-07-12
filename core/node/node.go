// Provides functions and data structures to interact with the system nodes in the context of the blockchain network.
package node

import (
	"context"
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/block"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zboxcore/logger"
	"github.com/ethereum/go-ethereum/common/math"
)

const statSize = 20
const defaultTimeout = 5 * time.Second

type NodeHolder struct {
	consensus int
	guard     sync.Mutex
	stats     map[string]*Node
	nodes     []string

	lfbMu       sync.RWMutex
	lfbSharders []string
	lfbExpiry   time.Time
	lfbUpdating atomic.Bool
}

type Node struct {
	id     string
	weight int64
	stats  []int
}

func NewHolder(nodes []string, consensus int) *NodeHolder {
	if len(nodes) < consensus {
		panic("consensus is not correct")
	}
	holder := NodeHolder{consensus: consensus, stats: make(map[string]*Node)}

	for _, n := range nodes {
		holder.nodes = append(holder.nodes, n)
		holder.stats[n] = NewNode(n)
	}
	return &holder
}

func NewNode(id string) *Node {
	return &Node{
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
	// *Node and panics — fatal in the wasm SDK (the whole Go runtime exits).
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

// HealthyVerify returns the sharder set to use for transaction confirmation (and
// other reads that must reflect the chain tip). It prefers the LFB-filtered
// in-sync set (HealthyByLFB) so a sharder that has fallen behind the chain tip
// cannot sink a confirmation while in-sync sharders exist — the weight-ranked
// Healthy() set is sync-unaware and will happily keep an always-reachable but
// lagging sharder. When LFB data is unavailable HealthyByLFB degrades to a
// single highest-weighted node; in that case (len <= 1) we fall back to the
// broader Healthy() set (then All()) so confirmation always has multiple
// sharders to try and the result is never empty.
func (h *NodeHolder) HealthyVerify() []string {
	if lfb := h.HealthyByLFB(); len(lfb) > 1 {
		return lfb
	}
	if hh := h.Healthy(); len(hh) > 0 {
		return hh
	}
	return h.All()
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
				logger.Logger.Debug("Sharder LFB check failed: " + r.sharder + ": " + r.err.Error())
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
			logger.Logger.Debug("LFB refresh deadline hit — proceeding with responders so far")
			break collect
		}
	}

	var workingSharders []string
	for _, s := range responding {
		if maxLFB-s.round <= lfbMaxDrift {
			workingSharders = append(workingSharders, s.sharder)
		} else {
			logger.Logger.Debug(fmt.Sprintf("Sharder %s too far behind: LFB %d vs highest %d (drift %d)",
				s.sharder, s.round, maxLFB, maxLFB-s.round))
			h.Fail(s.sharder)
		}
	}

	logger.Logger.Debug(fmt.Sprintf("LFB refresh: %d/%d sharders healthy (within %d blocks of LFB %d)",
		len(workingSharders), len(allNodes), lfbMaxDrift, maxLFB))

	h.lfbMu.Lock()
	// Always update the cache — even when workingSharders is empty. Keeping a stale
	// list that includes now-lagging sharders is worse than having an empty cache
	// (which triggers the single-node fallback in HealthyByLFB).
	h.lfbSharders = workingSharders
	h.lfbExpiry = time.Now().Add(lfbCacheTTL)
	h.lfbMu.Unlock()
}

const consensusThresh = 25
const (
	GET_BALANCE        = `/v1/client/get/balance?client_id=`
	CURRENT_ROUND      = "/v1/current-round"
	GET_BLOCK_INFO     = `/v1/block/get?`
	GET_HARDFORK_ROUND = `/v1/screst/6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d9/hardfork?name=`
)

// GetNonceFromSharders returns the highest nonce for clientID across all LFB-healthy
// sharders. Using the maximum protects against stale nonces from lagging or stuck sharders.
func (h *NodeHolder) GetNonceFromSharders(clientID string) (int64, string, error) {
	sharders := h.HealthyByLFB()
	if len(sharders) == 0 {
		return 0, "", errors.New("no_sharders", "no healthy sharders available")
	}

	type result struct {
		nonce int64
		info  string
		err   error
	}

	results := make(chan result, len(sharders))

	for _, sharder := range sharders {
		go func(s string) {
			url := fmt.Sprintf("%s%s", s, GET_BALANCE+clientID)
			req, err := util.NewHTTPGetRequest(url)
			if err != nil {
				results <- result{err: err}
				return
			}
			resp, err := req.Get()
			if err != nil {
				h.Fail(s)
				results <- result{err: err}
				return
			}
			if resp.StatusCode != http.StatusOK {
				h.Fail(s)
				results <- result{err: fmt.Errorf("status %d: %s", resp.StatusCode, resp.Body)}
				return
			}
			h.Success(s)
			var bal struct {
				Nonce int64  `json:"nonce"`
				Txn   string `json:"txn"`
			}
			if err := json.Unmarshal([]byte(resp.Body), &bal); err != nil {
				results <- result{err: err}
				return
			}
			results <- result{nonce: bal.Nonce, info: bal.Txn}
		}(sharder)
	}

	maxNonce := int64(-1)
	var maxInfo string
	var lastErr error
	for i := 0; i < len(sharders); i++ {
		r := <-results
		if r.err != nil {
			lastErr = r.err
			continue
		}
		if r.nonce > maxNonce {
			maxNonce = r.nonce
			maxInfo = r.info
		}
	}

	if maxNonce < 0 {
		if lastErr != nil {
			return 0, "", lastErr
		}
		return 0, "", errors.New("no_nonce", "could not get nonce from any sharder")
	}
	return maxNonce, maxInfo, nil
}

func (h *NodeHolder) GetBalanceFieldFromSharders(clientID, name string) (int64, string, error) {
	result := make(chan *util.GetResponse)
	defer close(result)
	// getMinShardersVerify
	numSharders := len(h.HealthyVerify())
	h.QueryFromSharders(numSharders, fmt.Sprintf("%v%v", GET_BALANCE, clientID), result)

	consensusMaps := util.NewHttpConsensusMaps(consensusThresh)

	for i := 0; i < numSharders; i++ {
		rsp := <-result
		if rsp == nil {
			logger.Logger.Error("nil response")
			continue
		}

		logger.Logger.Debug(rsp.Url, rsp.Status)
		if rsp.StatusCode != http.StatusOK {
			logger.Logger.Error(rsp.Body)

		} else {
			logger.Logger.Debug(rsp.Body)
		}

		if err := consensusMaps.Add(rsp.StatusCode, rsp.Body); err != nil {
			logger.Logger.Error(rsp.Body)
		}
	}

	rate := consensusMaps.MaxConsensus * 100 / numSharders
	if rate < consensusThresh {
		if strings.TrimSpace(consensusMaps.WinError) == `{"error":"value not present"}` {
			return 0, consensusMaps.WinError, nil
		}
		return 0, consensusMaps.WinError, errors.New("", "get balance failed. consensus not reached")
	}

	winValue, ok := consensusMaps.GetValue(name)
	if ok {
		winBalance, err := strconv.ParseInt(string(winValue), 10, 64)
		if err != nil {
			return 0, "", fmt.Errorf("get balance failed. %w", err)
		}

		return winBalance, consensusMaps.WinInfo, nil
	}

	return 0, consensusMaps.WinInfo, errors.New("", "get balance failed. balance field is missed")
}

func (h *NodeHolder) QueryFromSharders(numSharders int, query string,
	result chan *util.GetResponse) {

	h.QueryFromShardersContext(context.Background(), numSharders, query, result)
}

func (h *NodeHolder) QueryFromShardersContext(ctx context.Context, numSharders int,
	query string, result chan *util.GetResponse) {

	// Query the in-sync (LFB-filtered) sharder set so a read never lands only on
	// sharders that have fallen behind the chain tip. HealthyVerify falls back to
	// the weight-ranked Healthy() set when LFB data is unavailable, so it is never
	// empty. Callers pass numSharders (and wait for that many responses); if the
	// in-sync set is smaller, send nil for the shortfall so the caller's response
	// loop always completes, and never slice past the available set (no panic).
	sharders := util.Shuffle(h.HealthyVerify())

	for i := 0; i < numSharders; i++ {
		if i >= len(sharders) {
			result <- nil
			continue
		}
		go func(sharderurl string) {
			logger.Logger.Info("Query from ", sharderurl+query)
			url := fmt.Sprintf("%v%v", sharderurl, query)
			timeout, cancelFunc := context.WithTimeout(ctx, defaultTimeout)
			defer cancelFunc()

			req, err := util.NewHTTPGetRequestContext(timeout, url)
			if err != nil {
				logger.Logger.Error(sharderurl, " new get request failed. ", err.Error())
				h.Fail(sharderurl)
				result <- nil
				return
			}
			res, err := req.Get()
			if err != nil {
				logger.Logger.Error(sharderurl, " get error. ", err.Error())
			}

			if res.StatusCode > http.StatusBadRequest {
				h.Fail(sharderurl)
			} else {
				h.Success(sharderurl)
			}

			result <- res
		}(sharders[i])
	}
}

func (h *NodeHolder) GetBlockByRound(ctx context.Context, numSharders int, round int64) (b *block.Block, err error) {

	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	numSharders = len(h.HealthyVerify()) // overwrite, use all
	h.QueryFromShardersContext(ctx, numSharders,
		fmt.Sprintf("%sround=%d&content=full,header", GET_BLOCK_INFO, round),
		result)

	var (
		maxConsensus   int
		roundConsensus = make(map[string]int)
	)

	type respObj struct {
		Block  *block.Block  `json:"block"`
		Header *block.Header `json:"header"`
	}

	for i := 0; i < numSharders; i++ {
		var rsp = <-result
		if rsp == nil {
			logger.Logger.Error("nil response")
			continue
		}
		logger.Logger.Debug(rsp.Url, rsp.Status)

		if rsp.StatusCode != http.StatusOK {
			logger.Logger.Error(rsp.Body)
			continue
		}

		var respo respObj
		if err = json.Unmarshal([]byte(rsp.Body), &respo); err != nil {
			logger.Logger.Error("block parse error: ", err)
			err = nil
			continue
		}

		if respo.Block == nil {
			logger.Logger.Debug(rsp.Url, "no block in response:", rsp.Body)
			continue
		}

		if respo.Header == nil {
			logger.Logger.Debug(rsp.Url, "no block header in response:", rsp.Body)
			continue
		}

		if respo.Header.Hash != string(respo.Block.Hash) {
			logger.Logger.Debug(rsp.Url, "header and block hash mismatch:", rsp.Body)
			continue
		}

		b = respo.Block
		b.Header = respo.Header

		var h = encryption.FastHash([]byte(b.Hash))
		if roundConsensus[h]++; roundConsensus[h] > maxConsensus {
			maxConsensus = roundConsensus[h]
		}
	}

	if maxConsensus == 0 {
		return nil, errors.New("", "round info not found")
	}

	return
}

func (h *NodeHolder) GetRoundFromSharders() (int64, error) {

	sharders := h.Healthy()
	if len(sharders) == 0 {
		return 0, stdErrors.New("get round failed. no sharders")
	}

	result := make(chan *util.GetResponse, len(sharders))

	var numSharders = len(sharders)
	// use 5 sharders to get round
	if numSharders > 5 {
		numSharders = 5
	}

	h.QueryFromSharders(numSharders, fmt.Sprintf("%v", CURRENT_ROUND), result)

	const consensusThresh = float32(25.0)

	var rounds []int64

	consensus := int64(0)
	roundMap := make(map[int64]int64)

	round := int64(0)

	waitTimeC := time.After(10 * time.Second)
	for i := 0; i < numSharders; i++ {
		select {
		case <-waitTimeC:
			return 0, stdErrors.New("get round failed. consensus not reached")
		case rsp := <-result:
			if rsp == nil {
				logger.Logger.Error("nil response")
				continue
			}
			if rsp.StatusCode != http.StatusOK {
				continue
			}

			var respRound int64
			err := json.Unmarshal([]byte(rsp.Body), &respRound)

			if err != nil {
				continue
			}

			rounds = append(rounds, respRound)

			sort.Slice(rounds, func(i, j int) bool {
				return false
			})

			medianRound := rounds[len(rounds)/2]

			roundMap[medianRound]++

			if roundMap[medianRound] > consensus {

				consensus = roundMap[medianRound]
				round = medianRound
				rate := consensus * 100 / int64(numSharders)

				if rate >= int64(consensusThresh) {
					return round, nil
				}
			}
		}
	}

	return round, nil
}

func (h *NodeHolder) GetHardForkRound(hardFork string) (int64, error) {
	sharders := h.Healthy()
	if len(sharders) == 0 {
		return 0, stdErrors.New("get round failed. no sharders")
	}

	result := make(chan *util.GetResponse, len(sharders))

	var numSharders = len(sharders)
	// use 5 sharders to get round
	if numSharders > 5 {
		numSharders = 5
	}

	h.QueryFromSharders(numSharders, fmt.Sprintf("%s%s", GET_HARDFORK_ROUND, hardFork), result)

	const consensusThresh = float32(25.0)

	var rounds []int64

	consensus := int64(0)
	roundMap := make(map[int64]int64)
	// If error then set it to max int64
	round := int64(math.MaxInt64)

	waitTimeC := time.After(10 * time.Second)
	for i := 0; i < numSharders; i++ {
		select {
		case <-waitTimeC:
			return 0, stdErrors.New("get round failed. consensus not reached")
		case rsp := <-result:
			if rsp == nil {
				logger.Logger.Error("nil response")
				continue
			}
			if rsp.StatusCode != http.StatusOK {
				continue
			}

			var respRound int64
			var objmap map[string]string
			err := json.Unmarshal([]byte(rsp.Body), &objmap)
			if err != nil {
				continue
			}

			str := string(objmap["round"])
			respRound, err = strconv.ParseInt(str, 10, 64)
			if err != nil {
				continue
			}

			rounds = append(rounds, respRound)

			sort.Slice(rounds, func(i, j int) bool {
				return false
			})

			medianRound := rounds[len(rounds)/2]

			roundMap[medianRound]++

			if roundMap[medianRound] > consensus {

				consensus = roundMap[medianRound]
				round = medianRound
				rate := consensus * 100 / int64(numSharders)

				if rate >= int64(consensusThresh) {
					return round, nil
				}
			}
		}
	}

	return round, nil
}
