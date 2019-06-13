package common

import (
	"math/rand"
	"time"
)

/*Now - current datetime */
func Now() int64 {
	return time.Now().Unix()
}

/*Within ensures a given timestamp is within certain number of seconds */
func Within(ts int64, seconds int64) bool {
	now := time.Now().Unix()
	return now > ts-seconds && now < ts+seconds
}

var randGenerator = rand.New(rand.NewSource(time.Now().UnixNano()))
