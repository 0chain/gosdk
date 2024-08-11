package sdks

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/resty"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/client"
)

// ZBox  sdk client instance
type ZBox struct {
	// ClientID client id
	ClientID string
	// ClientKey client key
	ClientKey string
	// SignatureScheme signature scheme
	SignatureScheme string

	// Wallet wallet
	Wallet *zcncrypto.Wallet

	// NewRequest create http request
	NewRequest func(method, url string, body io.Reader) (*http.Request, error)
}

// New create an sdk client instance given its configuration
//   - clientID client id of the using client
//   - clientKey client key of the using client
//   - signatureScheme signature scheme for transaction encryption
//   - wallet wallet of the using client
func New(clientID, clientKey, signatureScheme string, wallet *zcncrypto.Wallet) *ZBox {
	s := &ZBox{
		ClientID:        clientID,
		ClientKey:       clientKey,
		SignatureScheme: signatureScheme,
		Wallet:          wallet,
		NewRequest:      http.NewRequest,
	}

	return s
}

// InitWallet init wallet from json
//   - js json string of wallet
func (z *ZBox) InitWallet(js string) error {
	return json.Unmarshal([]byte(js), &z.Wallet)
}

// SignRequest sign request with client_id, client_key and sign by adding headers to the request
//   - req http request
//   - allocationID allocation id
func (z *ZBox) SignRequest(req *http.Request, allocationID string) error {

	if req == nil {
		return errors.Throw(constants.ErrInvalidParameter, "req")
	}

	req.Header.Set("X-App-Client-ID", z.ClientID)
	req.Header.Set("X-App-Client-Key", z.ClientKey)

	hash := encryption.Hash(allocationID)

	sign, err := sys.Sign(hash, z.SignatureScheme, client.GetClientSysKeys())
	if err != nil {
		return err
	}

	// ClientSignatureHeader represents http request header contains signature.
	req.Header.Set("X-App-Client-Signature", sign)

	return nil
}

// CreateTransport create http.Transport with default dial timeout
func (z *ZBox) CreateTransport() *http.Transport {
	return &http.Transport{
		Dial: (&net.Dialer{
			Timeout: resty.DefaultDialTimeout,
		}).Dial,
		TLSHandshakeTimeout: resty.DefaultDialTimeout,
	}
}

// BuildUrls build full request url given base urls, query string, path format and path args
//   - baseURLs base urls
//   - queryString query string
//   - pathFormat path format
//   - pathArgs path args
func (z *ZBox) BuildUrls(baseURLs []string, queryString map[string]string, pathFormat string, pathArgs ...interface{}) []string {

	requestURL := pathFormat
	if len(pathArgs) > 0 {
		requestURL = fmt.Sprintf(pathFormat, pathArgs...)
	}

	if len(queryString) > 0 {
		requestQuery := make(url.Values)
		for k, v := range queryString {
			requestQuery.Add(k, v)
		}

		requestURL += "?" + requestQuery.Encode()
	}

	list := make([]string, len(baseURLs))
	for k, v := range baseURLs {
		list[k] = v + requestURL
	}

	return list
}

// DoPost do post request with request and handle
//   - req request instance
//   - handle handle function for the response
func (z *ZBox) DoPost(req *Request, handle resty.Handle) *resty.Resty {

	opts := make([]resty.Option, 0, 5)

	opts = append(opts, resty.WithRetry(resty.DefaultRetry))
	opts = append(opts, resty.WithRequestInterceptor(func(r *http.Request) error {
		return z.SignRequest(r, req.AllocationID) //nolint
	}))

	if len(req.ContentType) > 0 {
		opts = append(opts, resty.WithHeader(map[string]string{
			"Content-Type": req.ContentType,
		}))
	}

	opts = append(opts, resty.WithTransport(z.CreateTransport()))

	r := resty.New(opts...).Then(handle)

	return r
}
