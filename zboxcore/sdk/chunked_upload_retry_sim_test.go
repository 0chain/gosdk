package sdk

// Faithful simulation of the production failure: a blobber that DROPS the TCP
// connection (the "write ...:5051: broken pipe / connection reset" Olyad saw),
// driven through the REAL fasthttp client the upload path uses. Proves what error
// fasthttp actually surfaces on a real socket drop and that isRetryableUploadErr
// catches it. (Recovery-after-N-then-success is covered deterministically by the
// runUploadRetry tests in chunked_upload_retry_test.go — fasthttp's own internal
// retry of pre-first-byte drops makes a count-based TCP recovery test racy, so
// the loop semantics are tested at the layer this fix owns.)

import (
	"net"
	"testing"
	"time"

	"github.com/hitenjain14/fasthttp"
)

// alwaysDropServer accepts TCP connections on loopback and, on each, forces an
// RST (SetLinger(0) + close) after reading a little — never returning a response.
// A real socket, so the client sees real drop errors (broken pipe / reset /
// connection closed), not an in-memory-pipe artifact.
func alwaysDropServer(t *testing.T) (*fasthttp.Client, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			if tc, ok := c.(*net.TCPConn); ok {
				_ = tc.SetLinger(0) // RST on close
			}
			b := make([]byte, 64)
			_ = c.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
			_, _ = c.Read(b) // let the client start writing, then cut it
			_ = c.Close()
		}
	}()
	addr := ln.Addr().String()
	client := &fasthttp.Client{
		Dial: func(_ string) (net.Conn, error) { return net.Dial("tcp", addr) },
	}
	return client, func() { _ = ln.Close() }
}

func TestSim_RealTCPDropIsRetryable(t *testing.T) {
	client, stop := alwaysDropServer(t)
	defer stop()

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	req.SetRequestURI("http://blobber/v1/file/upload/alloc")
	req.Header.SetMethod("POST")
	// A body large enough that the client is still writing when the server RSTs —
	// this is the mid-write "broken pipe" case fasthttp cannot auto-replay, i.e.
	// the one that reaches the gosdk retry loop in production.
	req.SetBody(make([]byte, 1<<20))

	err := client.DoTimeout(req, resp, 3*time.Second)
	if err == nil {
		t.Fatal("expected an error when the server always drops the connection")
	}
	t.Logf("real-TCP drop surfaced: %T: %v", err, err)
	if !isRetryableUploadErr(err) {
		t.Fatalf("isRetryableUploadErr did NOT catch the real TCP drop error (%v) — broaden connDropMsgs", err)
	}
}
