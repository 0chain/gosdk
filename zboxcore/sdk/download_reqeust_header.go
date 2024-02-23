package sdk

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hitenjain14/fasthttp"
)

// DownloadRequestHeader download request header
type DownloadRequestHeader struct {
	ClientID       string
	PathHash       string
	BlockNum       int64
	NumBlocks      int64
	ReadMarker     []byte
	AuthToken      []byte
	DownloadMode   string
	VerifyDownload bool
	ConnectionID   string
	Version        string
}

// ToHeader update header
func (h *DownloadRequestHeader) ToHeader(req *http.Request) {
	if h.PathHash != "" {
		req.Header.Set("X-Path-Hash", h.PathHash)
	}

	if h.BlockNum > 0 {
		req.Header.Set("X-Block-Num", strconv.FormatInt(h.BlockNum, 10))
	}

	if h.NumBlocks > 0 {
		req.Header.Set("X-Num-Blocks", strconv.FormatInt(h.NumBlocks, 10))
	}

	if h.ReadMarker != nil {
		req.Header.Set("X-Read-Marker", string(h.ReadMarker))
	}

	if h.AuthToken != nil {
		token := base64.StdEncoding.EncodeToString(h.AuthToken)
		req.Header.Set("X-Auth-Token", token)
	}

	if h.DownloadMode != "" {
		req.Header.Set("X-Mode", h.DownloadMode)
	}

	if h.ConnectionID != "" {
		req.Header.Set("X-Connection-ID", h.ConnectionID)
	}

	if h.Version != "" {
		req.Header.Set("X-Version", h.Version)
	}

	req.Header.Set("X-Verify-Download", fmt.Sprint(h.VerifyDownload))
}

func (h *DownloadRequestHeader) ToFastHeader(req *fasthttp.Request) {
	if h.PathHash != "" {
		req.Header.Set("X-Path-Hash", h.PathHash)
	}

	if h.BlockNum > 0 {
		req.Header.Set("X-Block-Num", strconv.FormatInt(h.BlockNum, 10))
	}

	if h.NumBlocks > 0 {
		req.Header.Set("X-Num-Blocks", strconv.FormatInt(h.NumBlocks, 10))
	}

	if h.ReadMarker != nil {
		req.Header.Set("X-Read-Marker", string(h.ReadMarker))
	}

	if h.AuthToken != nil {
		token := base64.StdEncoding.EncodeToString(h.AuthToken)
		req.Header.Set("X-Auth-Token", token)
	}

	if h.DownloadMode != "" {
		req.Header.Set("X-Mode", h.DownloadMode)
	}

	if h.ConnectionID != "" {
		req.Header.Set("X-Connection-ID", h.ConnectionID)
	}

	req.Header.Set("X-Verify-Download", fmt.Sprint(h.VerifyDownload))
}
