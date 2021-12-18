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

func NewBytes(buf []byte) js.Value {

	uint8Array := js.Global().Get("Uint8Array").New(len(buf))

	js.CopyBytesToJS(uint8Array, buf)

	return uint8Array
}

// var arrayBuffer = new ArrayBuffer(100);
// var uint8Array = new Uint8Array(arrayBuffer);
// for (var i = 0; i < 100; i++) {
// 	uint8Array[i] = i;
// }

// var blob = new Blob([uint8Array], { type: "image/png" });
// var blobVal = URL.createObjectURL(blob);
// func CreateObjectURL(buf []byte) string {
// 	j := js.Global().Get("URL")

// 	options := js.Global().Get("Object").New()
// 	options.Set("type", "")

// 	blob := js.Global().Get("Blob").New(args ...interface{})

// 	u := j.Call("createObjectURL", object)

// 	return u.String()
// }
