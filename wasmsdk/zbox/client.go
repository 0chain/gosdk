//go:build js && wasm
// +build js,wasm

package zbox

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/resty"
	"github.com/0chain/gosdk/core/zcncrypto"
)

// 0box api client

type Client struct {
	Host      string
	ClientID  string
	PublicKey string
}

func NewClient(host, clientID, clientPublicKey string) *Client {
	return &Client{
		Host:      host,
		ClientID:  clientID,
		PublicKey: clientPublicKey,
	}
}

// getCSRFToken
func (c *Client) getCSRFToken() (string, error) {
	var respErr error
	var token string
	r := resty.New(resty.WithHeader(c.SignHeader())).
		DoGet(context.TODO(), strings.TrimRight(c.Host, "/")+"/csrftoken").
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			if err != nil {
				respErr = err
				return err
			}

			if resp.StatusCode != http.StatusOK {
				respErr = errors.New("0box: " + resp.Status)
				return respErr
			}

			token = string(respBody)

			return nil
		})

	r.Wait()

	return token, respErr

}

func (c *Client) SignHeader() map[string]string {
	authHeaders := make(map[string]string)
	authHeaders["X-App-Client-ID"] = c.ClientID
	authHeaders["X-App-Client-Key"] = c.PublicKey
	now := strconv.FormatInt(time.Now().Unix(), 10)
	authHeaders["X-App-Timestamp"] = now
	sign, _ := zcncrypto.Sign(encryption.Hash(fmt.Sprintf("%v:%v", c.ClientID, now)))

	authHeaders["X-App-Client-Signature"] = sign

	return authHeaders
}
