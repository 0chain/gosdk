package resty

import "net/http"

type Result struct {
	Request  *http.Request
	Response *http.Response
	Err      error
}
