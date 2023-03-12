package zcnsc

import "encoding/json"

// BurnPayload Payload to submit to ZCN chain `burn` smart contract
type BurnPayload struct {
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
