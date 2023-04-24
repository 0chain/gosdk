package tokenrate

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/machinebox/graphql"
)

type V2Pair struct {
	ID           string
	VolumeToken0 string
	VolumeToken1 string
	Token0Price  string
	Token1Price  string
	Token0       V2Token
	Token1       V2Token
}

type V2Token struct {
	ID     string
	Symbol string
}

type Query struct {
	ZCN  V2Pair
	USDC V2Pair
}

type uniswapQuoteQuery struct {
}

func (qq *uniswapQuoteQuery) getUSD(ctx context.Context, symbol string) (float64, error) {

	hql := graphql.NewClient("https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v2")

	// make a request
	req := graphql.NewRequest(`
	{
		zcn: pair(id:"0xa6890ac41e3a99a427bef68398bf06119fb5e211"){
			token0 {
				id
				symbol
				totalSupply
			}
			token1 {
				id
				symbol
				totalSupply
			}
			token0Price
			token1Price
			volumeToken0
			volumeToken1
		}
		
		usdc: pair(id:"0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc"){
			token0 {
				id
				symbol
				totalSupply
			}
			token1 {
				id
				symbol
				totalSupply
			}
			token0Price
			token1Price
			volumeToken0
			volumeToken1
		}
	}
`)

	// set header fields
	// req.Header.Set("Cache-Control", "no-cache")
	req.Header.Add("js.fetch:mode", "cors")

	// run it and capture the response
	q := &Query{}
	if err := hql.Run(ctx, req, q); err != nil {
		fmt.Println("uniswap: ", err)
		return 0, err
	}

	switch strings.ToUpper(symbol) {
	case "ZCN":
		ethPerZCN, _ := strconv.ParseFloat(q.ZCN.Token1Price, 64)
		usdcPerETH, _ := strconv.ParseFloat(q.USDC.Token0Price, 64)
		return ethPerZCN * usdcPerETH, nil
	case "ETH":
		usdcPerETH, _ := strconv.ParseFloat(q.USDC.Token0Price, 64)
		return usdcPerETH, nil
	}

	return 0, errors.New("uniswap: quote [" + symbol + "] is unimplemented yet")
}
