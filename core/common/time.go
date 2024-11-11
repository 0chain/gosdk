package common

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// Timestamp represents Unix time (e.g. in seconds)
type Timestamp int64

// Now current datetime
func Now() Timestamp {
	return Timestamp(time.Now().Unix())
}

// Within ensures a given timestamp is within certain number of seconds
func (t Timestamp) Within(seconds Timestamp) bool {
	var now = Now()
	return now > t-seconds && now < t+seconds
}

// ToTime converts the Timestamp to standard time.Time
func (t Timestamp) ToTime() time.Time {
	return time.Unix(int64(t), 0)
}

var ErrInvalidTime = errors.New("invalid time")

// ParseTime parse a time string with 4 formats
//
//	+1h5m : now (local timezone) + 1h5m
//	+3900 : now (local timezone) +3900s
//	1647858200 : Unix timestamp
//	2022-03-21 10:21:38 : parse UTC date string with YYYY-MM-dd HH:mm:ss
//
// Parameters
//   - now is the current time
//   - input is the time string to parse
func ParseTime(now time.Time, input string) (*time.Time, error) {

	if len(input) == 0 {
		return nil, ErrInvalidTime
	}

	if input[0] == '+' {
		val := strings.TrimLeft(input, "+")
		s, err := strconv.Atoi(val)
		// +3900 : now (local timezone) +3900s
		if err == nil {
			now = now.Add(time.Duration(s) * time.Second)
			return &now, nil
		}

		d, err := time.ParseDuration(val)
		// +1h5m : now (local timezone) + 1h5m
		if err == nil {
			now = now.Add(d)
			return &now, nil
		}

		return nil, ErrInvalidTime
	}

	s, err := strconv.ParseInt(input, 10, 64)
	if err == nil {
		now = time.Unix(s, 0)
		return &now, nil
	}

	// 2022-03-21 10:21:38 : parse UTC date string with YYYY-MM-dd HH:mm:ss
	t, err := time.Parse("2006-01-02 15:04:05 -0700", input+" +0000")
	if err == nil {
		return &t, nil
	}

	return nil, ErrInvalidTime

}
