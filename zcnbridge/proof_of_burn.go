package zcnbridge

type proofOfBurn struct {
	TxnID             string `json:"ethereum_txn_id"`
	Amount            int64  `json:"amount"`
	ReceivingClientID string `json:"receiving_client_id"` // 0ZCN address
	Nonce             int64  `json:"nonce"`
	Signature         string `json:"signature"`
}
