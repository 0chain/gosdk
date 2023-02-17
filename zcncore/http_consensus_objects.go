package zcncore

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
)

var ErrNilHttpConsensusObjects = errors.New("nil_httpconsensusmaps")

type HttpConsensusObjects struct {
	ConsensusThresh int
	MaxConsensus    int

	WinMap          json.RawMessage
	WinMapConsensus map[string]int

	WinError string
	WinInfo  string
}

func NewHttpConsensusObjects(consensusThresh int) *HttpConsensusObjects {
	return &HttpConsensusObjects{
		ConsensusThresh: consensusThresh,
		WinMapConsensus: make(map[string]int),
	}
}

func (c *HttpConsensusObjects) GetValue() (json.RawMessage, bool) {
	if c == nil {
		return nil, false
	}
	return c.WinMap, true
}

func (c *HttpConsensusObjects) Add(statusCode int, respBody string) error {
	if c == nil {
		return ErrNilHttpConsensusMaps
	}
	if statusCode != http.StatusOK {
		c.WinError = respBody
		return nil
	}

	m, hash, err := c.buildObject(respBody)
	if err != nil {
		return err
	}

	c.WinMapConsensus[hash]++

	if c.WinMapConsensus[hash] > c.MaxConsensus {
		c.MaxConsensus = c.WinMapConsensus[hash]
		c.WinMap = m
		c.WinInfo = respBody
	}

	return nil
}

func (c *HttpConsensusObjects) buildObject(respBody string) (json.RawMessage, string, error) {
	var m json.RawMessage
	err := json.Unmarshal([]byte(respBody), &m)
	if err != nil {
		return nil, "", err
	}

	keys := make([]int, 0, len(m))

	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	sb := strings.Builder{}

	for _, k := range keys {
		sb.WriteByte(m[k])
		sb.WriteString(":")
	}

	h := sha1.New()
	h.Write([]byte(sb.String()))
	hash := h.Sum(nil)

	return m, hex.EncodeToString(hash), nil
}
