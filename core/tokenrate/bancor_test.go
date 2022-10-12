package tokenrate

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBancorJson(t *testing.T) {
	js := ` {
		"data": {
				"dltId": "0xb9EF770B6A5e12E45983C5D80545258aA38F3B78",
				"symbol": "ZCN",
				"decimals": 10,
				"rate": {
						"bnt": "0.271257342312491431",
						"usd": "0.118837",
						"eur": "0.121062",
						"eth": "0.000089243665620809"
				},
				"rate24hAgo": {
						"bnt": "0.273260935543748855",
						"usd": "0.120972",
						"eur": "0.126301",
						"eth": "0.000094001761827049"
				}
		},
		"timestamp": {
				"ethereum": {
						"block": 15644407,
						"timestamp": 1664519843
				}
		}
	}`
	bs := &bancorResponse{}

	err := json.Unmarshal([]byte(js), bs)
	require.Nil(t, err)

	require.Equal(t, 0.118837, bs.Data.Rate["usd"].Value)
	require.Equal(t, 0.120972, bs.Data.Rate24hAgo["usd"].Value)

}
