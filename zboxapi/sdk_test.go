package zboxapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	BaseURL          = "https://0box.dev.zus.network"
	AppType          = "vult"
	ClientID         = "70e1318a9709786cf975f15ca941bee73d0f422305ecd78b0f358870ec17f41d"
	ClientPublicKey  = "4ec4b4dfb8c9ceb8fb6e84ef46e503c3445a0c6d770986a019cdbef4bc47b70dfadd5441f708f0df47df14e5cd6a0aa94ec31ca66e337692d9a92599d9456a81"
	ClientPrivateKey = "982801f352e886eaaf61196d83373b4cc09e9a598ffe1f49bf5adf905174cb0c"
	UserID           = "lWVZRhERosYtXR9MBJh5yJUtweI4"
	PhoneNumber      = "+917777777777"
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
	sessionID, err := c.CreateJwtSession(context.TODO(), UserID)

	require.Nil(t, err)
	require.GreaterOrEqual(t, sessionID, int64(0))

	token, err := c.CreateJwtToken(context.TODO(), UserID, sessionID) //any otp works on test phone number

	require.Nil(t, err)
	require.NotEmpty(t, token)

	refreshedToken, err := c.RefreshJwtToken(context.TODO(), UserID, token)
	require.Nil(t, err)
	require.NotEmpty(t, refreshedToken)
	require.NotEqual(t, token, refreshedToken)
}

func TestGetFreeStorage(t *testing.T) {
	t.Skip("Only for local debugging")

	c := NewClient()
	c.SetRequest(BaseURL, AppType)
	c.SetWallet(ClientID, ClientPrivateKey, ClientPublicKey)
	sessionID, err := c.CreateJwtSession(context.TODO(), UserID)

	require.Nil(t, err)
	require.GreaterOrEqual(t, sessionID, int64(0))

	token, err := c.CreateJwtToken(context.TODO(), UserID, sessionID) //any otp works on test phone number

	require.Nil(t, err)
	require.NotEmpty(t, token)

	marker, err := c.GetFreeStorage(context.TODO(), PhoneNumber, "jwt-"+token)

	require.NoError(t, err)
	require.NotEmpty(t, marker.Assigner)
	require.Equal(t, ClientID, marker.Recipient)
}

var testShareInfo = SharedInfo{
	AuthTicket:    "eyJjbGllbnRfaWQiOiIiLCJvd25lcl9pZCI6IjcwZTEzMThhOTcwOTc4NmNmOTc1ZjE1Y2E5NDFiZWU3M2QwZjQyMjMwNWVjZDc4YjBmMzU4ODcwZWMxN2Y0MWQiLCJhbGxvY2F0aW9uX2lkIjoiZjQ0OTUxZDkwODRiMTExZGMxNDliMmNkN2E5Nzg5YmU5MDVlYjFiMWRhNzdjMjYxNDZiMWNkY2IxNzE3NTI0NiIsImZpbGVfcGF0aF9oYXNoIjoiM2RlOWQ1ZTMzYWJlNWI3ZjhhNzM2OGY0ZmE4N2QwMmY1MjI1YzIzMzhmM2Q3YWI0MGQxNDczM2NiYmI4ZTc1YiIsImFjdHVhbF9maWxlX2hhc2giOiJkYTJjMzIxZmFiN2RkNmYyZDVlZTAzZWQwNDk2OGJlMTA0YjdjNmY2MTYyYTVmY2ZjNDFmZTEyZTY3ZDBkNjUzIiwiZmlsZV9uYW1lIjoiUVHlm77niYcyMDIxMDkzMDEyMDM1Ny5qcGciLCJyZWZlcmVuY2VfdHlwZSI6ImYiLCJleHBpcmF0aW9uIjowLCJ0aW1lc3RhbXAiOjE2OTgyODY5MjcsImVuY3J5cHRlZCI6ZmFsc2UsInNpZ25hdHVyZSI6ImRkMzg4NzI2YTcwYzBjN2Y5NDZkMTQwMTRjMjhhZTg1MjM4ZTliNmJkMmExMzRjMWUxOGE3MTE5NDViYzg4MGYifQ==",
	Message:       "shared by unit test",
	ShareInfoType: "public",
	Link:          "https://0box.page.link/cnfFExcvKKRaFzyE9",
}

