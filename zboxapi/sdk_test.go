package zboxapi

import (
	"context"
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
