package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/resty"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/sdks"
	"github.com/0chain/gosdk/sdks/blobber"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"
)

// WMLockStatus
type WMLockStatus int

const (
	WMLockStatusFailed WMLockStatus = iota
	WMLockStatusPending
	WMLockStatusOK
)

type WMLockResult struct {
	Status    WMLockStatus `json:"status,omitempty"`
	CreatedAt int64        `json:"created_at,omitempty"`
}

// WriteMarkerMutex blobber WriteMarkerMutex client
type WriteMarkerMutex struct {
	mutex          sync.Mutex
	zbox           *sdks.ZBox
	allocationObj  *Allocation
	lockedBlobbers map[string]bool
}

// CreateWriteMarkerMutex create WriteMarkerMutex for allocation
func CreateWriteMarkerMutex(client *client.Client, allocationObj *Allocation) (*WriteMarkerMutex, error) {

	if client == nil {
		return nil, errors.Throw(constants.ErrInvalidParameter, "client")
	}

	if allocationObj == nil {
		return nil, errors.Throw(constants.ErrInvalidParameter, "allocationObj")
	}

	return &WriteMarkerMutex{
		allocationObj: allocationObj,
		zbox:          sdks.New(client.ClientID, client.ClientKey, client.SignatureScheme, client.Wallet),
	}, nil
}

// Lock acquire WriteMarker lock from blobbers
func (m *WriteMarkerMutex) Lock(ctx context.Context, connectionID string) error {
	if m == nil {
		return errors.Throw(constants.ErrInvalidParameter, "WriteMarkerMutex")
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.allocationObj == nil {
		return errors.Throw(constants.ErrInvalidParameter, "allocationObj")
	}

	T := len(m.allocationObj.Blobbers)

	if T == 0 {
		return errors.Throw(constants.ErrInvalidParameter, "blobbers")
	}

	urls := make([]string, T)

	builder := &strings.Builder{}
	for i, b := range m.allocationObj.Blobbers {
		builder.Reset()
		builder.WriteString(strings.TrimRight(b.Baseurl, "/")) //nolint: errcheck
		builder.WriteString(blobber.EndpointWriteMarkerLock)
		builder.WriteString(m.allocationObj.Tx)
		urls[i] = builder.String()
	}

	//protocol detail is on https://github.com/0chain/blobber/wiki/Features-Upload#upload
	M := int(math.Ceil(float64(T) / float64(3) * float64(2))) //the minimum of M blobbers must accpet the marker

	//retry 3 times
	for retry := 0; ; retry++ {
		i := 1
		n := 0
		m.lockedBlobbers = nil

		body := url.Values{}
		body.Set("connection_id", connectionID)

		now := time.Now()
		body.Set("request_time", strconv.FormatInt(now.Unix(), 10))

		form := body.Encode()

		for {

			// M locks are acquired, it is safe to commit write
			if n >= M {
				return nil
			}

			// No more blobber, but n < M
			if i > T && n < M {
				//fails, release all locks
				err := m.Unlock(ctx, connectionID)
				if err != nil {
					return err
				}
				break
			}

			blobberIndex := i - 1

			result, err := m.lockOne(ctx, strings.NewReader(form), urls[blobberIndex])

			// current blobber fails or is down, try next blobber
			if err != nil {
				// fails on current blobber
				logger.Logger.Error("lock: ", err.Error())
				i++
				continue
			}

			// it is locked by other session, wait and retry
			if result.Status == WMLockStatusPending {
				logger.Logger.Info("WriteMarkerLock is pending, wait and retry")
				sys.Sleep(1 * time.Second)
				continue
			} else if result.Status == WMLockStatusOK {
				// locked on current blobber, count it and go to next blobber
				if m.lockedBlobbers == nil {
					m.lockedBlobbers = make(map[string]bool)
				}

				m.lockedBlobbers[m.allocationObj.Blobbers[blobberIndex].ID] = true
				i++
				n++
			}
		}

		if retry >= 2 {
			return constants.ErrNotLockedWritMarker
		}

		sys.Sleep(1 * time.Second)
	}
}

// lockOne acquire WriteMarker lock from a blobber
func (m *WriteMarkerMutex) lockOne(ctx context.Context, body io.Reader, url string) (*WMLockResult, error) {

	result := &WMLockResult{}

	options := []resty.Option{
		resty.WithRequestInterceptor(func(r *http.Request) {
			m.zbox.SignRequest(r, m.allocationObj.Tx) //nolint
		}),
		resty.WithHeader(map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		})}

	r := resty.New(options...)

	r.DoPost(ctx, body, url).
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			if err != nil {
				return err
			}

			if resp.StatusCode != http.StatusOK {
				return errors.Throw(constants.ErrNotLockedWritMarker, fmt.Sprint(resp.StatusCode, ":", string(respBody)))
			}

			err = json.Unmarshal(respBody, result)

			if err != nil {
				return err
			}

			return nil
		})

	err := r.Wait()
	if len(err) > 0 {
		return nil, err[0]
	}

	return result, nil
}

