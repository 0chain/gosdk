package zcnbridge

import (
	"fmt"

	"github.com/0chain/gosdk/core/common"

	"github.com/0chain/gosdk/zcnbridge/http"
	"github.com/0chain/gosdk/zcnbridge/wallet"
	"github.com/0chain/gosdk/zcncore"
)

// Models

type AuthorizerResponse struct {
	AuthorizerID string `json:"id"`
	URL          string `json:"url"`

	// Configuration
	Fee common.Balance `json:"fee"`

	// Geolocation
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`

	// Stats
	LastHealthCheck int64 `json:"last_health_check"`

	// stake_pool_settings
	DelegateWallet string         `json:"delegate_wallet"`
	MinStake       common.Balance `json:"min_stake"`
	MaxStake       common.Balance `json:"max_stake"`
	NumDelegates   int            `json:"num_delegates"`
	ServiceCharge  float64        `json:"service_charge"`
}

type AuthorizerNodesResponse struct {
	Nodes []*AuthorizerNode `json:"nodes"`
}

type AuthorizerNode struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// Rest endpoints

// getAuthorizers returns authorizers from smart contract
func getAuthorizers() ([]*AuthorizerNode, error) {
	var (
		authorizers = new(AuthorizerNodesResponse)
		cb          = wallet.NewZCNStatus(authorizers)
		err         error
	)

	cb.Begin()

	if err = GetAuthorizers(cb); err != nil {
		return nil, err
	}

	if err = cb.Wait(); err != nil {
		return nil, err
	}

	if len(authorizers.Nodes) == 0 {
		fmt.Println("no authorizers found")
		return nil, err
	}

	return authorizers.Nodes, nil
}

// GetAuthorizer returned authorizer by ID
func GetAuthorizer(id string, cb zcncore.GetInfoCallback) (err error) {
	err = zcncore.CheckConfig()
	if err != nil {
		return err
	}

	go http.MakeSCRestAPICall(
		zcncore.OpZCNSCGetAuthorizer,
		http.PathGetAuthorizer,
		http.Params{
			"id": id,
		},
		cb,
	)

	return
}

// GetAuthorizers Returns authorizers
func GetAuthorizers(cb zcncore.GetInfoCallback) (err error) {
	err = zcncore.CheckConfig()
	if err != nil {
		return err
	}
	go http.MakeSCRestAPICall(zcncore.OpZCNSCGetAuthorizerNodes, http.PathGetAuthorizerNodes, nil, cb)
	return
}

// GetGlobalConfig Returns global config
func GetGlobalConfig(cb zcncore.GetInfoCallback) (err error) {
	err = zcncore.CheckConfig()
	if err != nil {
		return err
	}
	go http.MakeSCRestAPICall(zcncore.OpZCNSCGetGlobalConfig, http.PathGetGlobalConfig, nil, cb)
	return
}
