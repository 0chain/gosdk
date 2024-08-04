package zboxapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/0chain/gosdk/zcncore"
	"net/http"
	"strconv"
	"time"

	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/core/resty"
)

var log logger.Logger

func GetLogger() *logger.Logger {
	return &log
}

type Client struct {
	baseUrl          string
	appType          string
	clientID         string
	clientPublicKey  string
	clientPrivateKey string
}

// NewClient create a zbox api client with wallet info
func NewClient() *Client {
	return &Client{}
}

// SetRequest set base url and app type of zbox api request
func (c *Client) SetRequest(baseUrl, appType string) {
	c.baseUrl = baseUrl
	c.appType = appType
}

func (c *Client) SetWallet(clientID, clientPrivateKey, clientPublicKey string) {
	c.clientID = clientID
	c.clientPrivateKey = clientPrivateKey
	c.clientPublicKey = clientPublicKey
}

func (c *Client) parseResponse(resp *http.Response, respBody []byte, result interface{}) error {

	log.Info("zboxapi: ", resp.StatusCode, " ", string(respBody))
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		if err := json.Unmarshal(respBody, result); err != nil {
			return thrown.Throw(ErrInvalidJsonResponse, string(respBody))
		}
		return nil
	}

	if len(respBody) == 0 {
		return errors.New(resp.Status)
	}

	errResp := &ErrorResponse{}
	if err := json.Unmarshal(respBody, errResp); err != nil {
		return thrown.Throw(ErrInvalidJsonResponse, string(respBody))
	}

	return errors.New(string(errResp.Error))
}

// GetCsrfToken obtain a fresh csrf token from 0box api server
func (c *Client) GetCsrfToken(ctx context.Context) (string, error) {
	r, err := c.createResty(ctx, "", "", nil)
	if err != nil {
		return "", err
	}
	result := &CsrfTokenResponse{}
	r.DoGet(ctx, c.baseUrl+"/v2/csrftoken").Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
		if err != nil {
			return err
		}
		return c.parseResponse(resp, respBody, result)
	})

	if err := r.Wait(); len(err) > 0 {
		return "", err[0]
	}

	return result.Token, nil
}

func (c *Client) createResty(ctx context.Context, csrfToken, userID string, headers map[string]string) (*resty.Resty, error) {
	h := make(map[string]string)
	h["X-App-Client-ID"] = c.clientID
	h["X-App-Client-Key"] = c.clientPublicKey
	h["X-App-User-ID"] = userID

	if c.clientPrivateKey != "" {
		data := fmt.Sprintf("%v:%v:%v", c.clientID, userID, c.clientPublicKey)
		hash := encryption.Hash(data)

		sign, err := zcncore.SignFn(hash)
		if err != nil {
			return nil, err
		}
		h["X-App-Client-Signature"] = sign
	}

	h["X-CSRF-TOKEN"] = csrfToken
	h["X-App-Timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)
	h["X-App-ID-Token"] = "*" //ignore firebase token in jwt requests
	h["X-App-Type"] = c.appType

	for k, v := range headers {
		h[k] = v
	}

	return resty.New(resty.WithHeader(h)), nil
}

// CreateJwtSession create a jwt session with user id
func (c *Client) CreateJwtSession(ctx context.Context, userID string) (int64, error) {

	csrfToken, err := c.GetCsrfToken(ctx)
	if err != nil {
		return 0, err
	}

	r, err := c.createResty(ctx, csrfToken, userID, nil)

	if err != nil {
		return 0, err
	}

	var sessionID int64

	r.DoPost(ctx, nil, c.baseUrl+"/v2/jwt/session").
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			if err != nil {
				return err
			}

			return c.parseResponse(resp, respBody, &sessionID)
		})

	if errs := r.Wait(); len(errs) > 0 {
		return 0, errs[0]
	}

	return sessionID, nil
}

// CreateJwtToken create a jwt token with jwt session id and otp
func (c *Client) CreateJwtToken(ctx context.Context, userID string, jwtSessionID int64) (string, error) {
	csrfToken, err := c.GetCsrfToken(ctx)
	if err != nil {
		return "", err
	}
	headers := map[string]string{
		"X-JWT-Session-ID": strconv.FormatInt(jwtSessionID, 10),
	}

	r, err := c.createResty(ctx, csrfToken, userID, headers)

	if err != nil {
		return "", err
	}

	result := &JwtTokenResponse{}
	r.DoPost(ctx, nil, c.baseUrl+"/v2/jwt/token").
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			if err != nil {
				return err
			}

			return c.parseResponse(resp, respBody, result)
		})

	if errs := r.Wait(); len(errs) > 0 {
		return "", errs[0]
	}

	return result.Token, nil
}

