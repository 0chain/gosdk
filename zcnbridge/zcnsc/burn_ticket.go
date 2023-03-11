package zcnsc

// BurnTicket represents WZCN burn ticket details
type BurnTicket struct {
	TransactionHash string `json:"transactionHash"`
	Nonce           string `json:"nonce"`
}
