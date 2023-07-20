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
}

type StatusCallback struct {
	sync.Mutex
	status map[string]*Status
}

func (c *StatusCallback) getStatus(remotePath string) *Status {
	c.Lock()
	defer c.Unlock()

	s, ok := c.status[remotePath]
	if !ok {
		s = &Status{}
		c.status[remotePath] = s
	}

	return s
}

func (c *StatusCallback) Started(allocationId, remotePath string, op int, totalBytes int) {
	log.Info("status: Started ", remotePath, " ", totalBytes)
	s := c.getStatus(remotePath)
	s.Started = true
	s.TotalBytes = totalBytes
}

func (c *StatusCallback) InProgress(allocationId, remotePath string, op int, completedBytes int, data []byte) {
	log.Info("status: InProgress ", remotePath, " ", completedBytes)
	s := c.getStatus(remotePath)
	s.CompletedBytes = completedBytes
}

func (c *StatusCallback) Error(allocationID string, remotePath string, op int, err error) {
	log.Info("status: Error ", remotePath, " ", err)
	s := c.getStatus(remotePath)
	s.Error = err.Error()
}

func (c *StatusCallback) Completed(allocationId, remotePath string, filename string, mimetype string, size int, op int) {
	log.Info("status: Completed ", remotePath)
	s := c.getStatus(remotePath)
	s.Completed = true
}

func (c *StatusCallback) CommitMetaCompleted(request, response string, err error) {
	//c.Callback.CommitMetaCompleted(request, response, err)
}

func (c *StatusCallback) RepairCompleted(filesRepaired int) {
	//c.Callback.RepairCompleted(filesRepaired)
}
