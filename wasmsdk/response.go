package main

import "github.com/0chain/gosdk/core/transaction"

type DownloadResponse struct {
	FileName string                   `json:"fileName,omitempty"`
	Url      string                   `json:"url,omitempty"`
	Txn      *transaction.Transaction `json:"txn,omitempty"`
	Error    string                   `json:"error,omitempty"`
}
