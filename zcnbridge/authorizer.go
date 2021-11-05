package zcnbridge

import (
	"encoding/json"

	"github.com/0chain/gosdk/zcnbridge/http"
	"github.com/pkg/errors"
)

type Authorizer struct {
}

// AddAuthorizer Adds new authorizer to the chain with ochain smart contract function
func AddAuthorizer() {
}

type AuthorizerNodes struct {
	NodeMap map[string]*AuthorizerNode `json:"node_map"`
}

type AuthorizerNode struct {
	ID        string `json:"id"`
	PublicKey string `json:"public_key"`
	URL       string `json:"url"`
}

// GetAuthorizers Returns authorizers
func GetAuthorizers() (*AuthorizerNodes, error) {
	resp, err := http.MakeSCRestAPICall(http.GetAuthorizersPath, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error requesting authorizers")
	}
	if len(resp) == 0 {
		return nil, errors.New("empty response")
	}

	an := &AuthorizerNodes{}

	if err = json.Unmarshal(resp, &an); err != nil {
		return nil, errors.Wrap(err, "error decoding response:")
	}

	return an, nil
}
