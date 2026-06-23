package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	zlogger "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/hitenjain14/fasthttp"
	"golang.org/x/sync/semaphore"
)

const (
	LockExists     = "lock_exists"
	RateLimitError = "rate_limit_error"
)

// blobberHTTPPort / blobberWriteHTTPPort, when set (e.g. "5051"), reroute blobber
// requests from the registered TLS URL (https://host, fronted by the cluster's
// Caddy TLS proxy) to the blobber's direct plaintext HTTP listener
// (http://host:<port>). CPU profiling of the gateway read path showed ~80% of CPU
// on TLS decrypt + per-record read syscalls + buffer copies (Reed-Solomon erasure
// was ~2%); the blobber already serves the identical router on its --port HTTP
// listener (Caddy merely proxies to it), reachable VPC-wide. Measured READ win:
// +40% S3 GET, +20% NFS read on a 4-vCPU gateway. Intended only for intra-VPC,
// erasure-coded clusters; the blobber does not validate Host.
//
// SIGNATURE CAVEAT (the reason read and write differ): the gosdk V2 client
// signature is Sign(Hash(allocation + baseURL)) and the blobber verifies it as
// allocation + node.Self.GetURLBase() — its OWN REGISTERED URL (https://host). So
// the request must be SIGNED over the registered URL even when CONNECTING to the
// plaintext :port. The read/download path rewrites the base URL BEFORE signing,
// which signs over the wrong URL — harmless because downloads don't enforce this
// V2 sig. Uploads/commits DO enforce it (verifySignatureFromRequest in WriteFile),
// so the write path must sign over the original URL and rewrite only the request
// URI (the connection target) AFTER the signature header is set — that's what
// rewriteFastReqURIPlaintext / rewriteHTTPReqURLPlaintext below do.
//
// READ (ZUS_BLOBBER_HTTP) and WRITE (ZUS_BLOBBER_WRITE_HTTP) are gated separately.
var (
	blobberHTTPPort      = strings.TrimSpace(os.Getenv("ZUS_BLOBBER_HTTP"))
	blobberWriteHTTPPort = strings.TrimSpace(os.Getenv("ZUS_BLOBBER_WRITE_HTTP"))
)

func rewritePlaintextURL(baseURL, port string) string {
	if port == "" {
		return baseURL
	}
	const scheme = "https://"
	if !strings.HasPrefix(baseURL, scheme) {
		return baseURL
	}
	host := baseURL[len(scheme):]
	if i := strings.IndexAny(host, ":/"); i >= 0 {
		host = host[:i]
	}
	return "http://" + host + ":" + port
}

// plaintextBlobberURL rewrites the base URL for the READ/download path BEFORE
// signing (ZUS_BLOBBER_HTTP). Safe only because downloads don't enforce the V2
// sig; do NOT use this on the write path (see SIGNATURE CAVEAT above).
func plaintextBlobberURL(baseURL string) string {
	return rewritePlaintextURL(baseURL, blobberHTTPPort)
}

// rewriteFastReqURIPlaintext retargets a fully-built fasthttp request to the
// blobber's plaintext :port listener AFTER its V2 signature header is set, so the
// signature stays over the original (registered) blobber URL that the blobber
// reconstructs in verifySignatureFromRequest. Used by the WRITE path.
func rewriteFastReqURIPlaintext(req *fasthttp.Request, port string) {
	if port == "" {
		return
	}
	u := req.URI()
	if string(u.Scheme()) != "https" {
		return
	}
	host := string(u.Host())
	if i := strings.IndexByte(host, ':'); i >= 0 {
		host = host[:i]
	}
	u.SetScheme("http")
	u.SetHost(host + ":" + port)
}

// rewriteHTTPReqURLPlaintext is the net/http analog (commit request).
func rewriteHTTPReqURLPlaintext(req *http.Request, port string) {
	if port == "" || req.URL == nil || req.URL.Scheme != "https" {
		return
	}
	host := req.URL.Hostname()
	req.URL.Scheme = "http"
	req.URL.Host = host + ":" + port
	req.Host = req.URL.Host
}

