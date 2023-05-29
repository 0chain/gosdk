//go:build js && wasm
// +build js,wasm

package jsbridge

import (
	"fmt"
	"reflect"
	"syscall/js"
)

// InputBinder convert inputs from js.Value to reflect.Value
type InputBinder func([]js.Value) ([]reflect.Value, error)

// InputBuilder binder builder
type InputBuilder struct {
	fn         reflect.Type
	numIn      int
	IsVariadic bool
	binders    []func(jv js.Value) reflect.Value
}

// NewInputBuilder create InputBuilder
func NewInputBuilder(fn reflect.Type) *InputBuilder {
	return &InputBuilder{
		fn:         fn,
		numIn:      fn.NumIn(),
		IsVariadic: fn.IsVariadic(),
	}
}

// Build build InputBinder
// js.ValueOf returns x as a JavaScript value:
//
//	| Go                     | JavaScript             |
//	| ---------------------- | ---------------------- |
//	| js.Value               | [its value]            |
//	| js.Func                | function               |
//	| nil                    | null                   |
//	| bool                   | boolean                |
//	| integers and floats    | number                 |
//	| string                 | string                 |
//	| []interface{}          | new array              |
//	| map[string]interface{} | new object             |
//
// Panics if x is not one of the expected types.
func (b *InputBuilder) Build() (InputBinder, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("[recover]InputBuilder.Build: ", r)
		}
	}()

	b.binders = make([]func(jv js.Value) reflect.Value, b.numIn)

	if b.IsVariadic {
		b.numIn--
	}

	for i := 0; i < b.numIn; i++ {
		inputType := b.fn.In(i)

		v := reflect.New(inputType).Interface()

		switch v.(type) {
		case *string:
			b.binders[i] = withRecover(i, jsValueToString)

		case *int:
			b.binders[i] = withRecover(i, jsValueToInt)
		case *int32:
			b.binders[i] = withRecover(i, jsValueToInt32)
		case *int64:
			b.binders[i] = withRecover(i, jsValueToInt64)
		case *uint64:
			b.binders[i] = withRecover(i, jsValueToUInt64)
		case *float32:
			b.binders[i] = withRecover(i, jsValueToFloat32)
		case *float64:
			b.binders[i] = withRecover(i, jsValueToFloat64)
		case *bool:
			b.binders[i] = withRecover(i, jsValueToBool)
		case *[]string:
			b.binders[i] = withRecover(i, jsValueToStringSlice)
		case *[]byte:
			b.binders[i] = withRecover(i, jsValueToBytes)
		default:
			fmt.Printf("TYPE: %#v\n", reflect.TypeOf(v))
			return nil, ErrBinderNotImplemented
		}

	}

	return b.Bind, nil
}

// Bind bind js inputs to reflect values
func (b *InputBuilder) Bind(args []js.Value) ([]reflect.Value, error) {
	if len(args) != b.numIn {
		return nil, ErrMismatchedInputLength
	}

	values := make([]reflect.Value, b.numIn)
	for i := 0; i < b.numIn; i++ {
		values[i] = b.binders[i](args[i])
	}

	return values, nil
}

func withRecover(inputIndex int, inputBinder func(jv js.Value) reflect.Value) func(jv js.Value) reflect.Value {

	return func(jv js.Value) reflect.Value {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("[recover]", inputIndex, ":", r)
			}
		}()

		return inputBinder(jv)
	}

}

func jsValueToString(jv js.Value) reflect.Value {
	i := ""
	if jv.Truthy() {
		i = jv.String()
	}

	return reflect.ValueOf(i)
}

func jsValueToInt(jv js.Value) reflect.Value {
	i := 0
	if jv.Truthy() {
		i = jv.Int()
	}

	return reflect.ValueOf(i)
}

func jsValueToInt32(jv js.Value) reflect.Value {
	i := 0
	if jv.Truthy() {
		i = jv.Int()
	}

	return reflect.ValueOf(int32(i))
}

func jsValueToInt64(jv js.Value) reflect.Value {
	i := 0
	if jv.Truthy() {
		i = jv.Int()
	}

	return reflect.ValueOf(int64(i))
}

func jsValueToUInt64(jv js.Value) reflect.Value {
	i := 0
	if jv.Truthy() {
		i = jv.Int()
	}

	return reflect.ValueOf(uint64(i))
}

func jsValueToBool(jv js.Value) reflect.Value {
	i := false
	if jv.Truthy() {
		i = jv.Bool()
	}

	return reflect.ValueOf(i)
}

func jsValueToFloat32(jv js.Value) reflect.Value {
	var i float64
	if jv.Truthy() {
		i = jv.Float()
	}

	return reflect.ValueOf(float32(i))
}

func jsValueToFloat64(jv js.Value) reflect.Value {
	var i float64
	if jv.Truthy() {
		i = jv.Float()
	}

	return reflect.ValueOf(i)
}

func jsValueToStringSlice(jv js.Value) reflect.Value {
	var list []string

	if jv.Truthy() {
		if js.Global().Get("Array").Call("isArray", jv).Bool() {
			list = make([]string, jv.Length())
			for i := 0; i < len(list); i++ {
				it := jv.Index(i)
				if it.Truthy() {
					list[i] = it.String()
				}
			}
		}
	}

	return reflect.ValueOf(list)
}

func jsValueToBytes(jv js.Value) reflect.Value {

	var buf []byte

	if jv.Truthy() {
		buf = make([]byte, jv.Length())
		js.CopyBytesToGo(buf, jv)
	}

	return reflect.ValueOf(buf)
}
