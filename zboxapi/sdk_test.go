package zboxapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	BaseURL          = "https://0box.demo.zus.network"
	AppType          = "vult"
	ClientID         = "8f6ce6457fc04cfb4eb67b5ce3162fe2b85f66ef81db9d1a9eaa4ffe1d2359e0"
	ClientPublicKey  = "c8c88854822a1039c5a74bdb8c025081a64b17f52edd463fbecb9d4a42d15608f93b5434e926d67a828b88e63293b6aedbaf0042c7020d0a96d2e2f17d3779a4"
	ClientPrivateKey = "72f480d4b1e7fb76e04327b7c2348a99a64f0ff2c5ebc3334a002aa2e66e8506"
	PhoneNumber      = "+919876543210"
)

func TestGetCsrfToken(t *testing.T) {
	t.Skip("Only for local debugging")

	c := NewClient()
	c.SetRequest(BaseURL, AppType)
	c.SetWallet(ClientID, ClientPrivateKey, ClientPublicKey)

	token, err := c.GetCsrfToken(context.TODO())

	require.Nil(t, err)
	require.True(t, len(token) > 0)
}

func TestJwtToken(t *testing.T) {
	t.Skip("Only for local debugging")

	c := NewClient()
	c.SetRequest(BaseURL, AppType)
	c.SetWallet(ClientID, ClientPrivateKey, ClientPublicKey)
	sessionID, err := c.CreateJwtSession(context.TODO(), PhoneNumber)

	require.Nil(t, err)
	require.GreaterOrEqual(t, sessionID, int64(0))

	token, err := c.CreateJwtToken(context.TODO(), PhoneNumber, sessionID, "123456") //any otp works on test phone number

	require.Nil(t, err)
	require.NotEmpty(t, token)

	refreshedToken, err := c.RefreshJwtToken(context.TODO(), PhoneNumber, token)
	require.Nil(t, err)
	require.NotEmpty(t, refreshedToken)
	require.NotEqual(t, token, refreshedToken)

}
