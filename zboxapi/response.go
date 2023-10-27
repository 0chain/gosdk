package zboxapi

import (
	"encoding/base64"
	"encoding/json"
	"errors"
)

var (
	ErrInvalidJsonResponse = errors.New("zbox-srv: invalid json response")
)

type ErrorResponse struct {
	Error json.RawMessage `json:"error"`
}

type CsrfTokenResponse struct {
	Token string `json:"csrf_token"`
}

type JwtTokenResponse struct {
	Token string `json:"jwt_token"`
}

type FreeStorageResponse struct {
	Data      string `json:"marker"`
	FundingID int    `json:"funding_id"`
}

type MarkerData struct {
	Marker             string `json:"marker"`
	RecipientPublicKey string `json:"recipient_public_key"`
}

func (fs *FreeStorageResponse) ToMarker() (*FreeMarker, error) {

	buf, err := base64.StdEncoding.DecodeString(fs.Data)

	if err != nil {
		return nil, err
	}

	data := &MarkerData{}
	err = json.Unmarshal(buf, &data)
	if err != nil {
		return nil, err
	}

	buf, err = base64.StdEncoding.DecodeString(data.Marker)
	if err != nil {
		return nil, err
	}

	fm := &FreeMarker{}

	err = json.Unmarshal(buf, fm)
	if err != nil {
		return nil, err
	}

	return fm, nil
}

type FreeMarker struct {
	Assigner   string  `json:"assigner"`
	Recipient  string  `json:"recipient"`
	FreeTokens float64 `json:"free_tokens"`
	Nonce      int64   `json:"nonce"`
	Signature  string  `json:"signature"`
}

type JsonResult[T any] struct {
	Message string `json:"message"`
	Error   string `json:"error"`
	Data    []T    `json:"data"`
}

type SharedInfo struct {
	AuthTicket    string `json:"auth_ticket"`
	Message       string `json:"message"`
	ShareInfoType string `json:"share_info_type"`
	Link          string `json:"link"`
}

type SharedInfoSent struct {
	AuthTicket    string `json:"auth_ticket"`
	Message       string `json:"message"`
	ShareInfoType string `json:"share_info_type"`
	Receiver      string `json:"receiver_client_id"`
	Link          string `json:"link"`
	ReceiverName  string `json:"receiver_name"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type SharedInfoReceived struct {
	AuthTicket    string `json:"auth_ticket"`
	Message       string `json:"message"`
	ShareInfoType string `json:"share_info_type"`
	ClientID      string `json:"client_id"`
	Receiver      string `json:"receiver_client_id"`
	LookupHash    string `json:"lookup_hash"`
	SenderName    string `json:"sender_name"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}
