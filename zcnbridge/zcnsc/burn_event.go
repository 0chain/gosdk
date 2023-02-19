package zcnsc

// BurnEvent represents WZCN burn event
type BurnEvent struct {
	Burneds []struct {
		TransactionHash string `json:"transactionHash"`
		Nonce           int    `json:"nonce"`
	} `json:"burneds"`
}
