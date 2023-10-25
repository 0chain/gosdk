package zboxapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	BaseURL          = "https://0box.dev.zus.network"
	AppType          = "vult"
	ClientID         = "175369aaf5bb11003fcbcc73bfa937ca2d1e255e58fe010b609670632617049d"
	ClientPublicKey  = "0dbac1d606943969b94f4af417de516366b1e9540da9189b9b1d1c3098cda6019c59dad0e2b3ed02b858b5f09aeea0bb1aee1735218d6b3977b0d893da970d80"
	ClientPrivateKey = "76ed3cb48bc76194ddcfd2c9785e855869b356b66e8dc67c968104d122469b13"
	PhoneNumber      = "+16026666666"
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

	token, err := c.CreateJwtToken(context.TODO(), PhoneNumber, sessionID, "000000") //any otp works on test phone number

	require.Nil(t, err)
	require.NotEmpty(t, token)

	refreshedToken, err := c.RefreshJwtToken(context.TODO(), PhoneNumber, token)
	require.Nil(t, err)
	require.NotEmpty(t, refreshedToken)
	require.NotEqual(t, token, refreshedToken)
}

func TestGetFreeStorage(t *testing.T) {
	t.Skip("Only for local debugging")

	c := NewClient()
	c.SetRequest(BaseURL, AppType)
	c.SetWallet(ClientID, ClientPrivateKey, ClientPublicKey)
	sessionID, err := c.CreateJwtSession(context.TODO(), PhoneNumber)

	require.Nil(t, err)
	require.GreaterOrEqual(t, sessionID, int64(0))

	token, err := c.CreateJwtToken(context.TODO(), PhoneNumber, sessionID, "000000") //any otp works on test phone number

	require.Nil(t, err)
	require.NotEmpty(t, token)

	marker, err := c.GetFreeStorage(context.TODO(), PhoneNumber, "jwt-"+token)

	require.NoError(t, err)
	require.NotEmpty(t, marker.Assigner)
	require.Equal(t, ClientID, marker.Recipient)
}

func TestGetSharedToMe(t *testing.T) {
	t.Skip("Only for local debugging")

	c := NewClient()
	c.SetRequest(BaseURL, AppType)
	c.SetWallet(ClientID, ClientPrivateKey, ClientPublicKey)
	sessionID, err := c.CreateJwtSession(context.TODO(), PhoneNumber)

	require.Nil(t, err)
	require.GreaterOrEqual(t, sessionID, int64(0))

	token, err := c.CreateJwtToken(context.TODO(), PhoneNumber, sessionID, "000000") //any otp works on test phone number

	require.Nil(t, err)
	require.NotEmpty(t, token)

	list, err := c.GetSharedToMe(context.TODO(), PhoneNumber, "jwt-"+token)

	require.NoError(t, err)
	require.Greater(t, len(list), 0)
	// require.NotEmpty(t, marker.Assigner)
	// require.Equal(t, ClientID, marker.Recipient)
}

func TestGetSharedByMe(t *testing.T) {
	t.Skip("Only for local debugging")

	c := NewClient()
	c.SetRequest(BaseURL, AppType)
	c.SetWallet(ClientID, ClientPrivateKey, ClientPublicKey)
	sessionID, err := c.CreateJwtSession(context.TODO(), PhoneNumber)

	require.Nil(t, err)
	require.GreaterOrEqual(t, sessionID, int64(0))

	token, err := c.CreateJwtToken(context.TODO(), PhoneNumber, sessionID, "000000") //any otp works on test phone number

	require.Nil(t, err)
	require.NotEmpty(t, token)

	list, err := c.GetSharedByMe(context.TODO(), PhoneNumber, "jwt-"+token)

	require.NoError(t, err)
	require.Greater(t, len(list), 0)
	// require.NotEmpty(t, marker.Assigner)
	// require.Equal(t, ClientID, marker.Recipient)
}
