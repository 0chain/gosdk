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

var ErrNilHttpConsensus = errors.New("nil_httpconsensus")

type HttpConsensusValue interface {
	map[string]json.RawMessage | json.RawMessage
}

type HttpConsensus[T HttpConsensusValue] struct {
	ConsensusThresh int
	MaxConsensus    int

	WinValue          T
	WinValueConsensus map[string]int

	WinError string
	WinInfo  string
}

func NewHttpConsensus[T HttpConsensusValue](consensusThresh int) *HttpConsensus[T] {
	return &HttpConsensus[T]{
		ConsensusThresh:   consensusThresh,
		WinValueConsensus: make(map[string]int),
	}
}

func (c *HttpConsensus[T]) GetValue(name string) (json.RawMessage, bool) {
	if c == nil || c.WinValue == nil {
		return nil, false
	}
	v, ok := c.WinMap[name]
	return v, ok
}

func (c *HttpConsensus[T]) Add(statusCode int, respBody string) error {
	if c == nil {
		return ErrNilHttpConsensus
	}
	if statusCode != http.StatusOK {
		c.WinError = respBody
		return nil
	}

	m, hash, err := c.build(respBody)
	if err != nil {
		return err
	}

	c.WinValueConsensus[hash]++

	if c.WinValueConsensus[hash] > c.MaxConsensus {
		c.MaxConsensus = c.WinValueConsensus[hash]
		c.WinMap = m
		c.WinInfo = respBody
	}

	return nil
}

func (c *HttpConsensus[T]) build(respBody string) (T, string, error) {
	var m T
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
