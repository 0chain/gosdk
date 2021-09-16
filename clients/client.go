package clients

import (
	"net/url"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
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
