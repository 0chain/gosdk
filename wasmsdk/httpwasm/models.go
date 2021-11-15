package httpwasm

type Transaction struct {
	Hash              string `json:"hash,omitempty"`
	Version           string `json:"version,omitempty"`
	ClientID          string `json:"client_id,omitempty"`
	PublicKey         string `json:"public_key,omitempty"`
	ToClientID        string `json:"to_client_id,omitempty"`
	ChainID           string `json:"chain_id,omitempty"`
	TransactionData   string `json:"transaction_data,omitempty"`
	Value             int64  `json:"transaction_value,omitempty"`
	Signature         string `json:"signature,omitempty"`
	CreationDate      int64  `json:"creation_date,omitempty"`
	Fee               int64  `json:"transaction_fee,omitempty"`
	TransactionType   int    `json:"transaction_type,omitempty"`
	TransactionOutput string `json:"transaction_output,omitempty"`
	OutputHash        string `json:"txn_output_hash"`
	Status            int    `json:"transaction_status,omitempty"`
}
