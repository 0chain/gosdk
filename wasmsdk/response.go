package main

import "github.com/0chain/gosdk/core/transaction"

type FileCommandResponse struct {
	CommandStatus bool                     `json:"commandStatus,omitempty"`
	CommitStatus  bool                     `json:"commitStatus,omitempty"`
	CommitTxn     *transaction.Transaction `json:"commitTxn,omitempty"`
	Error         string                   `json:"error,omitempty"`
}

type DownloadCommandResponse struct {
	CommandSuccess bool                     `json:"commandSuccess,omitempty"`
	CommitSuccess  bool                     `json:"commitSuccess,omitempty"`
	CommitTxn      *transaction.Transaction `json:"commitTxn,omitempty"`
	Error          string                   `json:"error,omitempty"`

	FileName string `json:"fileName,omitempty"`
	Url      string `json:"url,omitempty"`
}
