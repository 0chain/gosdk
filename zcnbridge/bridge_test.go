package zcnbridge

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/0chain/gosdk/core/common"
	"github.com/stretchr/testify/require"
)

type AuthorizerConfigTarget struct {
	Fee common.Balance `json:"fee"`
}

type AuthorizerNodeTarget struct {
	ID        string                  `json:"id"`
	PublicKey string                  `json:"public_key"`
	URL       string                  `json:"url"`
	Config    *AuthorizerConfigTarget `json:"config"`
}

type AuthorizerConfigSource struct {
	Fee string `json:"fee"`
}

type AuthorizerNodeSource struct {
	ID     string                  `json:"id"`
	Config *AuthorizerConfigSource `json:"config"`
}

func (an *AuthorizerNodeTarget) Decode(input []byte) error {
	var objMap map[string]*json.RawMessage
	err := json.Unmarshal(input, &objMap)
	if err != nil {
		return err
	}

	id, ok := objMap["id"]
	if ok {
		var idStr *string
		err = json.Unmarshal(*id, &idStr)
		if err != nil {
			return err
		}
		an.ID = *idStr
	}

	pk, ok := objMap["public_key"]
	if ok {
		var pkStr *string
		err = json.Unmarshal(*pk, &pkStr)
		if err != nil {
			return err
		}
		an.PublicKey = *pkStr
	}

	url, ok := objMap["url"]
	if ok {
		var urlStr *string
		err = json.Unmarshal(*url, &urlStr)
		if err != nil {
			return err
		}
		an.URL = *urlStr
	}

	rawCfg, ok := objMap["config"]
	if ok {
		var cfg = &AuthorizerConfigTarget{}
		err = cfg.Decode(*rawCfg)
		if err != nil {
			return err
		}

		an.Config = cfg
	}

	return nil
}

func (c *AuthorizerConfigTarget) Decode(input []byte) (err error) {
	const (
		Fee = "fee"
	)

	var objMap map[string]*json.RawMessage
	err = json.Unmarshal(input, &objMap)
	if err != nil {
		return err
	}

	fee, ok := objMap[Fee]
	if ok {
		var feeStr *string
		err = json.Unmarshal(*fee, &feeStr)
		if err != nil {
			return err
		}

		var balance, err = strconv.ParseInt(*feeStr, 10, 64)
		if err != nil {
			return err
		}

		c.Fee = common.Balance(balance)
	}

	return nil
}

func Test_UpdateAuthorizerConfigTest(t *testing.T) {
	source := &AuthorizerNodeSource{
		ID: "12345678",
		Config: &AuthorizerConfigSource{
			Fee: "999",
		},
	}
	target := &AuthorizerNodeTarget{}

	bytes, err := json.Marshal(source)
	require.NoError(t, err)

	err = target.Decode(bytes)
	require.NoError(t, err)

	require.Equal(t, "", target.URL)
	require.Equal(t, "", target.PublicKey)
	require.Equal(t, "12345678", target.ID)
	require.Equal(t, common.Balance(999), target.Config.Fee)
}
