package zcnbridge

import (
	"fmt"
	"net/url"

	"github.com/machinebox/graphql"
)

func (b *SubgraphConfig) CreateSubgraphClient() (*graphql.Client, error) {
	clientURL, err := url.Parse(b.BlockWorker)
	if err != nil {
		return nil, err
	}

	clientURL.Host = fmt.Sprintf("graphnode.%s", clientURL.Host)
	clientURL.Path = "/subgraphs/name/dex_subgraph"

	return graphql.NewClient(clientURL.String()), nil
}
