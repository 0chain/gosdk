// package blobber  wrap blobber's apis as sdk client
package blobber

import (
	"github.com/0chain/gosdk/clients"
)

// Client blobber client instance
type Client struct {
	clients.Client
}

// create blobber client instance
func NewClient(clientID string, clientPublicKey string, baseURL string) (Client, error) {
	c, err := clients.NewClient(clientID, clientPublicKey, baseURL)

	client := Client{}

	if err != nil {
		return client, err
	}

	client.Client = c

	return client, nil
}