// Unlock release WriteMarker lock on blobbers
func (m *WriteMarkerMutex) Unlock(ctx context.Context, connectionID string) error {
	if m == nil {
		return errors.Throw(constants.ErrInvalidParameter, "WriteMarkerMutex")
	}

	if m.allocationObj == nil {
		return errors.Throw(constants.ErrInvalidParameter, "allocationObj")
	}

	T := len(m.allocationObj.Blobbers)

	// no blobbers, it is unnecessary to release locks
	if T == 0 {
		return nil
	}

	urls := make([]string, 0, T)

	builder := &strings.Builder{}
	for _, b := range m.allocationObj.Blobbers {
		builder.Reset()
		builder.WriteString(strings.TrimRight(b.Baseurl, "/")) //nolint: errcheck
		builder.WriteString(blobber.EndpointWriteMarkerLock)
		builder.WriteString(m.allocationObj.Tx)
		builder.WriteString("/")
		builder.WriteString(connectionID)

		blobberUrl := builder.String()
		// only release lock on locked blobbers
		if m.lockedBlobbers != nil {
			if m.lockedBlobbers[b.ID] {
				urls = append(urls, blobberUrl)
			}
		} else { // Lock is not called here, try to release all blobbers
			urls = append(urls, blobberUrl)
		}

	}

	options := []resty.Option{
		resty.WithRequestInterceptor(func(r *http.Request) {
			m.zbox.SignRequest(r, m.allocationObj.Tx) //nolint
		}),
	}

	r := resty.New(options...)

	r.DoDelete(ctx, urls...)

	errs := r.Wait()

	if len(errs) == 0 {
		return nil
	}

	msgList := make([]string, 0, len(errs))
	for _, err := range errs {
		msgList = append(msgList, err.Error())
	}

	return errors.Throw(constants.ErrNotUnlockedWritMarker, msgList...)
}

// GetRootHashnode get root hash node from blobber
func (m *WriteMarkerMutex) GetRootHashnode(ctx context.Context, blobberBaseUrl string) (*fileref.Hashnode, error) {
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: resty.DefaultDialTimeout,
		}).Dial,
		TLSHandshakeTimeout: resty.DefaultDialTimeout,
	}

	root := &fileref.Hashnode{}

	r := resty.New(resty.WithTransport(transport))

	r.DoGet(ctx, fmt.Sprint(blobberBaseUrl, blobber.EndpointRootHashnode, m.allocationObj.Tx)).
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {

			if err != nil {
				return errors.Throw(constants.ErrInvalidHashnode, err.Error())
			}

			if resp.StatusCode != http.StatusOK {
				return errors.Throw(constants.ErrBadRequest, resp.Status)
			}

			if respBody != nil {
				if err := json.Unmarshal(respBody, root); err != nil {
					return errors.Throw(constants.ErrInvalidHashnode, err.Error())
				}

				return nil
			}

			return errors.Throw(constants.ErrInvalidHashnode, "no data")

		})

	errs := r.Wait()

	if len(errs) > 0 {
		return nil, errs[0]
	}

	return root, nil
}
