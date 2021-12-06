//go:build js && wasm
// +build js,wasm

package jsbridge

import (
	"encoding/json"
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

func NewJsError(message interface{}) js.Value {
	return js.ValueOf(map[string]interface{}{
		"error": fmt.Sprint(message),
	})
}

func NewObject(obj interface{}) js.Value {
	buf, err := json.Marshal(obj)
	if err != nil {
		return js.Null()
	}

	j := js.Global().Get("JSON")

	return j.Call("parse", string(buf))
}
