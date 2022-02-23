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
	"github.com/0chain/gosdk/core/zcncrypto"
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
	Wallet zcncrypto.Wallet

	// NewRequest create http request
	NewRequest func(method, url string, body io.Reader) (*http.Request, error)
}

// New create a sdk client instance
func New(clientID, clientKey, signatureScheme string) *ZBox {
	s := &ZBox{
		ClientID:        clientID,
		ClientKey:       clientKey,
		SignatureScheme: signatureScheme,
		NewRequest:      http.NewRequest,
	}

	return s
}

// InitWallet init wallet from json
func (z *ZBox) InitWallet(js string) error {
	return json.Unmarshal([]byte(js), &z.Wallet)
}

// SignRequest sign request with client_id, client_key and sign
func (z *ZBox) SignRequest(req *http.Request, allocationID string) error {

	if req == nil {
		return errors.Throw(constants.ErrInvalidParameter, "req")
	}

	req.Header.Set("X-App-Client-ID", z.ClientID)
	req.Header.Set("X-App-Client-Key", z.ClientKey)

	hash := encryption.Hash(allocationID)

	var err error
	sign := ""
	for _, kv := range z.Wallet.Keys {
		ss := zcncrypto.NewSignatureScheme(z.SignatureScheme)
		err = ss.SetPrivateKey(kv.PrivateKey)
		if err != nil {
			return err
		}

		if len(sign) == 0 {
			sign, err = ss.Sign(hash)
		} else {
			sign, err = ss.Add(sign, hash)
		}
		if err != nil {
			return err
		}
	}

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

// BuildUrls build full request url
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

func (z *ZBox) DoPost(req *Request, handle resty.Handle) *resty.Resty {

	opts := make([]resty.Option, 0)

	opts = append(opts, resty.WithRetry(resty.DefaultRetry))
	opts = append(opts, resty.WithTimeout(resty.DefaultRequestTimeout))
	opts = append(opts, resty.WithBefore(func(r *http.Request) {
		z.SignRequest(r, req.AllocationID) //nolint
	}))

	if len(req.ContentType) > 0 {
		opts = append(opts, resty.WithHeader(map[string]string{
			"Content-Type": req.ContentType,
		}))
	}

	r := resty.New(z.CreateTransport(), handle, opts...)

	return r
}
