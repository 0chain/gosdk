// +build js,wasm

package jsbridge

import (
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
//  | Go                     | JavaScript             |
//  | ---------------------- | ---------------------- |
//  | js.Value               | [its value]            |
//  | js.Func                | function               |
//  | nil                    | null                   |
//  | bool                   | boolean                |
//  | integers and floats    | number                 |
//  | string                 | string                 |
//  | []interface{}          | new array              |
//  | map[string]interface{} | new object             |
//
// Panics if x is not one of the expected types.
func (b *InputBuilder) Build() (InputBinder, error) {

	b.binders = make([]func(jv js.Value) reflect.Value, b.numIn)

	if b.IsVariadic {
		b.numIn--
	}

	for i := 0; i < b.numIn; i++ {
		inputType := b.fn.In(i)

		v := reflect.New(inputType).Interface()

		switch v.(type) {
		case *string:
			b.binders[i] = jsValueToString

		case *int:
			b.binders[i] = jsValueToInt
		case *int32:
			b.binders[i] = jsValueToInt32
		case *int64:
			b.binders[i] = jsValueToInt64

		case *float32:
			b.binders[i] = jsValueToFloat32
		case *float64:
			b.binders[i] = jsValueToFloat64
		case *bool:
			b.binders[i] = jsValueToBool

		default:
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
