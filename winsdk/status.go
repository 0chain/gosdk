//go:build windows
// +build windows

package main

import (
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
)

var (
	statusUpload, _   = lru.New[string, *Status](1000)
	statusDownload, _ = lru.New[string, *Status](1000)
)

type Status struct {
	Started        bool
	TotalBytes     int
	CompletedBytes int
	Error          string
	Completed      bool
	LookupHash     string
	wg             *sync.WaitGroup
}

type StatusCallback struct {
	key   string
	items *lru.Cache[string, *Status]
}

func NewStatusBar(items *lru.Cache[string, *Status], key string) *StatusCallback {
	sc := &StatusCallback{
		key:   key,
		items: items,
	}

	return sc
}

func (c *StatusCallback) getStatus(lookupHash string) *Status {

	if len(c.key) == 0 {
		s, ok := c.items.Get(lookupHash)

		if !ok {
			s = &Status{}
			c.items.Add(lookupHash, s)
		}
		return s
	}

	s, ok := c.items.Get(c.key)

	if !ok {
		s = &Status{}
		c.items.Add(c.key, s)
	}
	return s
}

func (c *StatusCallback) Started(allocationID, remotePath string, op int, totalBytes int) {
	lookupHash := getLookupHash(allocationID, remotePath)
	log.Info("status: Started ", remotePath, " ", totalBytes, " ", lookupHash)
	s := c.getStatus(lookupHash)
	s.Started = true
	s.TotalBytes = totalBytes
	s.LookupHash = lookupHash
}

func (c *StatusCallback) InProgress(allocationID, remotePath string, op int, completedBytes int, data []byte) {
	lookupHash := getLookupHash(allocationID, remotePath)
	log.Info("status: InProgress ", remotePath, " ", completedBytes, " ", lookupHash)
	s := c.getStatus(lookupHash)
	s.CompletedBytes = completedBytes
	s.LookupHash = lookupHash
	if completedBytes >= s.TotalBytes {
		s.Completed = true
		if s.wg != nil {
			s.wg.Done()
			s.wg = nil
		}

	}
}

func (c *StatusCallback) Error(allocationID string, remotePath string, op int, err error) {
	lookupHash := getLookupHash(allocationID, remotePath)
	log.Info("status: Error ", remotePath, " ", err, " ", lookupHash)
	s := c.getStatus(lookupHash)
	s.Error = err.Error()
	s.LookupHash = lookupHash
	if s.wg != nil {
		s.wg.Done()
		s.wg = nil
	}
}

func (c *StatusCallback) Completed(allocationID, remotePath string, filename string, mimetype string, size int, op int) {
	lookupHash := getLookupHash(allocationID, remotePath)
	log.Info("status: Completed ", remotePath, " ", lookupHash)
	s := c.getStatus(lookupHash)
	s.Completed = true
	s.LookupHash = lookupHash
	s.CompletedBytes = s.TotalBytes
	if s.wg != nil {
		s.wg.Done()
		s.wg = nil
	}
}

func (c *StatusCallback) CommitMetaCompleted(request, response string, err error) {
	//c.Callback.CommitMetaCompleted(request, response, err)
}

func (c *StatusCallback) RepairCompleted(filesRepaired int) {
	//c.Callback.RepairCompleted(filesRepaired)
}
