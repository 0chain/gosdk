package zcnbridge

import (
	"fmt"

	"github.com/0chain/gosdk/zcnbridge/http"
	"github.com/0chain/gosdk/zcnbridge/wallet"
	"github.com/0chain/gosdk/zcncore"
)

// Models

type AuthorizerNode struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type AuthorizerNodesResponse struct {
	Nodes []*AuthorizerNode `json:"nodes"`
}

// Rest endpoints

func init() {
	Logger.Init(defaultLogLevel, "0chain-zcnbridge-sdk")
}

// getAuthorizers returns authorizers from smart contract
func getAuthorizers() ([]*AuthorizerNode, error) {
	var (
		authorizers = new(AuthorizerNodesResponse)
		cb          = wallet.NewZCNStatus(authorizers)
		err         error
	)

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
