package sdk

// Tests for the chunk-upload retry/backoff policy added to handle transient
// broken-pipe / closed-connection resets on the gateway->blobber :5051 write
// (see chunked_upload_blobber.go). The upload path uses fasthttp, which the
// net/http mock harness used by the other sdk tests cannot intercept, so these
// tests target the shared decision helpers + the loop's control flow directly.

import (
	stderrors "errors"
	"fmt"
	"io"
	"syscall"
	"testing"
	"time"

	"github.com/hitenjain14/fasthttp"
)

func TestIsRetryableUploadErr(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"broken pipe (EPIPE)", syscall.EPIPE, true},
		{"connection reset (ECONNRESET)", syscall.ECONNRESET, true},
		{"fasthttp connection closed", fasthttp.ErrConnectionClosed, true},
		{"EPIPE wrapped", fmt.Errorf("write tcp ->blobber:5051: %w", syscall.EPIPE), true},
		{"conn-closed wrapped", fmt.Errorf("upload: %w", fasthttp.ErrConnectionClosed), true},
		// Bare unwrapped errorStrings — the shapes the simulation actually produced
		// and that the old two-sentinel check missed:
		{"bare 'connection closed'", stderrors.New("connection closed"), true},
		{"bare 'broken pipe'", stderrors.New("write tcp 172.18.0.8->:5051: write: broken pipe"), true},
		{"bare 'connection reset by peer'", stderrors.New("read: connection reset by peer"), true},
		{"use of closed network connection", stderrors.New("use of closed network connection"), true},
		{"unexpected EOF", io.ErrUnexpectedEOF, true},
		// Non-connection errors must NOT retry:
		{"plain EOF is not a drop", io.EOF, false},
		{"generic 400", stderrors.New("400 bad request"), false},
		{"file not found", stderrors.New("file_meta_error: not found"), false},
	}
	for _, c := range cases {
		if got := isRetryableUploadErr(c.err); got != c.want {
			t.Errorf("%s: isRetryableUploadErr(%v) = %v, want %v", c.name, c.err, got, c.want)
		}
	}
}

func TestUploadRetryBackoff(t *testing.T) {
	if d := uploadRetryBackoff(0); d != 0 {
		t.Errorf("first attempt must not wait: got %v, want 0", d)
	}
	if d := uploadRetryBackoff(1); d != 200*time.Millisecond {
		t.Errorf("attempt 1 backoff = %v, want 200ms", d)
	}
	if d := uploadRetryBackoff(4); d != 800*time.Millisecond {
		t.Errorf("attempt 4 backoff = %v, want 800ms", d)
	}
	// Must be monotonically non-decreasing so pressure eases across attempts.
	prev := time.Duration(-1)
	for i := 0; i < uploadMaxAttempts; i++ {
		d := uploadRetryBackoff(i)
		if d < prev {
			t.Errorf("backoff decreased at attempt %d: %v < %v", i, d, prev)
		}
		prev = d
	}
}

func TestUploadMaxAttempts(t *testing.T) {
	// The chunk path must retry more than once (the bug was effectively too few
	// under pressure) and stay in line with the commit path's 6x — allow [3,6].
	if uploadMaxAttempts < 3 || uploadMaxAttempts > 6 {
		t.Errorf("uploadMaxAttempts = %d, want between 3 and 6", uploadMaxAttempts)
	}
}

// runUploadRetry mirrors the production loop's control flow in
// chunked_upload_blobber.go: up to uploadMaxAttempts tries, backing off before
// each retry, retrying only when the attempt says so. `do` returns (err,
// retryable) exactly like the production per-attempt closure. Kept here so the
// retry SEMANTICS (attempt count, retry-only-on-retryable, immediate stop on a
// hard error) are pinned by a test even though the real loop is inline in a
// concurrent hot path we don't want to restructure.
func runUploadRetry(do func(attempt int) (err error, retryable bool)) (attempts int, err error) {
	for i := 0; i < uploadMaxAttempts; i++ {
		// backoff kept tiny in the mirror; production uses uploadRetryBackoff(i).
		attempts++
		var retry bool
		err, retry = do(i)
		if retry {
			continue
		}
		return attempts, err
	}
	return attempts, err
}

func TestUploadRetry_RecoversWhenTransientThenSucceeds(t *testing.T) {
	fail := 2 // broken pipe on the first 2 tries, then succeed
	attempts, err := runUploadRetry(func(i int) (error, bool) {
		if i < fail {
			return syscall.EPIPE, true
		}
		return nil, false
	})
	if err != nil {
		t.Fatalf("expected recovery after %d transient failures, got err=%v", fail, err)
	}
	if attempts != fail+1 {
		t.Errorf("attempts = %d, want %d (fail %d then succeed)", attempts, fail+1, fail)
	}
}

func TestUploadRetry_ExhaustsAndReturnsError(t *testing.T) {
	attempts, err := runUploadRetry(func(i int) (error, bool) {
		return syscall.EPIPE, true // always broken pipe
	})
	if !isRetryableUploadErr(err) {
		t.Fatalf("exhausted retry must return the transient error, got %v", err)
	}
	if attempts != uploadMaxAttempts {
		t.Errorf("attempts = %d, want %d (all attempts used)", attempts, uploadMaxAttempts)
	}
}

func TestUploadRetry_HardErrorStopsImmediately(t *testing.T) {
	hard := stderrors.New("400 bad request")
	attempts, err := runUploadRetry(func(i int) (error, bool) {
		return hard, false // non-retryable
	})
	if err != hard {
		t.Fatalf("hard error should propagate as-is, got %v", err)
	}
	if attempts != 1 {
		t.Errorf("a non-retryable error must not retry: attempts = %d, want 1", attempts)
	}
}
