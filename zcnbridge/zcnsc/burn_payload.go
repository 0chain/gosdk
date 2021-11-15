package zcnsc

import "encoding/json"

type BurnPayload struct {
	Nonce           int64  `json:"nonce"`
	EthereumAddress string `json:"ethereum_address"`
}

func (bp *BurnPayload) Encode() []byte {
	buff, _ := json.Marshal(bp)
	return buff
}

func (bp *BurnPayload) Decode(input []byte) error {
	err := json.Unmarshal(input, bp)
	return err
}
