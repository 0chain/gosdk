//go:build js && wasm
// +build js,wasm

package jsbridge

import (
	"reflect"
	"syscall/js"
)

type SyncInvoker func(fn reflect.Value, in []reflect.Value) js.Value

func Sync(funcType reflect.Type) (SyncInvoker, error) {
	outputBinder, err := NewOutputBuilder(funcType).Build()
	if err != nil {
		return nil, err
	}

	switch funcType.NumOut() {
	case 0:
		return func(fn reflect.Value, in []reflect.Value) (result js.Value) {

			defer func() {
				if r := recover(); r != nil {
					result = NewJsError(r)
				}
			}()

			fn.Call(in)

			return

		}, nil
	case 1:
		outputType := funcType.Out(0)
		//func(...)error
		if outputType.String() == TypeError {
			return func(fn reflect.Value, in []reflect.Value) (result js.Value) {

				defer func() {
					if r := recover(); r != nil {
						result = NewJsError(r)
					}
				}()

				err := fn.Call(in)[0]

				// err != nil
				if !err.IsNil() {
					result = NewJsError(err.Interface())
				}

				return
			}, nil
		} else { //func(...) T
			return func(fn reflect.Value, in []reflect.Value) (result js.Value) {

				defer func() {
					if r := recover(); r != nil {
						result = NewJsError(r)
					}
				}()

				output := fn.Call(in)

				result = outputBinder(output)[0]

				return

			}, nil
		}
	case 2:

		errOutputType := funcType.Out(1)

		if errOutputType.String() != TypeError {
			return nil, ErrFuncNotSupported
		}
		//func(...) (T,error)
		return func(fn reflect.Value, in []reflect.Value) (result js.Value) {
			defer func() {
				if r := recover(); r != nil {
					result = NewJsError(r)
				}
			}()
			output := fn.Call(in)

			err := output[1]

			// err == nil
			if err.IsNil() {
				result = outputBinder(output)[0]
			} else {
				result = NewJsError(err.Interface())
			}

			return
		}, nil

	default:
		return nil, ErrFuncNotSupported
	}
}
