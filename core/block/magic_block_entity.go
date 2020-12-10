package block

import "time"

type Node struct {
	ID           string `yaml:"id" json:"id"`
	Version      string `yaml:"version" json:"version"`
	CreationDate int64  `json:"creation_date"`
	PublicKey    string `yaml:"public_key" json:"public_key"`
	PrivateKey   string `yaml:"private_key" json:"-"`
	N2NHost      string `yaml:"n2n_ip" json:"n2n_host"`
	Host         string `yaml:"public_ip" json:"host"`
	Port         int    `yaml:"port" json:"port"`
	Path         string `yaml:"path" json:"path"`
	Type         int    `json:"type"`
	Description  string `yaml:"description" json:"description"`
	SetIndex     int    `yaml:"set_index" json:"set_index"`
	Status       int    `json:"status"`
	Info         Info   `json:"info"`
}

type Info struct {
	BuildTag                string        `json:"build_tag"`
	StateMissingNodes       int64         `json:"state_missing_nodes"`
	MinersMedianNetworkTime time.Duration `json:"miners_median_network_time"`
	AvgBlockTxns            int           `json:"avg_block_txns"`
}

type NodePool struct {
	Type  int             `json:"type"`
	Nodes map[string]Node `json:"nodes"`
}

type GroupSharesOrSigns struct {
	Shares map[string]*ShareOrSigns `json:"shares"`
}

type ShareOrSigns struct {
	ID           string                  `json:"id"`
	ShareOrSigns map[string]*DKGKeyShare `json:"share_or_sign"`
}

type DKGKeyShare struct {
	ID      string `json:"id"`
	Message string `json:"message"`
	Share   string `json:"share"`
	Sign    string `json:"sign"`
}

type Mpks struct {
	Mpks map[string]*MPK
}

type MPK struct {
	ID  string
	Mpk []string
}

type MagicBlock struct {
	Hash                   string              `json:"hash"`
	PreviousMagicBlockHash string              `json:"previous_hash"`
	MagicBlockNumber       int64               `json:"magic_block_number"`
	StartingRound          int64               `json:"starting_round"`
	Miners                 *NodePool           `json:"miners"`   //this is the pool of miners participating in the blockchain
	Sharders               *NodePool           `json:"sharders"` //this is the pool of sharders participaing in the blockchain
	ShareOrSigns           *GroupSharesOrSigns `json:"share_or_signs"`
	Mpks                   *Mpks               `json:"mpks"`
	T                      int                 `json:"t"`
	K                      int                 `json:"k"`
	N                      int                 `json:"n"`
}
