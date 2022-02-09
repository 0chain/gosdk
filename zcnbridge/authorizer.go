package zcnbridge

import (
	"encoding/json"

	"github.com/0chain/gosdk/zcnbridge/http"
	"github.com/pkg/errors"
)

type AuthorizerNodeResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// GetAuthorizers Returns authorizers
func GetAuthorizers() ([]*AuthorizerNodeResponse, error) {
	var (
		nodes []*AuthorizerNodeResponse
	)

	resp, err := http.MakeSCRestAPICall(http.GetAuthorizersPath, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error requesting authorizers")
	}
	if len(resp) == 0 {
		return nil, errors.New("empty response")
	}

	if err = json.Unmarshal(resp, &nodes); err != nil {
		return nil, errors.Wrap(err, "error decoding response:")
	}

	return nodes, nil
}
