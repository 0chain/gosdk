package sdk

import (
	"encoding/base64"
	"encoding/json"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/marker"
)

type AuthTicket struct {
	b64Ticket string
}

func InitAuthTicket(authTicket string) *AuthTicket {
	at := &AuthTicket{}
	at.b64Ticket = authTicket
	return at
}

func (at *AuthTicket) IsDir() (bool, error) {
	sEnc, err := base64.StdEncoding.DecodeString(at.b64Ticket)
	if err != nil {
		return false, common.NewError("auth_ticket_decode_error", "Error decoding the auth ticket."+err.Error())
	}
	authTicket := &marker.AuthTicket{}
	err = json.Unmarshal(sEnc, at)
	if err != nil {
		return false, common.NewError("auth_ticket_decode_error", "Error unmarshaling the auth ticket."+err.Error())
	}
	return authTicket.RefType == fileref.DIRECTORY, nil
}

func (at *AuthTicket) GetFileName() (string, error) {
	sEnc, err := base64.StdEncoding.DecodeString(at.b64Ticket)
	if err != nil {
		return "", common.NewError("auth_ticket_decode_error", "Error decoding the auth ticket."+err.Error())
	}
	authTicket := &marker.AuthTicket{}
	err = json.Unmarshal(sEnc, at)
	if err != nil {
		return "", common.NewError("auth_ticket_decode_error", "Error unmarshaling the auth ticket."+err.Error())
	}
	return authTicket.FileName, nil
}
