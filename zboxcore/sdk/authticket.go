package sdk

import (
	"encoding/base64"
	"encoding/json"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/marker"
)

type AuthTicket struct {
	b64Ticket string
}

// InitAuthTicket initialize auth ticket instance
//   - authTicket: base64 encoded auth ticket
func InitAuthTicket(authTicket string) *AuthTicket {
	at := &AuthTicket{}
	at.b64Ticket = authTicket
	return at
}

func (at *AuthTicket) IsDir() (bool, error) {
	sEnc, err := base64.StdEncoding.DecodeString(at.b64Ticket)
	if err != nil {
		return false, errors.New("auth_ticket_decode_error", "Error decoding the auth ticket."+err.Error())
	}
	authTicket := &marker.AuthTicket{}
	err = json.Unmarshal(sEnc, authTicket)
	if err != nil {
		return false, errors.New("auth_ticket_decode_error", "Error unmarshaling the auth ticket."+err.Error())
	}
	return authTicket.RefType == fileref.DIRECTORY, nil
}

func (at *AuthTicket) GetFileName() (string, error) {
	sEnc, err := base64.StdEncoding.DecodeString(at.b64Ticket)
	if err != nil {
		return "", errors.New("auth_ticket_decode_error", "Error decoding the auth ticket."+err.Error())
	}
	authTicket := &marker.AuthTicket{}
	err = json.Unmarshal(sEnc, authTicket)
	if err != nil {
		return "", errors.New("auth_ticket_decode_error", "Error unmarshaling the auth ticket."+err.Error())
	}
	return authTicket.FileName, nil
}

func (at *AuthTicket) GetLookupHash() (string, error) {
	sEnc, err := base64.StdEncoding.DecodeString(at.b64Ticket)
	if err != nil {
		return "", errors.New("auth_ticket_decode_error", "Error decoding the auth ticket."+err.Error())
	}
	authTicket := &marker.AuthTicket{}
	err = json.Unmarshal(sEnc, authTicket)
	if err != nil {
		return "", errors.New("auth_ticket_decode_error", "Error unmarshaling the auth ticket."+err.Error())
	}
	return authTicket.FilePathHash, nil
}

func (at *AuthTicket) Unmarshall() (*marker.AuthTicket, error) {
	sEnc, err := base64.StdEncoding.DecodeString(at.b64Ticket)
	if err != nil {
		return nil, errors.New("auth_ticket_decode_error", "Error decoding the auth ticket."+err.Error())
	}
	authTicket := &marker.AuthTicket{}
	err = json.Unmarshal(sEnc, authTicket)
	if err != nil {
		return nil, errors.New("auth_ticket_decode_error", "Error unmarshaling the auth ticket."+err.Error())
	}
	return authTicket, nil
}
