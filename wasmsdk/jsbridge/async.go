//go:build js && wasm
// +build js,wasm

package jsbridge

import (
	"reflect"
	"syscall/js"
)

type AsyncInvoker func(resolve, reject js.Value, fn reflect.Value, in []reflect.Value, err error)

func Async(funcType reflect.Type) (AsyncInvoker, error) {

	outputBinder, err := NewOutputBuilder(funcType).Build()
	if err != nil {
		return nil, err
	}

	switch funcType.NumOut() {
	case 0:
		return func(resolve, reject js.Value, fn reflect.Value, in []reflect.Value, err error) {
			if err != nil {
				jsErr := NewJsError(err.Error())
				resolve.Invoke(js.ValueOf(jsErr))
				return
			}

			fn.Call(in)
			resolve.Invoke()

		}, nil
	case 1:

		outputType := funcType.Out(0)
		//func(...)error
		if outputType.String() == TypeError {
			return func(resolve, reject js.Value, fn reflect.Value, in []reflect.Value, err error) {
				if err != nil {
					jsErr := NewJsError(err.Error())
					resolve.Invoke(js.ValueOf(jsErr))
					return
				}

				output := fn.Call(in)

				if output[0].IsNil() {
					resolve.Invoke()
				} else {
					args := outputBinder(output)
					resolve.Invoke(args[0])
				}
			}, nil
		} else { //func(...) T
			return func(resolve, reject js.Value, fn reflect.Value, in []reflect.Value, err error) {
				if err != nil {
					jsErr := NewJsError(err.Error())
					resolve.Invoke(js.ValueOf(jsErr))
					return
				}

				output := fn.Call(in)
				args := outputBinder(output)
				resolve.Invoke(args[0])
			}, nil
		}
	case 2:

		errOutputType := funcType.Out(1)

		if errOutputType.String() != TypeError {
			return nil, ErrFuncNotSupported
		}
		//func(...) (T,error)
		return func(resolve, reject js.Value, fn reflect.Value, in []reflect.Value, err error) {
			if err != nil {
				jsErr := NewJsError(err.Error())
				resolve.Invoke(js.ValueOf(jsErr))
				return
			}

			output := fn.Call(in)

			args := outputBinder(output)
			if output[1].IsNil() {
				resolve.Invoke(args[0])
			} else {
				resolve.Invoke(args[1])
			}

		}, nil

	default:
		return nil, ErrFuncNotSupported
	}

}

// This function try to execute wasm functions that are wrapped with "Promise"
// see: https://stackoverflow.com/questions/68426700/how-to-wait-a-js-async-function-from-golang-wasm/68427221#comment120939975_68427221
func Await(awaitable js.Value) ([]js.Value, []js.Value) {
	then := make(chan []js.Value)
	defer close(then)
	thenFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		then <- args
		return nil
	})
	defer thenFunc.Release()

	catch := make(chan []js.Value)
	defer close(catch)
	catchFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		catch <- args
		return nil
	})
	defer catchFunc.Release()

	awaitable.Call("then", thenFunc).Call("catch", catchFunc)

	select {
	case result := <-then:
		return result, []js.Value{js.Null()}
	case err := <-catch:
		return []js.Value{js.Null()}, err
	}
}
