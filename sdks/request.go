package sdks

import (
	"io"
)

// Request request payload
type Request struct {

	//AllocationID optional. allocation id
	AllocationID string
	//ConnectionID optional. session id
	ConnectionID string

	// ContentType content-type in header
	ContentType string
	// Body form data
	Body io.Reader
	// QueryString query string
	QueryString map[string]string
}
