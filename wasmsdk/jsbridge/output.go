//go:build js && wasm
// +build js,wasm

package jsbridge

import (
	"fmt"
	"reflect"
	"syscall/js"
)

// OutputBinder convert Outputs from js.Value to reflect.Value
type OutputBinder func([]reflect.Value) []js.Value

// OutputBuilder binder builder
type OutputBuilder struct {
	fn      reflect.Type
	numOut  int
	binders []func(rv reflect.Value) js.Value
}

// NewOutputBuilder create OutputBuilder
func NewOutputBuilder(fn reflect.Type) *OutputBuilder {
	return &OutputBuilder{
		fn:     fn,
		numOut: fn.NumOut(),
	}
}

// Build build OutputBinder
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
func (b *OutputBuilder) Build() (OutputBinder, error) {

	b.binders = make([]func(rv reflect.Value) js.Value, b.numOut)

	for i := 0; i < b.numOut; i++ {
		outputType := b.fn.Out(i)

		// TODO: Fast path for basic types that do not require reflection.
		switch outputType.String() {
		case TypeError:
			b.binders[i] = func(rv reflect.Value) js.Value {
				if rv.IsNil() {
					return js.Null()
				}

				err := rv.Interface().(error)
				if err != nil {
					jsErr := NewJsError(err.Error())
					return js.ValueOf(jsErr)
				}
				return js.Null()

			}
		case TypeString:
			b.binders[i] = func(rv reflect.Value) js.Value {
				if rv.IsZero() {
					return js.Null()
				}

				s := rv.Interface().(string)
				return js.ValueOf(s)
			}

		default:
			fmt.Println("Output: missing binder for ", outputType.String())
			b.binders[i] = func(rv reflect.Value) js.Value {
				return js.ValueOf(rv.Interface())
			}
		}
	}

	return b.Bind, nil
}

// Bind bind js Outputs to reflect values
func (b *OutputBuilder) Bind(args []reflect.Value) []js.Value {
	values := make([]js.Value, b.numOut)
	for i := 0; i < b.numOut; i++ {
		values[i] = b.binders[i](args[i])
	}
	return values
}

func NewJsError(message interface{}) js.Value {
	return js.ValueOf(map[string]interface{}{
		"error": fmt.Sprint(message),
	})
}
