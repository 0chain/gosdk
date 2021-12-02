//go:build js && wasm
// +build js,wasm

package jsbridge

import (
	"fmt"
	"syscall/js"
)

func NewArray(items ...interface{}) js.Value {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("NewStrings: ", r)
		}
	}()

	list := js.Global().Get("Array").New()

	for _, it := range items {
		list.Call("push", js.ValueOf(it))
	}

	return list
}
