package resty

import "time"

var (
	// DefaultDialTimeout default timeout of a dialer
	DefaultDialTimeout = 5 * time.Second
	// DefaultRequestTimeout default time out of a http request
	DefaultRequestTimeout = 10 * time.Second
	// DefaultRetry retry times if a request is failed with 5xx status code
	DefaultRetry = 3
)
