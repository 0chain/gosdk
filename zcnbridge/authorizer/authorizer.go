package authorizer

import (
	"encoding/json"

	"github.com/0chain/gosdk/zcnbridge/http"
	"github.com/pkg/errors"
)

type Authorizer struct {
}

type Nodes struct {
	NodeMap map[string]*Node `json:"node_map"`
}

type Node struct {
	ID        string `json:"id"`
	PublicKey string `json:"public_key"`
	URL       string `json:"url"`
}

// GetAuthorizers Returns authorizers
func GetAuthorizers() (*Nodes, error) {
	resp, err := http.MakeSCRestAPICall(http.GetAuthorizersPath, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error requesting authorizers")
	}
	if len(resp) == 0 {
		return nil, errors.New("empty response")
	}

	an := &Nodes{}

	if err = json.Unmarshal(resp, &an); err != nil {
		return nil, errors.Wrap(err, "error decoding response:")
	}

	return an, nil
}
