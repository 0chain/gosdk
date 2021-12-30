package common

import (
	"time"
)

// Timestamp represents Unix time (e.g. in seconds)
type Timestamp int64

// Now - current datetime
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
