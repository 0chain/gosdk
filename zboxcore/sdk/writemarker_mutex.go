package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/resty"
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
	RootNode  *HashNode    `json:"root_node,omitempty"`
}

// HashNode ref node in hash tree
type HashNode struct {
	// hash data
	AllocationID   string          `json:"allocation_id,omitempty"`
	Type           string          `json:"type,omitempty"`
	Name           string          `json:"name,omitempty"`
	Path           string          `json:"path,omitempty"`
	ContentHash    string          `json:"content_hash,omitempty"`
	MerkleRoot     string          `json:"merkle_root,omitempty"`
	ActualFileHash string          `json:"actual_file_hash,omitempty"`
	Attributes     json.RawMessage `json:"attributes,omitempty"`
	ChunkSize      int64           `json:"chunk_size,omitempty"`
	Size           int64           `json:"size,omitempty"`
	ActualFileSize int64           `json:"actual_file_size,omitempty"`

	Children []*HashNode `json:"children,omitempty"`
}

func (n *HashNode) AddChild(c *HashNode) {
	if n.Children == nil {
		n.Children = make([]*HashNode, 0, 10)
	}

	n.Children = append(n.Children, c)
}

// GetLookupHash get lookuphash
func (n *HashNode) GetLookupHash() string {
	return encryption.Hash(n.AllocationID + ":" + n.Path)
}

// GetHashCode get hash code
func (n *HashNode) GetHashCode() string {

	if len(n.Attributes) == 0 {
		n.Attributes = json.RawMessage("{}")
	}
	hashArray := []string{
		n.AllocationID,
		n.Type,
		n.Name,
		n.Path,
		strconv.FormatInt(n.Size, 10),
		n.ContentHash,
		n.MerkleRoot,
		strconv.FormatInt(n.ActualFileSize, 10),
		n.ActualFileHash,
		string(n.Attributes),
		strconv.FormatInt(n.ChunkSize, 10),
	}

	return strings.Join(hashArray, ":")
}

// WriteMarkerMutex blobber WriteMarkerMutex client
type WriteMarkerMutex struct {
	allocationObj *Allocation
}

// CreateWriteMarkerMutex create WriteMarkerMutex for allocation
func CreateWriteMarkerMutex(allocationObj *Allocation) *WriteMarkerMutex {
	return &WriteMarkerMutex{
		allocationObj: allocationObj,
	}
}

// Lock acquire WriteMarker lock from blobbers
func (m *WriteMarkerMutex) Lock(ctx context.Context, sessionID string) error {

	if m == nil || m.allocationObj == nil {
		return errors.Throw(constants.ErrInvalidParameter, "allocationObj")
	}

	T := len(m.allocationObj.Blobbers)

	if T == 0 {
		return errors.Throw(constants.ErrInvalidParameter, "blobbers")
	}

	urls := make([]string, T)

	for i, blobber := range m.allocationObj.Blobbers {
		urls[i] = path.Join(blobber.Baseurl, "/v1/writemarker/lock/"+m.allocationObj.Tx)
	}

	//protocol detail is on https://github.com/0chain/blobber/wiki/Features-Upload#upload
	i := 1
	M := int(math.Ceil(2 / float64(3*T))) //the minimum of M blobbers must accpet the marker
	n := 0

	//retry 3 times
	for retry := 0; ; retry++ {

		body := url.Values{}
		body.Set("session_id", sessionID)

		now := time.Now()
		body.Set("request_time", strconv.FormatInt(now.Unix(), 10))
		buf := bytes.NewBufferString(body.Encode())

		for {
			result, err := m.lockOne(ctx, buf, urls[i-1])

			// current blobber fails or is down
			if err != nil {
				// last blobber, but n < M
				if i == T && n < M {
					//fails, release all locks
					err = m.Unlock(ctx, sessionID)
					if err != nil {
						return err
					}

					time.Sleep(1 * time.Second)
					break
				}

				// fails on current blobber, try next blobber
				i++
				time.Sleep(1 * time.Second)
			}

			// it is locked by other session, wait and retry
			if result.Status == WMLockStatusPending {
				time.Sleep(1 * time.Second)
				continue
			} else if result.Status == WMLockStatusOK {
				// locked on current blobber, count it and go to next blobber
				i++
				n++
			}

			// M locks are acquired, it is safe to commit write
			if n >= M {
				return nil
			}
		}

		if retry >= 2 {
			return constants.ErrNotLockedWritMarker
		}
	}
}

// lockOne acquire WriteMarker lock from a blobber
func (m *WriteMarkerMutex) lockOne(ctx context.Context, buf *bytes.Buffer, url string) (*WMLockResult, error) {
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: resty.DefaultDialTimeout,
		}).Dial,
		TLSHandshakeTimeout: resty.DefaultDialTimeout,
	}

	result := &WMLockResult{}

	r := resty.New(transport, func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
		if err != nil {
			return err
		}

		err = json.Unmarshal(respBody, result)

		if err != nil {
			return err
		}

		return nil
	})

	r.DoPost(ctx, buf, url)

	err := r.Wait()
	if len(err) > 0 {
		return nil, err[0]
	}

	return nil, nil
}

// Unlock release WriteMarker lock on blobbers
func (m *WriteMarkerMutex) Unlock(ctx context.Context, sessionID string) error {

	if m == nil || m.allocationObj == nil {
		return errors.Throw(constants.ErrInvalidParameter, "allocationObj")
	}

	T := len(m.allocationObj.Blobbers)

	// no blobbers, it is unnecessary to release locks
	if T == 0 {
		return nil
	}

	urls := make([]string, T)

	for i, blobber := range m.allocationObj.Blobbers {
		urls[i] = path.Join(blobber.Baseurl, "/v1/writemarker/lock/"+m.allocationObj.Tx)
	}

	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: resty.DefaultDialTimeout,
		}).Dial,
		TLSHandshakeTimeout: resty.DefaultDialTimeout,
	}

	r := resty.New(transport, func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
		return err
	})

	r.DoDelete(ctx, urls...)

	errs := r.Wait()

	if len(errs) == 0 {
		return nil
	}

	msgList := make(string, 0, len(errs))
	for _, err := range errs {
		msgList = append(msgList, err.Error())
	}

	return errors.Throw(constants.ErrNotUnlockedWritMarker, msgList...)
}
