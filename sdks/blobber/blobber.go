// package blobber  wrap blobber's apis as sdk client
package blobber

import (
	"bytes"
	"context"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/sdks"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

// Blobber blobber sdk client instance
type Blobber struct {
	BaseURLs []string
	*sdks.ZBox
}

func New(zbox *sdks.ZBox, baseURLs ...string) *Blobber {
	b := &Blobber{
		BaseURLs: baseURLs,
		ZBox:     zbox,
	}

	return b
}

// CreateDir create new directories on blobbers
func (b *Blobber) CreateDir(allocationID, name string) error {
	if allocationID == "" {
		return errors.Throw(constants.ErrInvalidParameter, "allocationID")
	}

	if name == "" {
		return errors.Throw(constants.ErrInvalidParameter, "name")
	}

	req := &sdks.Request{}
	req.AllocationID = allocationID
	req.ConnectionID = zboxutil.NewConnectionId()

	body := &bytes.Buffer{}
	formWriter := multipart.NewWriter(body)
	err := formWriter.WriteField("connection_id", req.ConnectionID)
	if err != nil {
		return errors.Throw(constants.ErrInvalidOperation, "connection_id", req.ConnectionID)
	}
	err = formWriter.WriteField("name", name)
	if err != nil {
		return errors.Throw(constants.ErrInvalidOperation, "name", name)
	}
	formWriter.Close()
	req.Body = body
	req.ContentType = formWriter.FormDataContentType()

	msgList := make([]string, 0, len(b.BaseURLs))

	r := b.DoPost(req, func(req *http.Request, resp *http.Response, cf context.CancelFunc, err error) error {
		if err != nil {

			msgList = append(msgList, err.Error())
			return err
		}

		if resp != nil {
			//success
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
				return nil
			}

			buf, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				msgList = append(msgList, err.Error())
				return err
			}

			msgList = append(msgList, strconv.Itoa(resp.StatusCode)+":"+string(buf))
			return errors.Throw(constants.ErrInvalidOperation, strconv.Itoa(resp.StatusCode))
		}

		return constants.ErrUnknown

	})

	urls := b.BuildUrls(b.BaseURLs, nil, EndpointDirNew, allocationID)

	r.DoPut(context.TODO(), req.Body, urls...)

	errs := r.Wait()

	if len(errs) > 0 {
		return errors.Throw(constants.ErrBadRequest, msgList...)
	}

	return nil
}
