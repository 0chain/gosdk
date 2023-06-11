package main

import (
	"github.com/0chain/gosdk/core/transaction"
)

type StatusCallbackMocked interface {
	Started(allocationId, filePath string, op int, totalBytes int)
	InProgress(allocationId, filePath string, op int, completedBytes int, data []byte)
	Error(allocationID string, filePath string, op int, err error)
	Completed(allocationId, filePath string, filename string, mimetype string, size int, op int)
	CommitMetaCompleted(request, response string, err error)
	RepairCompleted(filesRepaired int)
}

type StatusCallbackWrapped struct {
	Callback StatusCallbackMocked
}

func (c *StatusCallbackWrapped) Started(allocationId, filePath string, op int, totalBytes int) {
	c.Callback.Started(allocationId, filePath, op, totalBytes)
}

func (c *StatusCallbackWrapped) InProgress(allocationId, filePath string, op int, completedBytes int, data []byte) {
	c.Callback.InProgress(allocationId, filePath, op, completedBytes, data)
}

func (c *StatusCallbackWrapped) Error(allocationID string, filePath string, op int, err error) {
	c.Callback.Error(allocationID, filePath, op, err)
}

func (c *StatusCallbackWrapped) Completed(allocationId, filePath string, filename string, mimetype string, size int, op int) {
	c.Callback.Completed(allocationId, filePath, filename, mimetype, size, op)
}

func (c *StatusCallbackWrapped) CommitMetaCompleted(request, response string, txn *transaction.Transaction, err error) {
	c.Callback.CommitMetaCompleted(request, response, err)
}

func (c *StatusCallbackWrapped) RepairCompleted(filesRepaired int) {
	c.Callback.RepairCompleted(filesRepaired)
}
