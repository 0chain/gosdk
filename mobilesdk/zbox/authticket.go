package zbox

import (
	"github.com/0chain/gosdk/zboxcore/sdk"
)

// AuthTicket - auth ticket structure
type AuthTicket struct {
	sdkAuthTicket *sdk.AuthTicket
}

// InitAuthTicket - init auth ticket from ID
func InitAuthTicket(authTicket string) *AuthTicket {
	at := &AuthTicket{}
	at.sdkAuthTicket = sdk.InitAuthTicket(authTicket)
	return at
}

// IsDir - checking if it's dir
func (at *AuthTicket) IsDir() (bool, error) {
	return at.sdkAuthTicket.IsDir()
}

// GetFilename - getting file name
func (at *AuthTicket) GetFilename() (string, error) {
	return at.sdkAuthTicket.GetFileName()
}