func TestCreateSharedInfo(t *testing.T) {
	t.Skip("Only for local debugging")

	c := NewClient()
	c.SetRequest(BaseURL, AppType)
	c.SetWallet(ClientID, ClientPrivateKey, ClientPublicKey)
	sessionID, err := c.CreateJwtSession(context.TODO(), UserID)

	require.Nil(t, err)
	require.GreaterOrEqual(t, sessionID, int64(0))

	token, err := c.CreateJwtToken(context.TODO(), UserID, sessionID) //any otp works on test phone number

	require.Nil(t, err)
	require.NotEmpty(t, token)

	err = c.CreateSharedInfo(context.TODO(), PhoneNumber, "jwt-"+token, testShareInfo)

	require.NoError(t, err)

}

func TestGetSharedToMe(t *testing.T) {
	t.Skip("Only for local debugging")

	c := NewClient()
	c.SetRequest(BaseURL, AppType)
	c.SetWallet(ClientID, ClientPrivateKey, ClientPublicKey)
	sessionID, err := c.CreateJwtSession(context.TODO(), UserID)

	require.Nil(t, err)
	require.GreaterOrEqual(t, sessionID, int64(0))

	token, err := c.CreateJwtToken(context.TODO(), UserID, sessionID) //any otp works on test phone number

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
	sessionID, err := c.CreateJwtSession(context.TODO(), UserID)

	require.Nil(t, err)
	require.GreaterOrEqual(t, sessionID, int64(0))

	token, err := c.CreateJwtToken(context.TODO(), UserID, sessionID) //any otp works on test phone number

	require.Nil(t, err)
	require.NotEmpty(t, token)

	list, err := c.GetSharedByMe(context.TODO(), PhoneNumber, "jwt-"+token)

	require.NoError(t, err)
	require.Greater(t, len(list), 0)

}

func TestGetSharedByPublic(t *testing.T) {
	t.Skip("Only for local debugging")

	c := NewClient()
	c.SetRequest(BaseURL, AppType)
	c.SetWallet(ClientID, ClientPrivateKey, ClientPublicKey)
	sessionID, err := c.CreateJwtSession(context.TODO(), UserID)

	require.Nil(t, err)
	require.GreaterOrEqual(t, sessionID, int64(0))

	token, err := c.CreateJwtToken(context.TODO(), UserID, sessionID) //any otp works on test phone number

	require.Nil(t, err)
	require.NotEmpty(t, token)

	list, err := c.GetSharedByPublic(context.TODO(), PhoneNumber, "jwt-"+token)

	require.NoError(t, err)
	require.Greater(t, len(list), 0)

}

func TestDeleteSharedInfo(t *testing.T) {
	t.Skip("Only for local debugging")

	c := NewClient()
	c.SetRequest(BaseURL, AppType)
	c.SetWallet(ClientID, ClientPrivateKey, ClientPublicKey)
	sessionID, err := c.CreateJwtSession(context.TODO(), UserID)

	require.Nil(t, err)
	require.GreaterOrEqual(t, sessionID, int64(0))

	token, err := c.CreateJwtToken(context.TODO(), UserID, sessionID) //any otp works on test phone number

	require.Nil(t, err)
	require.NotEmpty(t, token)

	buf, _ := base64.StdEncoding.DecodeString(testShareInfo.AuthTicket)

	items := make(map[string]any)

	json.Unmarshal(buf, &items)

	lookupHash := items["file_path_hash"].(string)

	err = c.DeleteSharedInfo(context.TODO(), PhoneNumber, "jwt-"+token, testShareInfo.AuthTicket, lookupHash)

	require.NoError(t, err)

}
