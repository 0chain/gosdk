package zcnsc

import (
	"encoding/json"

	"github.com/0chain/gosdk/core/common"
)

// MintPayload Payload to submit to ZCN chain `mint` smart contract
type MintPayload struct {
	EthereumTxnID     string                 `json:"ethereum_txn_id"`
	Amount            common.Balance         `json:"amount"`
	Nonce             int64                  `json:"nonce"`
	Signatures        []*AuthorizerSignature `json:"signatures"`
	ReceivingClientID string                 `json:"receiving_client_id"`
}

type AuthorizerSignature struct {
	ID        string `json:"authorizer_id"`
	Signature string `json:"signature"`
}

func (mp *MintPayload) Encode() []byte {
	buff, _ := json.Marshal(mp)
	return buff
}
