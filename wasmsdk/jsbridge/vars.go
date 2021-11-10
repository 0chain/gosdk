// +build js,wasm

package jsbridge

import (
	"errors"
	"reflect"
	"syscall/js"
)

var (
	jsFuncList = make([]js.Func, 0, 200)
)

var (
	TypeFunc   = reflect.TypeOf(func() {}).String()
	TypeError  = reflect.TypeOf(errors.New("type")).String()
	TypeString = reflect.TypeOf("string").String()
)

func Close() {
	for _, fn := range jsFuncList {
		fn.Release()
	}
}
