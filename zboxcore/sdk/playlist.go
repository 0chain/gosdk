package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/resty"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type PlaylistFile struct {
	LookupHash string `gorm:"column:lookup_hash" json:"lookup_hash"`
	Name       string `gorm:"column:name" json:"name"`
	Path       string `gorm:"column:path" json:"path"`
	NumBlocks  int64  `gorm:"column:num_of_blocks" json:"num_of_blocks"`
	ParentPath string `gorm:"column:parent_path" json:"parent_path"`
	Size       int64  `gorm:"column:size;" json:"size"`
	MimeType   string `gorm:"column:mimetype" json:"mimetype"`
	Type       string `gorm:"column:type" json:"type"`
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
		sb.WriteString(zboxutil.PLAYLIST_ENDPOINT)
		sb.WriteString(alloc.ID)
		sb.WriteString("?")
		sb.WriteString(query)

		urls[i] = sb.String()
	}

	opts := make([]resty.Option, 0, 3)

	opts = append(opts, resty.WithRetry(resty.DefaultRetry))
	opts = append(opts, resty.WithTimeout(resty.DefaultRequestTimeout))
	opts = append(opts, resty.WithRequestInterceptor(func(req *http.Request) {
		req.Header.Set("X-App-Client-ID", client.GetClientID())
		req.Header.Set("X-App-Client-Key", client.GetClientPublicKey())

		hash := encryption.Hash(alloc.ID)
		sign, err := sys.Sign(hash, client.GetClient().SignatureScheme, client.GetClientSysKeys())
		if err != nil {
			logger.Logger.Error("playlist: ", err)
		}

		// ClientSignatureHeader represents http request header contains signature.
		req.Header.Set("X-App-Client-Signature", sign)
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
					if e := c.Add(respBody); e != nil {
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

type playlistConsensus struct {
	files       map[string]PlaylistFile
	consensuses map[string]*Consensus

	threshConsensus float32
	fullConsensus   float32
	consensusOK     float32
}

func createPlaylistConsensus(fullConsensus, threshConsensus, consensusOK float32) *playlistConsensus {
	return &playlistConsensus{
		files:           make(map[string]PlaylistFile),
		consensuses:     make(map[string]*Consensus),
		threshConsensus: threshConsensus,
		fullConsensus:   fullConsensus,
		consensusOK:     consensusOK,
	}
}

func (c *playlistConsensus) Add(body []byte) error {
	var files []PlaylistFile

	if err := json.Unmarshal([]byte(body), &files); err != nil {
		return err
	}

	for _, f := range files {
		_, ok := c.files[f.LookupHash]

		if ok {
			c.consensuses[f.LookupHash].Done()
		} else {
			cons := &Consensus{}

			cons.Init(c.threshConsensus, c.fullConsensus, c.consensusOK)
			cons.Done()

			c.consensuses[f.LookupHash] = cons
			c.files[f.LookupHash] = f
		}

	}

	return nil

}

func (c *playlistConsensus) GetConsensusResult() []PlaylistFile {

	files := make([]PlaylistFile, 0, len(c.files))

	for _, file := range c.files {
		cons := c.consensuses[file.LookupHash]
		fmt.Println(file.Name, cons.consensus, cons.getConsensusRate(), cons.getConsensusRequiredForOk())
		if cons.isConsensusOk() {
			files = append(files, file)
		}
	}

	sort.Slice(files, func(i, j int) bool {
		l := files[i]
		r := files[j]

		if len(l.Name) < len(r.Name) {
			return true
		}

		if len(l.Name) > len(r.Name) {
			return false
		}

		return l.Name < r.Name
	})

	return files
}
