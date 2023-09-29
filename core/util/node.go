package util

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/zboxcore/logger"
)

const statSize = 20

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

	i := sort.Search(len(h.nodes), func(i int) bool {
		return h.stats[h.nodes[i]].weight < n.weight
	})
	h.nodes = append(h.nodes[:i], append([]string{n.id}, h.nodes[i:]...)...)
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
	GET_BALANCE = `/v1/client/get/balance?client_id=`
)

func (h *NodeHolder) GetNonceFromSharders(clientID string) (int64, string, error) {
	return h.GetBalanceFieldFromSharders(clientID, "nonce")
}

func (h *NodeHolder) GetBalanceFieldFromSharders(clientID, name string) (int64, string, error) {
	result := make(chan *GetResponse)
	defer close(result)
	// getMinShardersVerify
	numSharders := len(h.Healthy())
	h.queryFromSharders(numSharders, fmt.Sprintf("%v%v", GET_BALANCE, clientID), result)

	consensusMaps := NewHttpConsensusMaps(consensusThresh)

	for i := 0; i < numSharders; i++ {
		rsp := <-result

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

func (h *NodeHolder) queryFromSharders(numSharders int, query string,
	result chan *GetResponse) {

	h.queryFromShardersContext(context.Background(), numSharders, query, result)
}

func (h *NodeHolder) queryFromShardersContext(ctx context.Context, numSharders int,
	query string, result chan *GetResponse) {

	sharders := h.Healthy()
	for _, sharder := range Shuffle(sharders)[:numSharders] {
		go func(sharderurl string) {
			logger.Logger.Info("Query from ", sharderurl+query)
			url := fmt.Sprintf("%v%v", sharderurl, query)
			req, err := NewHTTPGetRequestContext(ctx, url)
			if err != nil {
				logger.Logger.Error(sharderurl, " new get request failed. ", err.Error())
				h.Fail(sharderurl)
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
