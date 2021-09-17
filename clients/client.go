package clients

import (
	"net/http"
	"net/url"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/zboxcore/client"
)

// Client a client instance of restful api
type Client struct {
	ClientID        string
	ClientPublicKey string
	BaseURL         string
}

// create client instance
func NewClient(clientID, clientPublicKey, baseURL string) (Client, error) {
	u, err := url.Parse(baseURL)

	c := Client{
		ClientID:        clientID,
		ClientPublicKey: clientPublicKey,
	}

	if err != nil {
		return c, errors.Throw(constants.ErrInvalidParameter, "baseURL")
	}

	c.BaseURL = u.String()

	return c, nil
}

func (c *Client) SignRequest(req *http.Request, allocation string) error {

	req.Header.Set("X-App-Client-ID", c.ClientID)
	req.Header.Set("X-App-Client-Key", c.ClientPublicKey)

	sign, err := client.Sign(encryption.Hash(allocation))
	if err != nil {
		return err
	}
	req.Header.Set("X-App-Client-Signature", sign)

	return nil
}
