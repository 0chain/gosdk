//go:build js && wasm
// +build js,wasm

package jsbridge

import (
	"reflect"
	"syscall/js"
)

var (
	jsFuncList = make([]js.Func, 0, 200)
)

var (
	TypeFunc   = reflect.TypeOf(func() {}).String()
	TypeError  = "error"
	TypeString = reflect.TypeOf("string").String()
)

func Close() {
	for _, fn := range jsFuncList {
		fn.Release()
	}
}
