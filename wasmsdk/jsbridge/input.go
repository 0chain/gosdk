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
	binders    []func(jv js.Value) (reflect.Value, error)
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

	b.binders = make([]func(jv js.Value) (reflect.Value, error), b.numIn)

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
		case *uint64:
			b.binders[i] = jsValueToUInt64
		case *float32:
			b.binders[i] = jsValueToFloat32
		case *float64:
			b.binders[i] = jsValueToFloat64
		case *bool:
			b.binders[i] = jsValueToBool
		case *[]string:
			b.binders[i] = jsValueToStringSlice
		case *[]byte:
			b.binders[i] = jsValueToBytes
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
		fmt.Println("args:", args)
		return nil, ErrMismatchedInputLength
	}

	values := make([]reflect.Value, b.numIn)
	for i := 0; i < b.numIn; i++ {
		val, err := b.binders[i](args[i])
		if err != nil {
			return nil, err
		}
		values[i] = val
	}

	return values, nil
}

func jsValueToString(jv js.Value) (val reflect.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("input: %s", r)
		}
	}()

	i := ""
	if jv.Truthy() {
		i = jv.String()
	}

	val = reflect.ValueOf(i)
	return
}

func jsValueToInt(jv js.Value) (val reflect.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("input: %s", r)
		}
	}()
	i := 0
	if jv.Truthy() {
		i = jv.Int()
	}

	val = reflect.ValueOf(i)
	return
}

func jsValueToInt32(jv js.Value) (val reflect.Value, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("input: %s", r)
		}
	}()
	i := 0
	if jv.Truthy() {
		i = jv.Int()
	}

	val = reflect.ValueOf(int32(i))
	return
}

func jsValueToInt64(jv js.Value) (val reflect.Value, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("input: %s", r)
		}
	}()
	i := 0
	if jv.Truthy() {
		i = jv.Int()
	}

	val = reflect.ValueOf(int64(i))
	return
}

func jsValueToUInt64(jv js.Value) (val reflect.Value, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("input: %s", r)
		}
	}()
	i := 0
	if jv.Truthy() {
		i = jv.Int()
	}

	val = reflect.ValueOf(uint64(i))
	return
}

func jsValueToBool(jv js.Value) (val reflect.Value, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("input: %s", r)
		}
	}()
	i := false
	if jv.Truthy() {
		i = jv.Bool()
	}

	val = reflect.ValueOf(i)
	return
}

func jsValueToFloat32(jv js.Value) (val reflect.Value, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("input: %s", r)
		}
	}()
	var i float64
	if jv.Truthy() {
		i = jv.Float()
	}

	val = reflect.ValueOf(float32(i))
	return
}

func jsValueToFloat64(jv js.Value) (val reflect.Value, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("input: %s", r)
		}
	}()
	var i float64
	if jv.Truthy() {
		i = jv.Float()
	}

	val = reflect.ValueOf(i)
	return
}

func jsValueToStringSlice(jv js.Value) (val reflect.Value, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("input: %s", r)
		}
	}()
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

	val = reflect.ValueOf(list)
	return
}

func jsValueToBytes(jv js.Value) (val reflect.Value, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("input: %s", r)
		}
	}()

	var buf []byte

	if jv.Truthy() {
		buf = make([]byte, jv.Length())
		js.CopyBytesToGo(buf, jv)
	}

	val = reflect.ValueOf(buf)
	return
}
