package main

import (
	"errors"
	"fmt"
	"net/textproto"
	"strconv"
	"strings"
)

var errNoOverlap = errors.New("invalid range: failed to overlap")

type httpRange struct {
	start, length, total int64
}

func (r httpRange) Header() string {
	return fmt.Sprintf("bytes %d-%d/%d", r.start, r.start+r.length-1, r.total)
}

// parseRange parses a Range header string as per RFC 7233.
// errNoOverlap is returned if none of the ranges overlap.
//
// Examples of valid byte ranges (assuming a resource of length 10240):
// bytes=0-499, the first 500 bytes
// bytes=1000-1999, 1000 bytes start from offset 1000
// bytes=-500, the final 500 bytes (byte offsets 9739-10239, inclusive)
// bytes=0-0,-1, the first and last bytes only
// bytes=0-, bytes=0-10250, be interpreted as bytes=0-10239
func parseRange(s string, total int64) ([]httpRange, error) {
	if s == "" {
		return nil, nil // header not present
	}
	const b = "bytes="
	if !strings.HasPrefix(s, b) {
		return nil, errors.New("invalid range")
	}
	var ranges []httpRange
	noOverlap := false
	for _, ra := range strings.Split(s[len(b):], ",") {
		ra = textproto.TrimString(ra)
		if ra == "" {
			continue
		}
		start, end, ok := strings.Cut(ra, "-")
		if !ok {
			return nil, errors.New("invalid range")
		}
		start, end = textproto.TrimString(start), textproto.TrimString(end)
		var r httpRange
		if start == "" {
			if end == "" || end[0] == '-' {
				return nil, errors.New("invalid range")
			}
			i, err := strconv.ParseInt(end, 10, 64)
			if i < 0 || err != nil {
				return nil, errors.New("invalid range")
			}
			if i > total {
				i = total
			}
			r.start = total - i
			r.length = total - r.start
		} else {
			i, err := strconv.ParseInt(start, 10, 64)
			if err != nil || i < 0 {
				return nil, errors.New("invalid range")
			}
			if i >= total {
				// If the range begins after the size of the content,
				// then it does not overlap.
				noOverlap = true
				i = total - 1
				//continue, since server may attempt to return a largeEnoughLen, errNoOverlap never happen
			}
			r.start = i
			if end == "" {
				r.length = sizePerRequest
				if r.length > total-r.start && !noOverlap {
					r.length = total - r.start
				}
			} else {
				i, err := strconv.ParseInt(end, 10, 64)
				if err != nil || r.start > i {
					return nil, errors.New("invalid range")
				}
				if i >= total {
					i = total - 1
				}
				r.length = i - r.start + 1
			}
		}
		r.total = total
		ranges = append(ranges, r)
	}
	if noOverlap && len(ranges) == 0 {
		// The specified ranges did not overlap with the content.
		return nil, errNoOverlap
	}
	return ranges, nil
}
