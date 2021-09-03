package resty

import "net/http"

// Result result of a http request
type Result struct {
	Request  *http.Request
	Response *http.Response
	Err      error
}
