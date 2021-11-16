// +build js,wasm

package jsbridge

import (
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
				err := rv.Interface().(error)
				if err != nil {
					return js.ValueOf(NewJsError(err.Error()))
				}
				return js.Null()

			}
		case TypeString:
			b.binders[i] = func(rv reflect.Value) js.Value {
				s := rv.Interface().(string)
				return js.ValueOf(s)
			}

		default:
			b.binders[i] = func(rv reflect.Value) js.Value {
				return js.ValueOf(rv.Interface())
			}
		}

		// v := reflect.New(OutputType).Interface()

		// switch v.(type) {
		// case *string:
		// //	b.binders[i] = jsValueToString

		// case *int:
		// //	b.binders[i] = jsValueToInt
		// case *int32:
		// //	b.binders[i] = jsValueToInt32
		// case *int64:
		// //	b.binders[i] = jsValueToInt64

		// case *float32:
		// //	b.binders[i] = jsValueToFloat32
		// case *float64:
		// //	b.binders[i] = jsValueToFloat64
		// case *bool:
		// //	b.binders[i] = jsValueToBool

		// switch x := x.(type) {
		// case js.Value:
		// 	return x
		// case js.Wrapper:
		// 	return x.JSValue()
		// case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr,
		// 	unsafe.Pointer, float32, float64, string:
		// 	return js.ValueOf(x)
		// case complex64:
		// 	return js.ValueOf(map[string]interface{}{
		// 		"real": real(x),
		// 		"imag": imag(x),
		// 	})
		// case complex128:
		// 	return js.ValueOf(map[string]interface{}{
		// 		"real": real(x),
		// 		"imag": imag(x),
		// 	})
		// }

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

func NewJsError(message string) map[string]string {
	return map[string]string{
		"error": message,
	}
}