// RefreshJwtToken refresh jwt token
func (c *Client) RefreshJwtToken(ctx context.Context, userID string, token string) (string, error) {
	csrfToken, err := c.GetCsrfToken(ctx)
	if err != nil {
		return "", err
	}
	headers := map[string]string{
		"X-JWT-Token": token,
	}

	r, err := c.createResty(ctx, csrfToken, userID, headers)

	if err != nil {
		return "", err
	}

	result := &JwtTokenResponse{}
	r.DoPut(ctx, nil, c.baseUrl+"/v2/jwt/token").
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			if err != nil {
				return err
			}

			return c.parseResponse(resp, respBody, result)
		})

	if errs := r.Wait(); len(errs) > 0 {
		return "", errs[0]
	}

	return result.Token, nil
}

func (c *Client) GetFreeStorage(ctx context.Context, phoneNumber, token string) (*FreeMarker, error) {
	csrfToken, err := c.GetCsrfToken(ctx)
	if err != nil {
		return nil, err
	}
	headers := map[string]string{
		"X-App-ID-Token": token,
	}

	r, err := c.createResty(ctx, csrfToken, phoneNumber, headers)

	if err != nil {
		return nil, err
	}

	result := &FreeStorageResponse{}
	r.DoGet(ctx, c.baseUrl+"/v2/freestorage").
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			if err != nil {
				return err
			}

			return c.parseResponse(resp, respBody, result)
		})

	if errs := r.Wait(); len(errs) > 0 {
		return nil, errs[0]
	}

	return result.ToMarker()

}

func (c *Client) CreateSharedInfo(ctx context.Context, phoneNumber, token string, s SharedInfo) error {
	csrfToken, err := c.GetCsrfToken(ctx)
	if err != nil {
		return err
	}
	headers := map[string]string{
		"X-App-ID-Token": token,
		"Content-Type":   "application/json",
	}

	r, err := c.createResty(ctx, csrfToken, phoneNumber, headers)

	if err != nil {
		return err
	}

	buf, err := json.Marshal(s)
	if err != nil {
		return err
	}

	result := &JsonResult[string]{}
	r.DoPost(ctx, bytes.NewReader(buf), c.baseUrl+"/v2/shareinfo").
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			if err != nil {
				return err
			}

			return c.parseResponse(resp, respBody, &result)
		})

	if errs := r.Wait(); len(errs) > 0 {
		return errs[0]
	}

	return nil
}

func (c *Client) DeleteSharedInfo(ctx context.Context, phoneNumber, token, authTicket string, lookupHash string) error {
	csrfToken, err := c.GetCsrfToken(ctx)
	if err != nil {
		return err
	}
	headers := map[string]string{
		"X-App-ID-Token": token,
	}

	r, err := c.createResty(ctx, csrfToken, phoneNumber, headers)

	if err != nil {
		return err
	}

	result := &JsonResult[string]{}
	r.DoDelete(ctx, c.baseUrl+"/v2/shareinfo?auth_ticket="+authTicket+"&lookup_hash="+lookupHash).
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			if err != nil {
				return err
			}

			return c.parseResponse(resp, respBody, &result)
		})

	if errs := r.Wait(); len(errs) > 0 {
		return errs[0]
	}

	return nil
}

func (c *Client) GetSharedByPublic(ctx context.Context, phoneNumber, token string) ([]SharedInfoSent, error) {
	return c.getShared(ctx, phoneNumber, token, false)
}

func (c *Client) GetSharedByMe(ctx context.Context, phoneNumber, token string) ([]SharedInfoSent, error) {
	return c.getShared(ctx, phoneNumber, token, true)
}

func (c *Client) getShared(ctx context.Context, phoneNumber, token string, isPrivate bool) ([]SharedInfoSent, error) {
	csrfToken, err := c.GetCsrfToken(ctx)
	if err != nil {
		return nil, err
	}
	headers := map[string]string{
		"X-App-ID-Token": token,
	}

	r, err := c.createResty(ctx, csrfToken, phoneNumber, headers)

	if err != nil {
		return nil, err
	}

	shareInfoType := "public"
	if isPrivate {
		shareInfoType = "private"
	}

	result := &JsonResult[SharedInfoSent]{}
	r.DoGet(ctx, c.baseUrl+"/v2/shareinfo/shared?share_info_type="+shareInfoType).
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			if err != nil {
				return err
			}

			return c.parseResponse(resp, respBody, &result)
		})

	if errs := r.Wait(); len(errs) > 0 {
		return nil, errs[0]
	}

	return result.Data, nil

}

func (c *Client) GetSharedToMe(ctx context.Context, phoneNumber, token string) ([]SharedInfoReceived, error) {
	csrfToken, err := c.GetCsrfToken(ctx)
	if err != nil {
		return nil, err
	}
	headers := map[string]string{
		"X-App-ID-Token": token,
	}

	r, err := c.createResty(ctx, csrfToken, phoneNumber, headers)

	if err != nil {
		return nil, err
	}

	result := &JsonResult[SharedInfoReceived]{}
	r.DoGet(ctx, c.baseUrl+"/v2/shareinfo/received?share_info_type=private").
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			if err != nil {
				return err
			}

			return c.parseResponse(resp, respBody, &result)
		})

	if errs := r.Wait(); len(errs) > 0 {
		return nil, errs[0]
	}

	return result.Data, nil

}