type BlockDownloadRequest struct {
	blobber            *blockchain.StorageNode
	allocationID       string
	allocationTx       string
	allocOwnerID       string
	blobberIdx         int
	maskIdx            int
	remotefilepath     string
	remotefilepathhash string
	chunkSize          int
	blockNum           int64
	encryptedKey       string
	contentMode        string
	numBlocks          int64
	authTicket         *marker.AuthTicket
	ctx                context.Context
	result             chan *downloadBlock
	shouldVerify       bool
	connectionID       string
	respBuf            []byte
}

type downloadResponse struct {
	Nodes   [][][]byte
	Indexes [][]int
	Data    []byte
}

type downloadBlock struct {
	BlockChunks [][]byte
	Success     bool               `json:"success"`
	LatestRM    *marker.ReadMarker `json:"latest_rm"`
	idx         int
	maskIdx     int
	err         error
	timeTaken   int64
}

var downloadBlockChan map[string]chan *BlockDownloadRequest
var initDownloadMutex sync.Mutex

func InitBlockDownloader(blobbers []*blockchain.StorageNode, workerCount int) {
	initDownloadMutex.Lock()
	defer initDownloadMutex.Unlock()
	if downloadBlockChan == nil {
		downloadBlockChan = make(map[string]chan *BlockDownloadRequest)
	}

	for _, blobber := range blobbers {
		if _, ok := downloadBlockChan[blobber.ID]; !ok {
			downloadBlockChan[blobber.ID] = make(chan *BlockDownloadRequest, workerCount)
			go startBlockDownloadWorker(downloadBlockChan[blobber.ID], workerCount)
		}
	}
}

func startBlockDownloadWorker(blobberChan chan *BlockDownloadRequest, workers int) {
	sem := semaphore.NewWeighted(int64(workers))
	fastClient := zboxutil.GetFastHTTPClient()
	for {
		blockDownloadReq, open := <-blobberChan
		if !open {
			break
		}
		if err := sem.Acquire(blockDownloadReq.ctx, 1); err != nil {
			blockDownloadReq.result <- &downloadBlock{Success: false, idx: blockDownloadReq.blobberIdx, err: err}
			continue
		}
		go func() {
			blockDownloadReq.downloadBlobberBlock(fastClient)
			sem.Release(1)
		}()
	}
}

func splitData(buf []byte, lim int) [][]byte {
	var chunk []byte
	chunks := make([][]byte, 0, common.MustAddInt(len(buf)/lim, 1))
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:])
	}
	return chunks
}

