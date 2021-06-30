package main

import (
	// "fmt"
	// "strconv"
	"syscall/js"
	"github.com/0chain/gosdk/zcncore"
)

// // JS does not have int64 so we must take a string instead of int64.
// func strToInt64(s string) int64 {
// 	tokens, err := strconv.ParseInt(s, 10, 64)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return tokens
// }
//
// func int64ToStr(x int64) string {
// 	return strconv.FormatInt(x, 10)
// }

func InitZCNSDK(this js.Value, p []js.Value) interface{} {
	blockWorker := p[0].String()
	signscheme := p[1].String()
	err := zcncore.InitZCNSDK(blockWorker, signscheme)
	if err != nil {
		return err.Error()
	}
	return nil
}
