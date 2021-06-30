package main

import (
	"fmt"
	"syscall/js"
	"github.com/0chain/gosdk/zcncore"
)

func InitZCNSDK(this js.Value, p []js.Value) interface{} {
	blockWorker := p[0].String()
	signscheme := p[1].String()

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		// reject := args[1]

		go func() {
			err := zcncore.InitZCNSDK(blockWorker, signscheme)
			if err != nil {
				fmt.Println("error:", err)
			}
			resolve.Invoke()
		}()

		return nil
	})

	// Create and return the Promise object
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}
