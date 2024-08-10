package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Printf("Now = %v\n", Now())
}

// Timestamp represents Unix time (e.g. in seconds)
type Timestamp int64

// Now - current datetime
func Now() Timestamp {
	return Timestamp(time.Now().Unix())
}
