package main

import (
	"errors"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
)

var (
	statusCaches, _ = lru.New[string, *Status](1000)
	statusSync      sync.RWMutex

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
}

func (c *StatusCallback) getOrCreate(lookupHash string) *Status {
	s, ok := statusCaches.Get(lookupHash)
	if !ok {
		s = &Status{}
		statusCaches.Add(lookupHash, s)
	}

	return s
}

func (c *StatusCallback) Started(allocationID, remotePath string, op int, totalBytes int) {
	log.Info("status: Started ", remotePath, " ", totalBytes)
	lookupHash := getLookupHash(allocationID, remotePath)
	s := c.getOrCreate(lookupHash)
	s.Started = true
	s.TotalBytes = totalBytes
	s.LookupHash = lookupHash
}

func (c *StatusCallback) InProgress(allocationID, remotePath string, op int, completedBytes int, data []byte) {
	log.Info("status: InProgress ", remotePath, " ", completedBytes)
	lookupHash := getLookupHash(allocationID, remotePath)
	s := c.getOrCreate(lookupHash)
	s.CompletedBytes = completedBytes
	s.LookupHash = lookupHash
}

func (c *StatusCallback) Error(allocationID string, remotePath string, op int, err error) {
	log.Info("status: Error ", remotePath, " ", err)
	lookupHash := getLookupHash(allocationID, remotePath)
	s := c.getOrCreate(lookupHash)
	s.Error = err.Error()
	s.LookupHash = lookupHash
}

func (c *StatusCallback) Completed(allocationID, remotePath string, filename string, mimetype string, size int, op int) {
	log.Info("status: Completed ", remotePath)
	lookupHash := getLookupHash(allocationID, remotePath)
	s := c.getOrCreate(lookupHash)
	s.Completed = true
	s.LookupHash = lookupHash
}

func (c *StatusCallback) CommitMetaCompleted(request, response string, err error) {
	//c.Callback.CommitMetaCompleted(request, response, err)
}

func (c *StatusCallback) RepairCompleted(filesRepaired int) {
	//c.Callback.RepairCompleted(filesRepaired)
}
