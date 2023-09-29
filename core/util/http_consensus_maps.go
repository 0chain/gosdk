package util

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
)

var ErrNilHttpConsensusMaps = errors.New("nil_httpconsensusmaps")

type HttpConsensusMaps struct {
	ConsensusThresh int
	MaxConsensus    int

	WinMap          map[string]json.RawMessage
	WinMapConsensus map[string]int

	WinError string
	WinInfo  string
}

func NewHttpConsensusMaps(consensusThresh int) *HttpConsensusMaps {
	return &HttpConsensusMaps{
		ConsensusThresh: consensusThresh,
		WinMapConsensus: make(map[string]int),
	}
}

func (c *HttpConsensusMaps) GetValue(name string) (json.RawMessage, bool) {
	if c == nil || c.WinMap == nil {
		return nil, false
	}
	v, ok := c.WinMap[name]
	return v, ok
}

func (c *HttpConsensusMaps) Add(statusCode int, respBody string) error {
	if c == nil {
		return ErrNilHttpConsensusMaps
	}
	if statusCode != http.StatusOK {
		c.WinError = respBody
		return nil
	}

	m, hash, err := c.buildMap(respBody)
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

func (c *HttpConsensusMaps) buildMap(respBody string) (map[string]json.RawMessage, string, error) {
	var m map[string]json.RawMessage
	err := json.Unmarshal([]byte(respBody), &m)
	if err != nil {
		return nil, "", err
	}

	keys := make([]string, 0, len(m))

	for k := range m {

		keys = append(keys, k)
	}
	sort.Strings(keys)

	sb := strings.Builder{}

	for _, k := range keys {
		sb.Write(m[k])
		sb.WriteString(":")
	}

	h := sha1.New()
	h.Write([]byte(sb.String()))
	hash := h.Sum(nil)

	return m, hex.EncodeToString(hash), nil
}
