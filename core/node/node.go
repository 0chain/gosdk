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

const consensusThresh = 25
const (
	GET_BALANCE        = `/v1/client/get/balance?client_id=`
	CURRENT_ROUND      = "/v1/current-round"
	GET_BLOCK_INFO     = `/v1/block/get?`
	GET_HARDFORK_ROUND = `/v1/screst/6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d9/hardfork?name=`
)

func (h *NodeHolder) GetNonceFromSharders(clientID string) (int64, string, error) {
	return h.GetBalanceFieldFromSharders(clientID, "nonce")
}

func (h *NodeHolder) GetBalanceFieldFromSharders(clientID, name string) (int64, string, error) {
	result := make(chan *util.GetResponse)
	defer close(result)
	// getMinShardersVerify
	numSharders := len(h.Healthy())
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

	sharders := h.Healthy()

	for _, sharder := range util.Shuffle(sharders)[:numSharders] {
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
		}(sharder)
	}
}

func (h *NodeHolder) GetBlockByRound(ctx context.Context, numSharders int, round int64) (b *block.Block, err error) {

	var result = make(chan *util.GetResponse, numSharders)
	defer close(result)

	numSharders = len(h.Healthy()) // overwrite, use all
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
