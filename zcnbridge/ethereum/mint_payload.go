package ethereum

import "encoding/json"

// MintPayload Payload to submit to the ethereum bridge contract
type MintPayload struct {
	ZCNTxnID   string                 `json:"zcn_txn_id"`
	ClientID   string                 `json:"client_id"`
	Amount     int64                  `json:"amount"`
	Nonce      int64                  `json:"nonce"`
	Signatures []*AuthorizerSignature `json:"signatures"`
}

type AuthorizerSignature struct {
	ID        string `json:"authorizer_id"`
	Signature []byte `json:"signature"`
}

func (mp *MintPayload) Encode() []byte {
	buff, _ := json.Marshal(mp)
	return buff
}
