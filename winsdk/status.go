package main

import (
	"errors"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
)

var (
	statusCaches, _ = lru.New[string, *StatusCallback](100)

	ErrInvalidUploadID   = errors.New("bulkupload: invalid upload id")
	ErrInvalidRemotePath = errors.New("bulkupload: invalid remotePath")
)

type Status struct {
	Started        bool
	TotalBytes     int
	CompletedBytes int
	Error          string
	Completed      bool
	LookupHash     string
}

type StatusCallback struct {
	sync.Mutex
	status map[string]*Status
}

func (c *StatusCallback) Get(remotePath string) *Status {
	c.Lock()
	defer c.Unlock()
	s, ok := c.status[remotePath]
	if ok {
		return s
	}

	return s
}

func (c *StatusCallback) getStatus(remotePath string) *Status {
	s, ok := c.status[remotePath]
	if !ok {
		s = &Status{}
		c.status[remotePath] = s
	}

	return s
}

func (c *StatusCallback) Started(allocationID, remotePath string, op int, totalBytes int) {
	c.Lock()
	defer c.Unlock()
	log.Info("status: Started ", remotePath, " ", totalBytes)
	s := c.getStatus(remotePath)
	s.Started = true
	s.TotalBytes = totalBytes
	s.LookupHash = getLookupHash(allocationID, remotePath)
}

func (c *StatusCallback) InProgress(allocationID, remotePath string, op int, completedBytes int, data []byte) {
	c.Lock()
	defer c.Unlock()
	log.Info("status: InProgress ", remotePath, " ", completedBytes)
	s := c.getStatus(remotePath)
	s.CompletedBytes = completedBytes
	s.LookupHash = getLookupHash(allocationID, remotePath)
}

func (c *StatusCallback) Error(allocationID string, remotePath string, op int, err error) {
	c.Lock()
	defer c.Unlock()
	log.Info("status: Error ", remotePath, " ", err)
	s := c.getStatus(remotePath)
	s.Error = err.Error()
	s.LookupHash = getLookupHash(allocationID, remotePath)
}

func (c *StatusCallback) Completed(allocationID, remotePath string, filename string, mimetype string, size int, op int) {
	c.Lock()
	defer c.Unlock()
	log.Info("status: Completed ", remotePath)
	s := c.getStatus(remotePath)
	s.Completed = true
	s.LookupHash = getLookupHash(allocationID, remotePath)
}

func (c *StatusCallback) CommitMetaCompleted(request, response string, err error) {
	//c.Callback.CommitMetaCompleted(request, response, err)
}

func (c *StatusCallback) RepairCompleted(filesRepaired int) {
	//c.Callback.RepairCompleted(filesRepaired)
}
