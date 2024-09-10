package zcnbridge

import (
	"encoding/json"
	"fmt"
	coreHttp "github.com/0chain/gosdk/core/client"
	"github.com/0chain/gosdk/core/common"

	"github.com/0chain/gosdk/zcncore"
)

const (
	// SCRestAPIPrefix represents base URL path to execute smart contract rest points.
	SCRestAPIPrefix        = "v1/screst/"
	RestPrefix             = SCRestAPIPrefix + zcncore.ZCNSCSmartContractAddress
	PathGetAuthorizerNodes = "/getAuthorizerNodes?active=%t"
	PathGetGlobalConfig    = "/getGlobalConfig"
	PathGetAuthorizer      = "/getAuthorizer"
)

// AuthorizerResponse represents the response of the request to get authorizer info from the sharders.
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

// AuthorizerNodesResponse represents the response of the request to get authorizers
type AuthorizerNodesResponse struct {
	Nodes []*AuthorizerNode `json:"nodes"`
}

// AuthorizerNode represents an authorizer node
type AuthorizerNode struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// Rest endpoints

// getAuthorizers returns authorizers from smart contract
func getAuthorizers(active bool) ([]*AuthorizerNode, error) {
	var (
		authorizers = new(AuthorizerNodesResponse)
		err         error
		res         []byte
	)

	if res, err = GetAuthorizers(active); err != nil {
		return nil, err
	}

	err = json.Unmarshal(res, authorizers)
	if err != nil {
		return nil, err
	}

	if len(authorizers.Nodes) == 0 {
		fmt.Println("no authorizers found")
		return nil, err
	}

	return authorizers.Nodes, nil
}

// GetAuthorizer returned authorizer information from Züs Blockchain by the ID
//   - id is the authorizer ID
//   - cb is the callback function to handle the response asynchronously
func GetAuthorizer(id string) (res []byte, err error) {
	err = zcncore.CheckConfig()
	if err != nil {
		return nil, err
	}

	return coreHttp.MakeSCRestAPICall(zcncore.ZCNSCSmartContractAddress, PathGetAuthorizer, zcncore.Params{
		"id": id,
	}, nil)
}

// GetAuthorizers Returns all or only active authorizers
//   - active is the flag to get only active authorizers
//   - cb is the callback function to handle the response asynchronously
func GetAuthorizers(active bool) (res []byte, err error) {
	err = zcncore.CheckConfig()
	if err != nil {
		return nil, err
	}
	return coreHttp.MakeSCRestAPICall(zcncore.ZCNSCSmartContractAddress, fmt.Sprintf(PathGetAuthorizerNodes, active), nil, nil)
}

// GetGlobalConfig Returns global config
//   - cb is the callback function to handle the response asynchronously
func GetGlobalConfig() (res []byte, err error) {
	err = zcncore.CheckConfig()
	if err != nil {
		return nil, err
	}
	return coreHttp.MakeSCRestAPICall(zcncore.ZCNSCSmartContractAddress, PathGetGlobalConfig, nil, nil)
}