func (req *BlockDownloadRequest) downloadBlobberBlock(fastClient *fasthttp.Client) {
	if req.numBlocks <= 0 {
		req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: errors.New("invalid_request", "Invalid number of blocks for download")}
		return
	}
	retry := 0
	var err error
	for retry < 3 {
		if len(req.remotefilepath) > 0 {
			req.remotefilepathhash = fileref.GetReferenceLookup(req.allocationID, req.remotefilepath)
		}

		httpreq, err := zboxutil.NewFastDownloadRequest(plaintextBlobberURL(req.blobber.Baseurl), req.allocationID, req.allocationTx)
		if err != nil {
			req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: errors.Wrap(err, "Error creating download request")}
			return
		}

		header := &DownloadRequestHeader{}
		header.PathHash = req.remotefilepathhash
		header.BlockNum = req.blockNum
		header.NumBlocks = req.numBlocks
		header.VerifyDownload = req.shouldVerify
		header.ConnectionID = req.connectionID
		header.Version = "v2"

		if req.authTicket != nil {
			header.AuthToken, _ = json.Marshal(req.authTicket) //nolint: errcheck
		}
		if len(req.contentMode) > 0 {
			header.DownloadMode = req.contentMode
		}
		if req.chunkSize == 0 {
			req.chunkSize = CHUNK_SIZE
		}
		shouldRetry := false

		header.ToFastHeader(httpreq)

		err = func() error {
			now := time.Now()
			statuscode, respBuf, err := fastClient.GetWithRequest(httpreq, req.respBuf)
			fasthttp.ReleaseRequest(httpreq)
			timeTaken := time.Since(now).Milliseconds()
			if err != nil {
				zlogger.Logger.Error("Error downloading block: ", err)
				if errors.Is(err, fasthttp.ErrConnectionClosed) || errors.Is(err, syscall.EPIPE) {
					shouldRetry = true
					return errors.New("connection_closed", "Connection closed")
				}
				return err
			}

			if statuscode == http.StatusTooManyRequests {
				shouldRetry = true
				time.Sleep(time.Second * 2)
				return errors.New(RateLimitError, "Rate limit error")
			}

			if statuscode == http.StatusInternalServerError {
				shouldRetry = true
				return errors.New("internal_server_error", "Internal server error")
			}

			var rspData downloadBlock
			if statuscode != http.StatusOK {
				zlogger.Logger.Error(fmt.Sprintf("downloadBlobberBlock FAIL - blobberID: %v, clientID: %v, blockNum: %d, retry: %d, response: %v", req.blobber.ID, client.GetClientID(), header.BlockNum, retry, string(respBuf)))
				if err = json.Unmarshal(respBuf, &rspData); err == nil {
					return errors.New("download_error", fmt.Sprintf("Response status: %d, Error: %v,", statuscode, rspData.err))
				}
				return errors.New("response_error", string(respBuf))
			}

			dR := downloadResponse{}
			if req.shouldVerify {
				err = json.Unmarshal(respBuf, &dR)
				if err != nil {
					return err
				}
			} else {
				dR.Data = respBuf
			}
			if req.contentMode == DOWNLOAD_CONTENT_FULL && req.shouldVerify {
				zlogger.Logger.Info("verifying multiple blocks")
			}

			rspData.idx = req.blobberIdx
			rspData.maskIdx = req.maskIdx
			rspData.timeTaken = timeTaken
			rspData.Success = true

			if req.encryptedKey != "" {
				if req.authTicket != nil {
					// ReEncryptionHeaderSize for the additional header bytes for ReEncrypt,  where chunk_size - EncryptionHeaderSize is the encrypted data size
					rspData.BlockChunks = splitData(dR.Data, req.chunkSize-EncryptionHeaderSize+ReEncryptionHeaderSize)
				} else {
					rspData.BlockChunks = splitData(dR.Data, req.chunkSize)
				}
			} else {
				if req.chunkSize == 0 {
					req.chunkSize = CHUNK_SIZE
				}
				rspData.BlockChunks = splitData(dR.Data, req.chunkSize)
			}

			zlogger.Logger.Debug(fmt.Sprintf("downloadBlobberBlock 200 OK: blobberID: %v, clientID: %v, blockNum: %d", req.blobber.ID, client.GetClientID(), header.BlockNum))

			req.result <- &rspData
			return nil
		}()

		if err != nil {
			if shouldRetry {
				if retry >= 3 {
					req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: err}
					return
				}
				shouldRetry = false
				zlogger.Logger.Debug("Retrying for Error occurred: ", err)
				retry++
				continue
			} else {
				req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: err, maskIdx: req.maskIdx}
			}
		}
		return
	}

	req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: err, maskIdx: req.maskIdx}

}

func AddBlockDownloadReq(ctx context.Context, req *BlockDownloadRequest, rb zboxutil.DownloadBuffer, effectiveBlockSize int) {
	if rb != nil {
		reqCtx, cncl := context.WithTimeout(ctx, (time.Second * 45))
		defer cncl()
		req.respBuf = rb.RequestChunk(reqCtx, int(req.blockNum))
		if len(req.respBuf) == 0 {
			req.respBuf = make([]byte, int(req.numBlocks)*effectiveBlockSize)
		}
	} else {
		req.respBuf = make([]byte, int(req.numBlocks)*effectiveBlockSize)
	}
	downloadBlockChan[req.blobber.ID] <- req
}
