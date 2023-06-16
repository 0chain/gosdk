package zboxapi

import (
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
