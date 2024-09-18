package sdk

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/0chain/gosdk/core/client"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/resty"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type PlaylistFile struct {
	LookupHash     string `gorm:"column:lookup_hash" json:"lookup_hash"`
	Name           string `gorm:"column:name" json:"name"`
	Path           string `gorm:"column:path" json:"path"`
	NumBlocks      int64  `gorm:"column:num_of_blocks" json:"num_of_blocks"`
	ParentPath     string `gorm:"column:parent_path" json:"parent_path"`
	Size           int64  `gorm:"column:size;" json:"size"`
	ActualFileSize int64
	MimeType       string `gorm:"column:mimetype" json:"mimetype"`
	Type           string `gorm:"column:type" json:"type"`
}

func GetPlaylist(ctx context.Context, alloc *Allocation, path, since string) ([]PlaylistFile, error) {

	q := &url.Values{}
	q.Add("path", path)
	q.Add("since", since)

	return getPlaylistFromBlobbers(ctx, alloc, q.Encode())
}

func GetPlaylistByAuthTicket(ctx context.Context, alloc *Allocation, authTicket, lookupHash, since string) ([]PlaylistFile, error) {

	q := &url.Values{}
	q.Add("auth_token", authTicket)
	q.Add("lookup_hash", lookupHash)
	q.Add("since", since)

	return getPlaylistFromBlobbers(ctx, alloc, q.Encode())
}

func getPlaylistFromBlobbers(ctx context.Context, alloc *Allocation, query string) ([]PlaylistFile, error) {

	urls := make([]string, len(alloc.Blobbers))
	for i, b := range alloc.Blobbers {
		sb := &strings.Builder{}
		sb.WriteString(strings.TrimRight(b.Baseurl, "/"))
		sb.WriteString(zboxutil.PLAYLIST_LATEST_ENDPOINT)
		sb.WriteString(alloc.ID)
		sb.WriteString("?")
		sb.WriteString(query)

		urls[i] = sb.String()
	}

	opts := make([]resty.Option, 0, 3)

	opts = append(opts, resty.WithRetry(resty.DefaultRetry))
	opts = append(opts, resty.WithRequestInterceptor(func(req *http.Request) error {
		req.Header.Set("X-App-Client-ID", client.Id())
		req.Header.Set("X-App-Client-Key", client.PublicKey())

		hash := encryption.Hash(alloc.ID)
		sign, err := sys.Sign(hash, client.SignatureScheme(), client.GetClientSysKeys())
		if err != nil {
			return err
		}

		// ClientSignatureHeader represents http request header contains signature.
		req.Header.Set("X-App-Client-Signature", sign)

		return nil
	}))

	c := createPlaylistConsensus(alloc.getConsensuses())

	r := resty.New(opts...).
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			if err != nil {
				logger.Logger.Error("playlist: ", err)
				return err
			}

			if resp != nil {
				if resp.StatusCode == http.StatusOK {
					if e := c.AddFiles(respBody); e != nil {
						logger.Logger.Error("playlist: ", e, resp.Request.URL)
					}

					return nil
				}

				logger.Logger.Error("playlist: ", resp.Status, resp.Request.URL)
				return nil

			}

			return nil
		})

	r.DoGet(ctx, urls...)

	r.Wait()

	return c.GetConsensusResult(), nil
}

func GetPlaylistFile(ctx context.Context, alloc *Allocation, path string) (*PlaylistFile, error) {
	q := &url.Values{}
	q.Add("lookup_hash", fileref.GetReferenceLookup(alloc.ID, path))

	return getPlaylistFileFromBlobbers(ctx, alloc, q.Encode())
}

func GetPlaylistFileByAuthTicket(ctx context.Context, alloc *Allocation, authTicket, lookupHash string) (*PlaylistFile, error) {
	q := &url.Values{}
	q.Add("auth_token", authTicket)
	q.Add("lookup_hash", lookupHash)

	return getPlaylistFileFromBlobbers(ctx, alloc, q.Encode())
}

func getPlaylistFileFromBlobbers(ctx context.Context, alloc *Allocation, query string) (*PlaylistFile, error) {

	urls := make([]string, len(alloc.Blobbers))
	for i, b := range alloc.Blobbers {
		sb := &strings.Builder{}
		sb.WriteString(strings.TrimRight(b.Baseurl, "/"))
		sb.WriteString(zboxutil.PLAYLIST_FILE_ENDPOINT)
		sb.WriteString(alloc.ID)
		sb.WriteString("?")
		sb.WriteString(query)

		urls[i] = sb.String()
	}

	opts := make([]resty.Option, 0, 3)

	opts = append(opts, resty.WithRetry(resty.DefaultRetry))
	opts = append(opts, resty.WithRequestInterceptor(func(req *http.Request) error {
		req.Header.Set("X-App-Client-ID", client.Id())
		req.Header.Set("X-App-Client-Key", client.PublicKey())

		hash := encryption.Hash(alloc.ID)
		sign, err := sys.Sign(hash, client.SignatureScheme(), client.GetClientSysKeys())
		if err != nil {
			return err
		}

		// ClientSignatureHeader represents http request header contains signature.
		req.Header.Set("X-App-Client-Signature", sign)

		return nil
	}))

	c := createPlaylistConsensus(alloc.getConsensuses())

	r := resty.New(opts...).
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			if err != nil {
				logger.Logger.Error("playlist: ", err)
				return err
			}

			if resp != nil {
				if resp.StatusCode == http.StatusOK {
					if e := c.AddFile(respBody); e != nil {
						logger.Logger.Error("playlist: ", e, resp.Request.URL)
					}

					return nil
				}

				logger.Logger.Error("playlist: ", resp.Status, resp.Request.URL)
				return nil

			}

			return nil
		})

	r.DoGet(ctx, urls...)

	r.Wait()

	files := c.GetConsensusResult()

	if len(files) > 0 {
		return &files[0], nil
	}

	return nil, errors.New("playlist: playlist file not found")
}
