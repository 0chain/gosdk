// DEPRECATED: This package is deprecated and will be removed in a future release.
package time

import (
	"time"
)

const (
	// RFC3339 is useful for formatting time.
	RFC3339 = "2006-01-02T15:04:05Z07:00"
)

type (
	// A Duration represents the elapsed time between two instants
	// as an int64 nanosecond count. The representation limits the
	// largest representable duration to approximately 290 years.
	Duration = time.Duration

	// Time is a copy of time from golang std lib
	// to avoid import it from other packages.
	Time = time.Time

	// Timestamp represents a wrapper to control the json encoding.
	Timestamp int64
)

// Now returns current Unix time.
func Now() Timestamp {
	return Timestamp(time.Now().Unix())
}

// NowTime returns the current local time.
func NowTime() Time {
	return time.Now()
}
