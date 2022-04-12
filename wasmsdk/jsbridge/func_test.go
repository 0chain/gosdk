//go:build js && wasm
// +build js,wasm

package jsbridge

import (
	"syscall/js"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBindAsyncFunc(t *testing.T) {
	tests := []struct {
		Name   string
		Func   func() js.Func
		Output func(outputs []js.Value) interface{}
		Result interface{}
	}{
		{Name: "ReturnString", Func: func() js.Func {
			fn, _ := promise(func() string {
				return "ReturnString"
			})

			return fn
		}, Output: func(outputs []js.Value) interface{} {
			return outputs[0].String()
		}, Result: "ReturnString"},
		{Name: "ReturnInt", Func: func() js.Func {
			fn, _ := promise(func() int {
				return 1
			})

			return fn
		}, Output: func(outputs []js.Value) interface{} {
			return outputs[0].Int()
		}, Result: 1},
		{Name: "ReturnInt32", Func: func() js.Func {
			fn, _ := promise(func() int32 {
				return int32(1)
			})

			return fn
		}, Output: func(outputs []js.Value) interface{} {
			return int32(outputs[0].Int())
		}, Result: int32(1)},
		{Name: "ReturnInt64", Func: func() js.Func {
			fn, _ := promise(func() int64 {
				return int64(1)
			})

			return fn
		}, Output: func(outputs []js.Value) interface{} {
			return int64(outputs[0].Int())
		}, Result: int64(1)},
		{Name: "ReturnFloat32", Func: func() js.Func {
			fn, _ := promise(func() float32 {
				return float32(1)
			})

			return fn
		}, Output: func(outputs []js.Value) interface{} {
			return float32(outputs[0].Float())
		}, Result: float32(1)},
		{Name: "ReturnFloat64", Func: func() js.Func {
			fn, _ := promise(func() float64 {
				return float64(1)
			})

			return fn
		}, Output: func(outputs []js.Value) interface{} {
			return outputs[0].Float()
		}, Result: float64(1)},
	}

	for _, it := range tests {
		t.Run(it.Name, func(test *testing.T) {

			jsFunc := it.Func()

			outputs, _ := Await(jsFunc.Invoke())

			require.Equal(test, it.Result, it.Output(outputs))

		})
	}

}
